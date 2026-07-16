package domain

type Pod struct {
	Name      string
	Namespace string
	UID       string

	OwnerUID  string
	OwnerKind string
	OwnerName string

	Phase    string
	NodeName string
	Ready    bool
	Restarts int32
}
