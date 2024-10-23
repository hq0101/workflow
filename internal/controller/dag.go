package controller

import (
	"fmt"
	skyv1alpha1 "github.com/hq0101/workflow/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

type Node struct {
	Name string
	Prev []*Node
	Next []*Node
}

type Dag struct {
	Nodes map[string]*Node
}

func BuildDAG(tasks []skyv1alpha1.Task) (*Dag, error) {
	dag := &Dag{Nodes: make(map[string]*Node)}
	for _, task := range tasks {

		node := &Node{Name: task.Name}
		dag.Nodes[task.Name] = node

		for _, dependency := range task.Dependencies {
			dependencyNode, ok := dag.Nodes[dependency]
			if !ok {
				return nil, fmt.Errorf("dependency not found: %s", dependency)
			}
			node.Prev = append(node.Prev, dependencyNode)
			dependencyNode.Next = append(dependencyNode.Next, node)
		}
	}

	return dag, nil
}

func (dag *Dag) hasCycle() bool {
	for _, node := range dag.Nodes {
		if dag.detectCycleDFS(node, make(map[*Node]bool)) {
			return true
		}
	}
	return false
}

func (dag *Dag) hasSingleRoot() bool {
	rootCount := 0
	for _, node := range dag.Nodes {
		if len(node.Prev) == 0 {
			rootCount++
		}
	}
	return rootCount == 1
}

func (dag *Dag) Validate() bool {
	return !dag.hasCycle() && dag.hasSingleRoot()
}

func (dag *Dag) detectCycleDFS(node *Node, visited map[*Node]bool) bool {
	if visited[node] {
		return visited[node] == true
	}
	visited[node] = true
	for _, prev := range node.Prev {
		if dag.detectCycleDFS(prev, visited) {
			return true
		}
	}
	visited[node] = false
	return false
}

func (dag *Dag) GetRootNode() *Node {
	for _, node := range dag.Nodes {
		if len(node.Prev) == 0 {
			return node
		}
	}
	return nil
}

func FindSchedulableTasks(nodes []*Node, tasks []skyv1alpha1.Task) []skyv1alpha1.Task {
	var nextTasks []skyv1alpha1.Task
	for _, node := range nodes {
		for _, task := range tasks {
			if task.Name == node.Name {
				nextTasks = append(nextTasks, task)
			}
		}
	}
	return nextTasks
}

func FindCompletedTasks(flow *skyv1alpha1.Workflow) []string {
	var completedTasks []string
	for _, status := range flow.Status.TaskStatus {
		if status.Status == v1.PodSucceeded || status.Status == v1.PodFailed {
			completedTasks = append(completedTasks, status.Name)
		}
	}
	return completedTasks
}

func FindSchedulableNodes(dag *Dag, completedTasks []string, taskStatus map[string]skyv1alpha1.TaskStatus) []*Node {
	if dag == nil || len(dag.Nodes) == 0 {
		return []*Node{}
	}

	var nextNodes []*Node
	for _, node := range dag.Nodes {
		if _, ok := taskStatus[node.Name]; ok {
			continue
		}

		if node.Prev == nil {
			nextNodes = append(nextNodes, node)
			continue
		}

		allPrevCompleted := false
		for _, prev := range node.Prev {
			if prevTask, ok := taskStatus[prev.Name]; ok {
				if prevTask.Status == v1.PodSucceeded {
					allPrevCompleted = true
					break
				}
			}
		}
		if allPrevCompleted {
			nextNodes = append(nextNodes, node)
		}
	}
	return nextNodes
}
