package repository

import (
	"context"
	"fmt"

	"hawk/internal/domain"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DeploymentRepository interface {
	Get(
		ctx context.Context,
		namespace string,
		name string,
	) (*domain.Deployment, error)
}

type KubernetesDeploymentRepository struct {
	client kubernetes.Interface
}

func NewKubernetesDeploymentRepository(
	client kubernetes.Interface,
) *KubernetesDeploymentRepository {
	return &KubernetesDeploymentRepository{
		client: client,
	}
}

func (r *KubernetesDeploymentRepository) Get(
	ctx context.Context,
	namespace string,
	name string,
) (*domain.Deployment, error) {
	deployment, err := r.client.
		AppsV1().
		Deployments(namespace).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		return nil, fmt.Errorf(
			"get deployment %s/%s: %w",
			namespace,
			name,
			err,
		)
	}

	var desired int32
	if deployment.Spec.Replicas != nil {
		desired = *deployment.Spec.Replicas
	}

	return &domain.Deployment{
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		UID:               string(deployment.UID),
		DesiredReplicas:   desired,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		Labels:            deployment.Labels,
		Selector:          deployment.Spec.Selector.MatchLabels,
	}, nil
}
