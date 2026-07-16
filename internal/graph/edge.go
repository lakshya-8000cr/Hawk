package graph

type Relationship string

const (
	Owns     Relationship = "OWNS"
	Selects  Relationship = "SELECTS"
	RoutesTo Relationship = "ROUTES_TO"
	Targets  Relationship = "TARGETS"
	Uses     Relationship = "USES"
)

type Edge struct {
	From         string
	To           string
	Relationship Relationship
}