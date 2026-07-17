package service

import (
	"context"
	"fmt"

	"hawk/internal/domain"
	"hawk/internal/graph"
	"hawk/internal/repositry" // Fixed typo in import path from "repositry"
)

type Analyzer struct {
	deployments repository.DeploymentRepository
	replicaSets repository.ReplicaSetRepository
	pods        repository.PodRepository
	services    repository.ServiceRepository
	ingresses   repository.IngressRepository
}

func NewAnalyzer(
	deployments repository.DeploymentRepository,
	replicaSets repository.ReplicaSetRepository,
	pods repository.PodRepository,
	services repository.ServiceRepository,
	ingresses repository.IngressRepository,
) *Analyzer {
	return &Analyzer{
		deployments: deployments,
		replicaSets: replicaSets,
		pods:        pods,
		services:    services,
		ingresses:   ingresses,
	}
}

func (a *Analyzer) AnalyzeDeployment(
	ctx context.Context,
	namespace string,
	name string,
) (*domain.DeploymentAnalysis, error) {
	// 1. Fetch Deployment
	deployment, err := a.deployments.Get(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("analyze deployment: %w", err)
	}

	// 2. Fetch ReplicaSets owned by Deployment
	replicaSets, err := a.replicaSets.FindOwnedBy(ctx, namespace, deployment.UID)
	if err != nil {
		return nil, fmt.Errorf("collect owned replicasets: %w", err)
	}

	replicaSetUIDs := make([]string, 0, len(replicaSets))
	for _, replicaSet := range replicaSets {
		replicaSetUIDs = append(replicaSetUIDs, replicaSet.UID)
	}

	// 3. Fetch Pods owned by ReplicaSets
	pods, err := a.pods.FindOwnedBy(ctx, namespace, replicaSetUIDs)
	if err != nil {
		return nil, fmt.Errorf("collect owned pods: %w", err)
	}

	// 4. Fetch Services
	allServices, err := a.services.List(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("collect services: %w", err)
	}

	// 5. Fetch Ingresses (Sequential execution)
	allIngresses, err := a.ingresses.List(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("collect ingresses: %w", err)
	}

	// 6. Initialize Resource Graph
	resourceGraph := graph.New()

	deploymentNode := graph.NewNode("Deployment", deployment.Namespace, deployment.Name)
	resourceGraph.AddNode(deploymentNode)

	replicaSetNodeIDs := make(map[string]string)
	podNodeIDs := make(map[string]string)

	// 7. Add ReplicaSet nodes and edges
	for _, replicaSet := range replicaSets {
		replicaSetNode := graph.NewNode("ReplicaSet", replicaSet.Namespace, replicaSet.Name)
		resourceGraph.AddNode(replicaSetNode)
		replicaSetNodeIDs[replicaSet.UID] = replicaSetNode.ID

		err := resourceGraph.AddEdge(graph.Edge{
			From:         deploymentNode.ID,
			To:           replicaSetNode.ID,
			Relationship: graph.Owns,
		})
		if err != nil {
			return nil, fmt.Errorf("connect deployment to replicaset: %w", err)
		}
	}

	// 8. Add Pod nodes and edges
	for _, pod := range pods {
		podNode := graph.NewNode("Pod", pod.Namespace, pod.Name)
		resourceGraph.AddNode(podNode)
		podNodeIDs[pod.UID] = podNode.ID

		replicaSetNodeID, exists := replicaSetNodeIDs[pod.OwnerUID]
		if !exists {
			continue
		}

		err := resourceGraph.AddEdge(graph.Edge{
			From:         replicaSetNodeID,
			To:           podNode.ID,
			Relationship: graph.Owns,
		})
		if err != nil {
			return nil, fmt.Errorf("connect replicaset to pod: %w", err)
		}
	}

	// 9. Process Services & build lookup map
	serviceNodeIDs := make(map[string]string)
	matchedServices := make([]domain.Service, 0)

	for _, svc := range allServices {
		serviceMatched := false

		for _, pod := range pods {
			if !selectorMatches(svc.Selector, pod.Labels) {
				continue
			}

			serviceKey := svc.Namespace + "/" + svc.Name
			serviceNodeID, exists := serviceNodeIDs[serviceKey]
			if !exists {
				node := graph.NewNode("Service", svc.Namespace, svc.Name)
				resourceGraph.AddNode(node)
				serviceNodeIDs[serviceKey] = node.ID
				serviceNodeID = node.ID
			}

			if !serviceMatched {
				matchedServices = append(matchedServices, svc)
				serviceMatched = true
			}

			podNodeID, exists := podNodeIDs[pod.UID]
			if !exists {
				continue
			}

			err := resourceGraph.AddEdge(graph.Edge{
				From:         serviceNodeID,
				To:           podNodeID,
				Relationship: graph.Selects,
			})
			if err != nil {
				return nil, fmt.Errorf("connect service to pod: %w", err)
			}
		}
	}

	// 10. Process Ingress -> Service edges
	matchedIngresses := make([]domain.Ingress, 0)
	matchedIngressNames := make(map[string]struct{})

	for _, ingress := range allIngresses {
		ingressMatched := false

		for _, backend := range ingress.Backends {
			serviceKey := ingress.Namespace + "/" + backend.ServiceName

			serviceNodeID, exists := serviceNodeIDs[serviceKey]
			if !exists {
				continue
			}

			if !ingressMatched {
				key := ingress.Namespace + "/" + ingress.Name
				if _, alreadyAdded := matchedIngressNames[key]; !alreadyAdded {
					matchedIngresses = append(matchedIngresses, ingress)
					matchedIngressNames[key] = struct{}{}
				}
				ingressMatched = true
			}

			ingressNode := graph.NewNode("Ingress", ingress.Namespace, ingress.Name)
			resourceGraph.AddNode(ingressNode)

			err := resourceGraph.AddEdge(graph.Edge{
				From:         ingressNode.ID,
				To:           serviceNodeID,
				Relationship: graph.RoutesTo,
			})
			if err != nil {
				return nil, fmt.Errorf("connect ingress to service: %w", err)
			}
		}
	}

	// 11. Run structural calculation targeting the complete system impact
	impactReport, err := AnalyzeImpact(resourceGraph, deploymentNode.ID)
	if err != nil {
		return nil, fmt.Errorf("analyze blast radius: %w", err)
	}

	return &domain.DeploymentAnalysis{
		Deployment:  deployment,
		ReplicaSets: replicaSets,
		Pods:        pods,
		Services:    matchedServices,
		Ingresses:   matchedIngresses,
		Graph:       resourceGraph,
		Impact:      impactReport,
	}, nil
}
