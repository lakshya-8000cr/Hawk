package domain

type Pod struct {
	Name      string
	Namespace string
	UID       string
	Labels    map[string]string

	OwnerUID  string
	OwnerKind string
	OwnerName string

	Phase    string
	NodeName string
	Ready    bool
	Restarts int32
}
