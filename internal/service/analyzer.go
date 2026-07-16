package service

import (
	"context"
	"fmt"

	"hawk/internal/domain"
	"hawk/internal/graph"
	"hawk/internal/repositry"
)

type Analyzer struct {
	deployments repository.DeploymentRepository
	replicaSets repository.ReplicaSetRepository
}

func NewAnalyzer(
	deployments repository.DeploymentRepository,
	replicaSets repository.ReplicaSetRepository,
) *Analyzer {
	return &Analyzer{
		deployments: deployments,
		replicaSets: replicaSets,
	}
}

func (a *Analyzer) AnalyzeDeployment(
	ctx context.Context,
	namespace string,
	name string,
) (*domain.DeploymentAnalysis, error) {
	deployment, err := a.deployments.Get(
		ctx,
		namespace,
		name,
	)
	if err != nil {
		return nil, fmt.Errorf("analyze deployment: %w", err)
	}

	replicaSets, err := a.replicaSets.FindOwnedBy(
		ctx,
		namespace,
		deployment.UID,
	)
	if err != nil {
		return nil, fmt.Errorf("collect owned replicasets: %w", err)
	}

	resourceGraph := graph.New()

	deploymentNode := graph.NewNode(
		"Deployment",
		deployment.Namespace,
		deployment.Name,
	)
	resourceGraph.AddNode(deploymentNode)

	for _, replicaSet := range replicaSets {
		replicaSetNode := graph.NewNode(
			"ReplicaSet",
			replicaSet.Namespace,
			replicaSet.Name,
		)

		resourceGraph.AddNode(replicaSetNode)

		if err := resourceGraph.AddEdge(graph.Edge{
			From:         deploymentNode.ID,
			To:           replicaSetNode.ID,
			Relationship: graph.Owns,
		}); err != nil {
			return nil, fmt.Errorf(
				"connect deployment to replicaset: %w",
				err,
			)
		}
	}

	return &domain.DeploymentAnalysis{
		Deployment:  deployment,
		ReplicaSets: replicaSets,
		Graph:       resourceGraph,
	}, nil
}