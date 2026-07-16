package domain

import "hawk/internal/graph"

type DeploymentAnalysis struct {
	Deployment  *Deployment
	ReplicaSets []ReplicaSet
	Pods        []Pod
	Services    []Service
	Ingresses   []Ingress
	Graph       *graph.Graph
	Impact      *ImpactReport
}
// typically micrservice behaviour
