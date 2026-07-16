package repository

import (
	"context"
	"fmt"

	"hawk/internal/domain"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ReplicaSetRepository interface {
	FindOwnedBy(
		ctx context.Context,
		namespace string,
		ownerUID string,
	) ([]domain.ReplicaSet, error)
}

type KubernetesReplicaSetRepository struct {
	client kubernetes.Interface
}

func NewKubernetesReplicaSetRepository(
	client kubernetes.Interface,
) *KubernetesReplicaSetRepository {
	return &KubernetesReplicaSetRepository{
		client: client,
	}
}

func (r *KubernetesReplicaSetRepository) FindOwnedBy(
	ctx context.Context,
	namespace string,
	ownerUID string,
) ([]domain.ReplicaSet, error) {
	list, err := r.client.
		AppsV1().
		ReplicaSets(namespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf(
			"list replicasets in namespace %s: %w",
			namespace,
			err,
		)
	}

	replicaSets := make([]domain.ReplicaSet, 0)

	for _, rs := range list.Items {
		for _, owner := range rs.OwnerReferences {
			if string(owner.UID) != ownerUID {
				continue
			}

			replicaSets = append(replicaSets, domain.ReplicaSet{
				Name:       rs.Name,
				Namespace:  rs.Namespace,
				UID:        string(rs.UID),
				OwnerUID:   string(owner.UID),
				OwnerKind:  owner.Kind,
				OwnerName:  owner.Name,
				Replicas:   rs.Status.Replicas,
				ReadyCount: rs.Status.ReadyReplicas,
			})

			break
		}
	}

	return replicaSets, nil
}