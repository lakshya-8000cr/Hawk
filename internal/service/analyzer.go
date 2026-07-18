package service

import (
	"context"
	"fmt"

	"hawk/internal/domain"
	"hawk/internal/graph"
	"hawk/internal/repositry"

	corev1 "k8s.io/api/core/v1"
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

func (a *Analyzer) AnalyzeDeployment(ctx context.Context, namespace, name string) (*domain.DeploymentAnalysis, error) {
	// 1. Fetch Resources
	deployment, err := a.deployments.Get(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("analyze deployment: %w", err)
	}

	replicaSets, err := a.replicaSets.FindOwnedBy(ctx, namespace, deployment.UID)
	if err != nil {
		return nil, fmt.Errorf("collect owned replicasets: %w", err)
	}

	replicaSetUIDs := make([]string, 0, len(replicaSets))
	for _, rs := range replicaSets {
		replicaSetUIDs = append(replicaSetUIDs, rs.UID)
	}

	pods, err := a.pods.FindOwnedBy(ctx, namespace, replicaSetUIDs)
	if err != nil {
		return nil, fmt.Errorf("collect owned pods: %w", err)
	}

	allServices, err := a.services.List(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("collect services: %w", err)
	}

	allIngresses, err := a.ingresses.List(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("collect ingresses: %w", err)
	}

	configMaps, secrets := collectConfigurationDependencies(deployment)
	persistentVolumeClaims := collectPersistentVolumeClaims(deployment)

	// 2. Initialize Resource Graph
	resourceGraph := graph.New()
	deploymentNode := graph.NewNode("Deployment", deployment.Namespace, deployment.Name)
	resourceGraph.AddNode(deploymentNode)

	replicaSetNodeIDs := make(map[string]string)
	podNodeIDs := make(map[string]string)

	for _, rs := range replicaSets {
		rsNode := graph.NewNode("ReplicaSet", rs.Namespace, rs.Name)
		resourceGraph.AddNode(rsNode)
		replicaSetNodeIDs[rs.UID] = rsNode.ID

		if err := resourceGraph.AddEdge(graph.Edge{From: deploymentNode.ID, To: rsNode.ID, Relationship: graph.Owns}); err != nil {
			return nil, fmt.Errorf("connect deployment to replicaset: %w", err)
		}
	}

	for _, pod := range pods {
		podNode := graph.NewNode("Pod", pod.Namespace, pod.Name)
		resourceGraph.AddNode(podNode)
		podNodeIDs[pod.UID] = podNode.ID

		if rsNodeID, exists := replicaSetNodeIDs[pod.OwnerUID]; exists {
			if err := resourceGraph.AddEdge(graph.Edge{From: rsNodeID, To: podNode.ID, Relationship: graph.Owns}); err != nil {
				return nil, fmt.Errorf("connect replicaset to pod: %w", err)
			}
		}
	}

	// Helper function to link ConfigMap/Secret dependencies to Pods (reduces copy-paste code)
	addPodDependencies := func(nodeID string) {
		for _, pod := range pods {
			if podID, ok := podNodeIDs[pod.UID]; ok {
				resourceGraph.AddEdge(graph.Edge{From: nodeID, To: podID, Relationship: graph.Uses})
			}
		}
	}

	for _, cm := range configMaps {
		cmNode := graph.NewNode("ConfigMap", cm.Namespace, cm.Name)
		resourceGraph.AddNode(cmNode)
		addPodDependencies(cmNode.ID)
	}

	for _, sec := range secrets {
		secNode := graph.NewNode("Secret", sec.Namespace, sec.Name)
		resourceGraph.AddNode(secNode)
		addPodDependencies(secNode.ID)
	}

	// 3. Process Services & build lookup map
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

			if podNodeID, exists := podNodeIDs[pod.UID]; exists {
				if err := resourceGraph.AddEdge(graph.Edge{From: serviceNodeID, To: podNodeID, Relationship: graph.Selects}); err != nil {
					return nil, fmt.Errorf("connect service to pod: %w", err)
				}
			}
		}
	}

	// 4. Process Ingress -> Service edges
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

			if err := resourceGraph.AddEdge(graph.Edge{From: ingressNode.ID, To: serviceNodeID, Relationship: graph.RoutesTo}); err != nil {
				return nil, fmt.Errorf("connect ingress to service: %w", err)
			}
		}
	}

	// 5. Run structural calculation targeting the complete system impact
	impactReport, err := AnalyzeImpact(resourceGraph, deploymentNode.ID)
	if err != nil {
		return nil, fmt.Errorf("analyze blast radius: %w", err)
	}

	return &domain.DeploymentAnalysis{
		Deployment:             deployment,
		ReplicaSets:            replicaSets,
		Pods:                   pods,
		Services:               matchedServices,
		Ingresses:              matchedIngresses,
		Graph:                  resourceGraph,
		Impact:                 impactReport,
		Secrets:                secrets,
		PersistentVolumeClaims: persistentVolumeClaims,
	}, nil
}

func collectPersistentVolumeClaims(deployment *domain.Deployment) []domain.PersistentVolumeClaim {
	names := make(map[string]struct{})
	for _, volume := range deployment.Template.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName != "" {
			names[volume.PersistentVolumeClaim.ClaimName] = struct{}{}
		}
	}

	claims := make([]domain.PersistentVolumeClaim, 0, len(names))
	for name := range names {
		claims = append(claims, domain.PersistentVolumeClaim{Name: name, Namespace: deployment.Namespace})
	}
	return claims
}

func collectConfigurationDependencies(deployment *domain.Deployment) ([]domain.ConfigMap, []domain.Secret) {
	configMapNames := make(map[string]struct{})
	secretNames := make(map[string]struct{})
	spec := deployment.Template.Spec

	for _, container := range spec.InitContainers {
		collectContainerConfiguration(container, configMapNames, secretNames)
	}
	for _, container := range spec.Containers {
		collectContainerConfiguration(container, configMapNames, secretNames)
	}

	for _, volume := range spec.Volumes {
		if volume.ConfigMap != nil && volume.ConfigMap.Name != "" {
			configMapNames[volume.ConfigMap.Name] = struct{}{}
		}
		if volume.Secret != nil && volume.Secret.SecretName != "" {
			secretNames[volume.Secret.SecretName] = struct{}{}
		}
	}

	configMaps := make([]domain.ConfigMap, 0, len(configMapNames))
	for name := range configMapNames {
		configMaps = append(configMaps, domain.ConfigMap{Name: name, Namespace: deployment.Namespace})
	}

	secrets := make([]domain.Secret, 0, len(secretNames))
	for name := range secretNames {
		secrets = append(secrets, domain.Secret{Name: name, Namespace: deployment.Namespace})
	}

	return configMaps, secrets
}

func collectContainerConfiguration(container corev1.Container, configMapNames, secretNames map[string]struct{}) {
	for _, envFrom := range container.EnvFrom {
		if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name != "" {
			configMapNames[envFrom.ConfigMapRef.Name] = struct{}{}
		}
		if envFrom.SecretRef != nil && envFrom.SecretRef.Name != "" {
			secretNames[envFrom.SecretRef.Name] = struct{}{}
		}
	}

	for _, env := range container.Env {
		if env.ValueFrom == nil {
			continue
		}
		if env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name != "" {
			configMapNames[env.ValueFrom.ConfigMapKeyRef.Name] = struct{}{}
		}
		if env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name != "" {
			secretNames[env.ValueFrom.SecretKeyRef.Name] = struct{}{}
		}
	}
}