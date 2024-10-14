package controller

type Node struct {
	Name string
	Prev []*Node
	Next []*Node
}

type Dag struct {
	Nodes map[string]*Node
}
