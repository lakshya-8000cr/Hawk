package service

import (
	"fmt"

	"hawk/internal/domain"
	"hawk/internal/graph"
)

func AnalyzeImpact(
	resourceGraph *graph.Graph,
	targetID string,
) (*domain.ImpactReport, error) {
	if _, exists := resourceGraph.GetNode(targetID); !exists {
		return nil, fmt.Errorf(
			"target node %q does not exist in graph",
			targetID,
		)
	}

	report := &domain.ImpactReport{
		Risk:     domain.RiskLow,
		Affected: make([]domain.ImpactResource, 0),
	}

	visited := make(map[string]bool)
	queue := []string{targetID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}

		visited[currentID] = true

		// Follow resources owned/used by current node.
		for _, edge := range resourceGraph.GetOutgoing(currentID) {
			appendImpactResource(
				report,
				resourceGraph,
				edge.To,
				string(edge.Relationship),
			)

			if !visited[edge.To] {
				queue = append(queue, edge.To)
			}
		}

		// Follow resources referring to current node.
		for _, edge := range resourceGraph.GetIncoming(currentID) {
			appendImpactResource(
				report,
				resourceGraph,
				edge.From,
				string(edge.Relationship),
			)

			if !visited[edge.From] {
				queue = append(queue, edge.From)
			}
		}
	}

	evaluateRisk(report)

	return report, nil
}

func appendImpactResource(
	report *domain.ImpactReport,
	resourceGraph *graph.Graph,
	nodeID string,
	relationship string,
) {
	node, exists := resourceGraph.GetNode(nodeID)
	if !exists {
		return
	}

	for _, existing := range report.Affected {
		if existing.Kind == node.Kind &&
			existing.Name == node.Name &&
			existing.Namespace == node.Namespace {
			return
		}
	}

	report.Affected = append(
		report.Affected,
		domain.ImpactResource{
			Kind:         node.Kind,
			Name:         node.Name,
			Namespace:    node.Namespace,
			Relationship: relationship,
		},
	)
}

func evaluateRisk(report *domain.ImpactReport) {
	hasService := false
	hasIngress := false

	for _, resource := range report.Affected {
		switch resource.Kind {
		case "Ingress":
			hasIngress = true
			report.ExternalAccess = true

		case "Service":
			hasService = true
		}
	}

	switch {
	case hasIngress:
		report.Risk = domain.RiskHigh
		report.Summary =
			"External traffic reaches this workload through an Ingress."

	case hasService:
		report.Risk = domain.RiskMedium
		report.Summary =
			"One or more Services currently route traffic to this workload."

	default:
		report.Risk = domain.RiskLow
		report.Summary =
			"No Service or Ingress dependencies were discovered."
	}
}
