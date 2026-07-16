package domain

import "hawk/internal/graph"

type DeploymentAnalysis struct {
	Deployment  *Deployment
	ReplicaSets []ReplicaSet
	Graph       *graph.Graph
} 

// typically micrservice behaviour