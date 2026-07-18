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
	ConfigMaps []ConfigMap
    Secrets    []Secret
	PersistentVolumeClaims []PersistentVolumeClaim
}
// typicall micrservice behaviour
