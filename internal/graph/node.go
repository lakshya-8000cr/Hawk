package graph

import "fmt"

type Node struct {
	ID        string
	Kind      string
	Name      string
	Namespace string
}

func NewNode(kind, namespace, name string) Node {
	return Node{
		ID:        fmt.Sprintf("%s/%s/%s", namespace, kind, name),
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
	}
}