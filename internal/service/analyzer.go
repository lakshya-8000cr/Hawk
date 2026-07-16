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
	pods        repository.PodRepository
}

func NewAnalyzer(
	deployments repository.DeploymentRepository,
	replicaSets repository.ReplicaSetRepository,
	pods repository.PodRepository,
) *Analyzer {
	return &Analyzer{
		deployments: deployments,
		replicaSets: replicaSets,
		pods:        pods,
	}
}

func (a *Analyzer) AnalyzeDeployment(
	ctx context.Context,
	namespace string,
	name string,
) (*domain.DeploymentAnalysis, error) {
	// 1. Fetch target deployment
	deployment, err := a.deployments.Get(
		ctx,
		namespace,
		name,
	)
	if err != nil {
		return nil, fmt.Errorf("analyze deployment: %w", err)
	}

	// 2. Fetch ReplicaSets owned by deployment
	replicaSets, err := a.replicaSets.FindOwnedBy(
		ctx,
		namespace,
		deployment.UID,
	)
	if err != nil {
		return nil, fmt.Errorf("collect owned replicasets: %w", err)
	}

	// 3. Collect ReplicaSet UIDs
	replicaSetUIDs := make([]string, 0, len(replicaSets))

	for _, replicaSet := range replicaSets {
		replicaSetUIDs = append(replicaSetUIDs, replicaSet.UID)
	}

	// 4. Fetch Pods owned by those ReplicaSets
	pods, err := a.pods.FindOwnedBy(
		ctx,
		namespace,
		replicaSetUIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("collect owned pods: %w", err)
	}

	// 5. Create dependency graph
	resourceGraph := graph.New()

	deploymentNode := graph.NewNode(
		"Deployment",
		deployment.Namespace,
		deployment.Name,
	)

	resourceGraph.AddNode(deploymentNode)

	// UID -> graph node ID map
	replicaSetNodeIDs := make(map[string]string)

	// 6. Deployment -> ReplicaSet edges
	for _, replicaSet := range replicaSets {
		replicaSetNode := graph.NewNode(
			"ReplicaSet",
			replicaSet.Namespace,
			replicaSet.Name,
		)

		resourceGraph.AddNode(replicaSetNode)

		replicaSetNodeIDs[replicaSet.UID] = replicaSetNode.ID

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

	// 7. ReplicaSet -> Pod edges
	for _, pod := range pods {
		podNode := graph.NewNode(
			"Pod",
			pod.Namespace,
			pod.Name,
		)

		resourceGraph.AddNode(podNode)

		replicaSetNodeID, exists := replicaSetNodeIDs[pod.OwnerUID]
		if !exists {
			continue
		}

		if err := resourceGraph.AddEdge(graph.Edge{
			From:         replicaSetNodeID,
			To:           podNode.ID,
			Relationship: graph.Owns,
		}); err != nil {
			return nil, fmt.Errorf(
				"connect replicaset to pod: %w",
				err,
			)
		}
	}

	return &domain.DeploymentAnalysis{
		Deployment:  deployment,
		ReplicaSets: replicaSets,
		Pods:        pods,
		Graph:       resourceGraph,
	}, nil
}
