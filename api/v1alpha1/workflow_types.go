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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkStatus string

const (
	WorkFlowStatusWaiting WorkStatus = "Waiting"
	WorkFlowStatusRunning WorkStatus = "Running"
	WorkFlowStatusSuccess WorkStatus = "Success"
	WorkFlowStatusFailed  WorkStatus = "Failed"
	WorkFlowStatusCancel  WorkStatus = "Cancel"
	WorkFlowStatusPause   WorkStatus = "Pause"
)

type Step struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image"`
	Script      string `json:"script"`
	Args        string `json:"args,omitempty"`
}

type Task struct {
	Name         string           `json:"name"`
	DisplayName  string           `json:"displayName,omitempty"`
	Description  string           `json:"description,omitempty"`
	Dependencies []string         `json:"dependencies,omitempty"`
	Results      []TaskOutput     `json:"results,omitempty"`
	Timeout      *metav1.Duration `json:"timeout,omitempty"`
	Steps        []Step           `json:"steps"`
}

type TaskOutput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type TaskStatus struct {
	Name           string       `json:"name"`
	PodName        string       `json:"podName"`
	Message        string       `json:"message,omitempty"`
	Status         v1.PodPhase  `json:"status"`
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	Results        []*Output    `json:"results,omitempty"`
}

type Output struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Input struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WorkflowSpec defines the desired state of Workflow
type WorkflowSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Inputs []Input `json:"inputs,omitempty"`
	Tasks  []Task  `json:"tasks"`
}

// WorkflowStatus defines the observed state of Workflow
type WorkflowStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status         WorkStatus            `json:"status"`
	Message        string                `json:"message,omitempty"`
	StartTime      *metav1.Time          `json:"startTime,omitempty"`
	CompletionTime *metav1.Time          `json:"completionTime,omitempty"`
	TaskStatus     map[string]TaskStatus `json:"taskStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Workflow is the Schema for the workflows API
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkflowList contains a list of Workflow
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

func (w *Workflow) ValidateUniqueTaskNames() bool {
	taskNames := make(map[string]bool)
	for _, task := range w.Spec.Tasks {
		if taskNames[task.Name] {
			return true
		}
		taskNames[task.Name] = true
	}
	return false
}

func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}
