package graph

import "fmt"

type Graph struct {
	nodes    map[string]Node
	outgoing map[string][]Edge
	incoming map[string][]Edge
}

func New() *Graph {
	return &Graph{
		nodes:    make(map[string]Node),
		outgoing: make(map[string][]Edge),
		incoming: make(map[string][]Edge),
	}
}

func (g *Graph) AddNode(node Node) {
	g.nodes[node.ID] = node
}

func (g *Graph) AddEdge(edge Edge) error {
	if _, exists := g.nodes[edge.From]; !exists {
		return fmt.Errorf("source node %q does not exist", edge.From)
	}

	if _, exists := g.nodes[edge.To]; !exists {
		return fmt.Errorf("target node %q does not exist", edge.To)
	}

	g.outgoing[edge.From] = append(g.outgoing[edge.From], edge)
	g.incoming[edge.To] = append(g.incoming[edge.To], edge)

	return nil
}

func (g *Graph) GetNode(id string) (Node, bool) {
	node, exists := g.nodes[id]
	return node, exists
}

func (g *Graph) GetOutgoing(id string) []Edge {
	return g.outgoing[id]
}

func (g *Graph) GetIncoming(id string) []Edge {
	return g.incoming[id]
}

func (g *Graph) Nodes() []Node {
	nodes := make([]Node, 0, len(g.nodes))

	for _, node := range g.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}