package graph

import "testing"

func TestGraphRelationships(t *testing.T) {
	g := New()

	deployment := NewNode(
		"Deployment",
		"default",
		"forge-frontend",
	)

	replicaSet := NewNode(
		"ReplicaSet",
		"default",
		"forge-frontend-abc123",
	)

	g.AddNode(deployment)
	g.AddNode(replicaSet)

	err := g.AddEdge(Edge{
		From:         deployment.ID,
		To:           replicaSet.ID,
		Relationship: Owns,
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(g.GetOutgoing(deployment.ID)) != 1 {
		t.Fatal("expected one outgoing relationship")
	}

	if len(g.GetIncoming(replicaSet.ID)) != 1 {
		t.Fatal("expected one incoming relationship")
	}
}