package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	skyv1alpha1 "github.com/hq0101/workflow/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

const (
	entrypointVolumeName   = "entrypoint-bin"
	scriptsVolumeName      = "internal-scripts"
	outputsVolumeName      = "internal-outputs"
	downwardVolumeName     = "internal-downward-api"
	runVolumeName          = "internal-run"
	scriptDir              = "/tmp/sky/scripts"
	outputDir              = "/tmp/sky/outputs"
	runDir                 = "/tmp/sky/run"
	downwardDir            = "/tmp/sky/downward"
	terminationMessagePath = "/tmp/termination-log"
	taskLabelKey           = "task_name"
)

func generatePod(ctx context.Context, task skyv1alpha1.Task, steps []skyv1alpha1.Step, taskName, podName string, taskOutput []skyv1alpha1.TaskOutput, workFlow *skyv1alpha1.Workflow) (*v1.Pod, error) {
	pod := &v1.Pod{}
	pod.Namespace = workFlow.Namespace
	pod.Name = podName
	pod.Labels = map[string]string{
		taskLabelKey: taskName,
	}
	pod.Annotations = map[string]string{
		"0": "0",
	}
	pod.Spec.RestartPolicy = v1.RestartPolicyNever

	copySteps := steps

	replacements := []string{}
	for _, input := range workFlow.Spec.Inputs {
		replacements = append(replacements, fmt.Sprintf("{{inputs.%s}}", input.Name), input.Value)
	}

	for _, taskStatus := range workFlow.Status.TaskStatus {
		for _, output := range taskStatus.Results {
			replacements = append(replacements, fmt.Sprintf("{{tasks.%s.outputs.%s}}", taskStatus.Name, output.Name), output.Value)
		}
	}

	replacer := strings.NewReplacer(replacements...)
	for i, step := range copySteps {
		copySteps[i].Args = replacer.Replace(step.Args)
		copySteps[i].Script = replacer.Replace(step.Script)
	}

	pod.Spec.InitContainers = initContainers(copySteps)

	outputs := ""
	for _, output := range taskOutput {
		outputs = fmt.Sprintf("%s %s", outputs, output.Name)
	}

	containers, err := stepContainers(steps, strings.TrimSpace(outputs))
	if err != nil {
		return nil, err
	}

	activeDeadlineSeconds := int64(task.GetTimeout().Seconds())
	pod.Spec.ActiveDeadlineSeconds = &activeDeadlineSeconds

	pod.Spec.Containers = containers
	pod.Spec.Volumes = []v1.Volume{
		{
			Name: entrypointVolumeName,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: scriptsVolumeName,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: outputsVolumeName,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: downwardVolumeName,
			VolumeSource: v1.VolumeSource{
				DownwardAPI: &v1.DownwardAPIVolumeSource{
					Items: []v1.DownwardAPIVolumeFile{
						{
							Path: "0",
							FieldRef: &v1.ObjectFieldSelector{
								FieldPath: fmt.Sprintf("metadata.annotations['%s']", "0"),
							},
						},
					},
				},
			},
		},
		{
			Name: runVolumeName,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
	}

	pod.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(workFlow, schema.GroupVersionKind{
			Kind:    skyv1alpha1.KindName,
			Group:   skyv1alpha1.GroupVersion.Group,
			Version: skyv1alpha1.GroupVersion.Version,
		}),
	}

	return pod, nil
}

func initContainers(steps []skyv1alpha1.Step) []v1.Container {
	scriptTemplate := "cp /app/entrypoint /app/bin;"
	for index, step := range steps {
		scriptTemplate = strings.TrimSpace(scriptTemplate)
		scriptTemplate += fmt.Sprintf(`
scriptfile="%s/%s";
touch ${scriptfile} && chmod +x ${scriptfile};
echo "%s" > ${scriptfile};
/app/bin/entrypoint --encode_script ${scriptfile};
`, scriptDir, fmt.Sprintf("%s-%d", step.Name, index), encodeScript(step.Script))
	}

	return []v1.Container{
		{
			Name:            "init-step",
			Image:           "registry.cn-shanghai.aliyuncs.com/sky/entrypoint:v0.0.1",
			ImagePullPolicy: v1.PullIfNotPresent,
			Command:         []string{"sh"},
			Args:            []string{"-c", scriptTemplate},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      entrypointVolumeName,
					MountPath: "/app/bin",
				},
				{
					Name:      scriptsVolumeName,
					MountPath: scriptDir,
				},
				{
					Name:      outputsVolumeName,
					MountPath: outputDir,
				},
				{
					Name:      downwardVolumeName,
					MountPath: downwardDir,
				},
				{
					Name:      runVolumeName,
					MountPath: runDir,
				},
			},
		},
	}
}

func stepContainers(steps []skyv1alpha1.Step, results string) ([]v1.Container, error) {
	var containers []v1.Container
	for index, step := range steps {
		waitFile := fmt.Sprintf("%s/%d", downwardDir, index)
		waitContent := fmt.Sprintf("%d", index)
		if index != 0 {
			waitFile = fmt.Sprintf("%s/%d", runDir, index-1)
			waitContent = fmt.Sprintf("%d", index-1)
		}
		postFile := fmt.Sprintf("%s/%d", runDir, index)

		args := []string{
			"--wait_file", waitFile,
			"--wait_content", waitContent,
			"--post_file", postFile,
			"--post_content", fmt.Sprintf("%d", index),
			"--command", fmt.Sprintf("%s/%s-%d", scriptDir, step.Name, index),
			"--encode", fmt.Sprintf("%t", true),
			"--results", results,
			"--termination_message_path", terminationMessagePath,
			"--params", step.Args,
		}
		containers = append(containers, v1.Container{
			Name:                     step.Name,
			Image:                    step.Image,
			ImagePullPolicy:          v1.PullIfNotPresent,
			Command:                  []string{"/app/bin/entrypoint"},
			Args:                     args,
			TerminationMessagePath:   terminationMessagePath,
			TerminationMessagePolicy: v1.TerminationMessageReadFile,
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      entrypointVolumeName,
					MountPath: "/app/bin",
				},
				{
					Name:      scriptsVolumeName,
					MountPath: scriptDir,
				},
				{
					Name:      outputsVolumeName,
					MountPath: outputDir,
				},
				{
					Name:      downwardVolumeName,
					MountPath: downwardDir,
				},
				{
					Name:      runVolumeName,
					MountPath: runDir,
				},
			},
		})
	}

	return containers, nil
}

func encodeScript(script string) string {
	return base64.StdEncoding.EncodeToString([]byte(script))
}
