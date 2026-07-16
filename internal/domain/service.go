package domain

type Service struct {
	Name      string
	Namespace string
	Selector  map[string]string
	Type      string
	ClusterIP string
}
