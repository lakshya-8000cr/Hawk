package domain

import "hawk/internal/graph"

type DeploymentAnalysis struct {
	Deployment  *Deployment
	ReplicaSets []ReplicaSet
	Pods        []Pod
	Services    []Service
	Graph       *graph.Graph
	Ingresses   []Ingress
}

// typically micrservice behaviour
