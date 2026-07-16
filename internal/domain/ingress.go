package domain

type IngressBackend struct {
	ServiceName string
	ServicePort string
	Host        string
	Path        string
}

type Ingress struct {
	Name                  string
	Namespace             string
	ClassName             string
	Backends              []IngressBackend
	TLSHosts              []string
	LoadBalancerAddresses []string
}
