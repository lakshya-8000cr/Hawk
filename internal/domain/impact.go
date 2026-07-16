package domain

type RiskLevel string

const (
	RiskLow    RiskLevel = "LOW"
	RiskMedium RiskLevel = "MEDIUM"
	RiskHigh   RiskLevel = "HIGH"
)

type ImpactResource struct {
	Kind         string
	Name         string
	Namespace    string
	Relationship string
}

type ImpactReport struct {
	Risk           RiskLevel
	Summary        string
	Affected       []ImpactResource
	ExternalAccess bool
}