package repository

import (
	"context"
	"fmt"
	"strconv"

	"hawk/internal/domain"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type IngressRepository interface {
	List(
		ctx context.Context,
		namespace string,
	) ([]domain.Ingress, error)
}

type KubernetesIngressRepository struct {
	client kubernetes.Interface
}

func NewKubernetesIngressRepository(
	client kubernetes.Interface,
) *KubernetesIngressRepository {
	return &KubernetesIngressRepository{
		client: client,
	}
}

func (r *KubernetesIngressRepository) List(
	ctx context.Context,
	namespace string,
) ([]domain.Ingress, error) {
	list, err := r.client.
		NetworkingV1().
		Ingresses(namespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf(
			"list ingresses in namespace %s: %w",
			namespace,
			err,
		)
	}

	ingresses := make([]domain.Ingress, 0, len(list.Items))

	for _, item := range list.Items {
		ingress := domain.Ingress{
			Name:      item.Name,
			Namespace: item.Namespace,
		}

		if item.Spec.IngressClassName != nil {
			ingress.ClassName = *item.Spec.IngressClassName
		}

		// Optional default backend.
		if item.Spec.DefaultBackend != nil &&
			item.Spec.DefaultBackend.Service != nil {

			ingress.Backends = append(
				ingress.Backends,
				domain.IngressBackend{
					ServiceName: item.Spec.DefaultBackend.Service.Name,
					ServicePort: ingressServicePort(
						item.Spec.DefaultBackend.Service.Port.Name,
						item.Spec.DefaultBackend.Service.Port.Number,
					),
					Host: "*",
					Path: "/",
				},
			)
		}

		// Rule-based backends.
		for _, rule := range item.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}

			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service == nil {
					continue
				}

				ingress.Backends = append(
					ingress.Backends,
					domain.IngressBackend{
						ServiceName: path.Backend.Service.Name,
						ServicePort: ingressServicePort(
							path.Backend.Service.Port.Name,
							path.Backend.Service.Port.Number,
						),
						Host: rule.Host,
						Path: path.Path,
					},
				)
			}
		}

		for _, tls := range item.Spec.TLS {
			ingress.TLSHosts = append(
				ingress.TLSHosts,
				tls.Hosts...,
			)
		}

		for _, lb := range item.Status.LoadBalancer.Ingress {
			if lb.IP != "" {
				ingress.LoadBalancerAddresses = append(
					ingress.LoadBalancerAddresses,
					lb.IP,
				)
			}

			if lb.Hostname != "" {
				ingress.LoadBalancerAddresses = append(
					ingress.LoadBalancerAddresses,
					lb.Hostname,
				)
			}
		}

		ingresses = append(ingresses, ingress)
	}

	return ingresses, nil
}

func ingressServicePort(
	name string,
	number int32,
) string {
	if name != "" {
		return name
	}

	if number != 0 {
		return strconv.Itoa(int(number))
	}

	return "unknown"
}
