package repository

import (
	"context"
	"fmt"

	"hawk/internal/domain"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceRepository interface {
	List(
		ctx context.Context,
		namespace string,
	) ([]domain.Service, error)
}

type KubernetesServiceRepository struct {
	client kubernetes.Interface
}

func NewKubernetesServiceRepository(
	client kubernetes.Interface,
) *KubernetesServiceRepository {
	return &KubernetesServiceRepository{
		client: client,
	}
}

func (r *KubernetesServiceRepository) List(
	ctx context.Context,
	namespace string,
) ([]domain.Service, error) {
	list, err := r.client.
		CoreV1().
		Services(namespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf(
			"list services in namespace %s: %w",
			namespace,
			err,
		)
	}

	services := make([]domain.Service, 0, len(list.Items))

	for _, svc := range list.Items {
		services = append(services, domain.Service{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Selector:  svc.Spec.Selector,
			Type:      string(svc.Spec.Type),
			ClusterIP: svc.Spec.ClusterIP,
		})
	}

	return services, nil
}
