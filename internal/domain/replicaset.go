package domain

type ReplicaSet struct {
	Name       string
	Namespace  string
	UID        string
	OwnerUID   string
	OwnerKind  string
	OwnerName  string
	Replicas   int32
	ReadyCount int32
}  // this will contain all the var related to the replica , like microservice architecture