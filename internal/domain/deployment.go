package domain // this is typically microservices architecture
//eve for the structure  we have made a diff file

import corev1 "k8s.io/api/core/v1"
type Deployment struct {
	Name              string
	Namespace         string
	UID               string
	DesiredReplicas   int32
	AvailableReplicas int32
	Labels            map[string]string
	Selector          map[string]string
	Template corev1.PodTemplateSpec
}
