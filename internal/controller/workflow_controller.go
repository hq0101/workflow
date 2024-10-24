/*
Copyright 2024 hq0101.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/storage/names"
	"time"

	skyv1alpha1 "github.com/hq0101/workflow/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// WorkflowReconciler reconciles a Workflow object
type WorkflowReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=sky.my.domain,resources=workflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sky.my.domain,resources=workflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=sky.my.domain,resources=workflows/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Workflow object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *WorkflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	workflow := &skyv1alpha1.Workflow{}

	if err := r.Get(ctx, req.NamespacedName, workflow); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Workflow resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Failed to get Workflow")
		return ctrl.Result{}, err
	}

	r.setStartTime(workflow)

	if !workflow.DeletionTimestamp.IsZero() {
		logger.Info("Workflow resource is being deleted")
		return ctrl.Result{}, r.clearFinalizers(ctx, workflow)
	}

	if workflow.ValidateUniqueTaskNames() {
		logger.Info("WorkFlow has duplicate task names")
		workflow.Status.Message = "WorkFlow has duplicate task names"
		workflow.Status.Status = skyv1alpha1.WorkFlowStatusFailed
		if _err := r.Client.Update(ctx, workflow); _err != nil {
			logger.Error(_err, "Failed to update WorkFlow status")
			return ctrl.Result{}, _err
		}
		return ctrl.Result{}, nil
	}

	taskStatus := make(map[string]skyv1alpha1.TaskStatus)

	var compeletedCount int
	var isFailed bool
	for _, task := range workflow.Status.TaskStatus {
		switch task.Status {
		case corev1.PodFailed:
			compeletedCount += 1
			taskStatus[task.Name] = task
			isFailed = true
			continue
		case corev1.PodSucceeded:
			compeletedCount += 1
			taskStatus[task.Name] = task
			continue
		}
		_pod, _err := r.getPod(ctx, task.PodName, workflow.GetNamespace())
		if _err != nil {
			logger.Error(_err, "Failed to get Pod")
			return ctrl.Result{}, _err
		}

		status := skyv1alpha1.TaskStatus{
			Name:    task.Name,
			PodName: _pod.Name,
			Status:  _pod.Status.Phase,
		}
		if _pod.Status.Phase == corev1.PodFailed || _pod.Status.Phase == corev1.PodSucceeded {
			status.Message = _pod.Status.Message
			state := _pod.Status.ContainerStatuses[len(_pod.Spec.Containers)-1].State
			if state.Terminated != nil {
				status.CompletionTime = &state.Terminated.FinishedAt
			}
		}
		for _, containerStatus := range _pod.Status.ContainerStatuses {
			if containerStatus.State.Terminated != nil {
				if containerStatus.State.Terminated.Message != "" {
					outputs := []*skyv1alpha1.Output{}
					if _err = json.Unmarshal([]byte(containerStatus.State.Terminated.Message), &outputs); _err == nil {
						status.Outputs = outputs
					} else {
						logger.Error(_err, "Failed to unmarshal results")
					}
				}
			}
		}
		taskStatus[task.Name] = status
	}

	workflow.Status.TaskStatus = taskStatus

	if compeletedCount == len(workflow.Status.TaskStatus) && len(workflow.Status.TaskStatus) != 0 {
		workflow.Status.Status = skyv1alpha1.WorkFlowStatusSuccess
		if isFailed {
			workflow.Status.Status = skyv1alpha1.WorkFlowStatusFailed
		}
	}

	if _err := r.Status().Update(ctx, workflow); _err != nil {
		logger.Error(_err, "Failed to update WorkFlow")
		return ctrl.Result{}, _err
	}

	d, err := BuildDAG(workflow.Spec.Tasks)
	if err != nil {
		logger.Error(err, "Failed to build DAG")
		return ctrl.Result{}, err
	}

	if !d.Validate() {
		logger.Info("WorkFlow has invalid dependencies")
		workflow.Status.Message = "WorkFlow has invalid dependencies"
		workflow.Status.Status = skyv1alpha1.WorkFlowStatusFailed
		if _err := r.Client.Update(ctx, workflow); _err != nil {
			logger.Error(_err, "Failed to update WorkFlow status")
			return ctrl.Result{}, _err
		}
		return ctrl.Result{}, nil
	}

	nextNodes := FindSchedulableNodes(d, FindCompletedTasks(workflow), taskStatus)
	if len(nextNodes) == 0 {
		if workflow.Status.Status == skyv1alpha1.WorkFlowStatusSuccess || workflow.Status.Status == skyv1alpha1.WorkFlowStatusFailed {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{
			RequeueAfter: 5 * time.Second,
		}, nil
	}

	nextTasks := FindSchedulableTasks(nextNodes, workflow.Spec.Tasks)
	if len(nextTasks) == 0 {
		if workflow.Status.Status == skyv1alpha1.WorkFlowStatusSuccess || workflow.Status.Status == skyv1alpha1.WorkFlowStatusFailed {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{
			RequeueAfter: 5 * time.Second,
		}, nil
	}

	for _, task := range nextTasks {
		pod, _err := r.createPod(ctx, task, workflow)
		if _err != nil {
			logger.Error(_err, "Failed to create Task")
			workflow.Status.Message = _err.Error()
			workflow.Status.Status = skyv1alpha1.WorkFlowStatusFailed
			if _err = r.Status().Update(ctx, workflow); _err != nil {
				logger.Error(_err, "Failed to update WorkFlow")
			}
			return ctrl.Result{}, _err
		}
		taskStatus[task.Name] = skyv1alpha1.TaskStatus{
			Name:    task.Name,
			PodName: pod.Name,
			Status:  pod.Status.Phase,
		}
		workflow.Finalizers = append(workflow.Finalizers, pod.Name)
	}

	if workflow.Status.Status == "" {
		workflow.Status.Status = skyv1alpha1.WorkFlowStatusRunning
	}
	workflow.Status.TaskStatus = taskStatus
	if _err := r.Status().Update(ctx, workflow); _err != nil {
		logger.Error(_err, "Failed to update WorkFlow", "workflow", workflow.Name)
		return ctrl.Result{}, _err
	}

	return ctrl.Result{
		RequeueAfter: 5,
	}, nil
}

func (r *WorkflowReconciler) getPod(ctx context.Context, podName, namespace string) (*corev1.Pod, error) {
	pod := &corev1.Pod{}
	if _err := r.Client.Get(ctx, client.ObjectKey{Name: podName, Namespace: namespace}, pod); _err != nil {
		return nil, _err
	}
	return pod, nil
}

func (r *WorkflowReconciler) findPod(ctx context.Context, workflow *skyv1alpha1.Workflow) (*corev1.Pod, error) {
	pods := &corev1.PodList{}

	if err := r.Client.List(ctx, pods, client.HasLabels{taskLabelKey}, client.InNamespace(workflow.GetNamespace())); err != nil {
		return nil, err
	}
	for _, _pod := range pods.Items {
		if metav1.IsControlledBy(&_pod, workflow) {
			return &_pod, nil
		}
	}

	return nil, nil
}

func (r *WorkflowReconciler) createPod(ctx context.Context, task skyv1alpha1.Task, workFlow *skyv1alpha1.Workflow) (*corev1.Pod, error) {
	podName := names.SimpleNameGenerator.GenerateName(fmt.Sprintf("%s-%s-", workFlow.Name, task.Name))

	coreV1Pod, err := generatePod(ctx, task, task.Steps, task.Name, podName, task.Outputs, workFlow)
	if err != nil {
		return coreV1Pod, err
	}

	if _err := r.Client.Create(ctx, coreV1Pod); _err != nil {
		return coreV1Pod, err
	}

	return coreV1Pod, nil
}

func (r *WorkflowReconciler) clearFinalizers(ctx context.Context, workflow *skyv1alpha1.Workflow) error {
	workflow.Finalizers = []string{}
	return r.Client.Update(ctx, workflow)
}

func (r *WorkflowReconciler) setStartTime(workflow *skyv1alpha1.Workflow) {
	if workflow.Status.StartTime == nil {
		workflow.Status.StartTime = &workflow.CreationTimestamp
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&skyv1alpha1.Workflow{}).
		Complete(r)
}
