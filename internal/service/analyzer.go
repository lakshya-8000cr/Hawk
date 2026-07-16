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
	services    repository.ServiceRepository
}

func NewAnalyzer(
	deployments repository.DeploymentRepository,
	replicaSets repository.ReplicaSetRepository,
	pods repository.PodRepository,
	services repository.ServiceRepository,
) *Analyzer {
	return &Analyzer{
		deployments: deployments,
		replicaSets: replicaSets,
		pods:        pods,
		services:    services,
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
		return nil, fmt.Errorf(
			"collect owned replicasets: %w",
			err,
		)
	}

	replicaSetUIDs := make([]string, 0, len(replicaSets))

	for _, replicaSet := range replicaSets {
		replicaSetUIDs = append(
			replicaSetUIDs,
			replicaSet.UID,
		)
	}

	pods, err := a.pods.FindOwnedBy(
		ctx,
		namespace,
		replicaSetUIDs,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"collect owned pods: %w",
			err,
		)
	}

	allServices, err := a.services.List(
		ctx,
		namespace,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"collect services: %w",
			err,
		)
	}

	resourceGraph := graph.New()

	deploymentNode := graph.NewNode(
		"Deployment",
		deployment.Namespace,
		deployment.Name,
	)

	resourceGraph.AddNode(deploymentNode)

	replicaSetNodeIDs := make(map[string]string)
	podNodeIDs := make(map[string]string)

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

	for _, pod := range pods {
		podNode := graph.NewNode(
			"Pod",
			pod.Namespace,
			pod.Name,
		)

		resourceGraph.AddNode(podNode)
		podNodeIDs[pod.UID] = podNode.ID

		replicaSetNodeID, exists :=
			replicaSetNodeIDs[pod.OwnerUID]

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

	matchedServices := make([]domain.Service, 0)

	for _, svc := range allServices {
		serviceMatched := false

		for _, pod := range pods {
			if !selectorMatches(
				svc.Selector,
				pod.Labels,
			) {
				continue
			}

			if !serviceMatched {
				matchedServices = append(
					matchedServices,
					svc,
				)

				serviceMatched = true
			}

			serviceNode := graph.NewNode(
				"Service",
				svc.Namespace,
				svc.Name,
			)

			resourceGraph.AddNode(serviceNode)

			podNodeID, exists := podNodeIDs[pod.UID]
			if !exists {
				continue
			}

			if err := resourceGraph.AddEdge(graph.Edge{
				From:         serviceNode.ID,
				To:           podNodeID,
				Relationship: graph.Selects,
			}); err != nil {
				return nil, fmt.Errorf(
					"connect service to pod: %w",
					err,
				)
			}
		}
	}

	return &domain.DeploymentAnalysis{
		Deployment:  deployment,
		ReplicaSets: replicaSets,
		Pods:        pods,
		Services:    matchedServices,
		Graph:       resourceGraph,
	}, nil
}
