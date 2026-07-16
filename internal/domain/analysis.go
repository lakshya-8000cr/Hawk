package domain

import "hawk/internal/graph"

type DeploymentAnalysis struct {
	Deployment  *Deployment
	ReplicaSets []ReplicaSet
	Pods        []Pod
	Graph       *graph.Graph
}

// typically micrservice behaviour
