package repository

import (
	"context"
	"fmt"

	"hawk/internal/domain"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodRepository interface {
	FindOwnedBy(
		ctx context.Context,
		namespace string,
		ownerUIDs []string,
	) ([]domain.Pod, error)
}

type KubernetesPodRepository struct {
	client kubernetes.Interface
}

func NewKubernetesPodRepository(
	client kubernetes.Interface,
) *KubernetesPodRepository {
	return &KubernetesPodRepository{
		client: client,
	}
}

func (r *KubernetesPodRepository) FindOwnedBy(
	ctx context.Context,
	namespace string,
	ownerUIDs []string,
) ([]domain.Pod, error) {
	list, err := r.client.
		CoreV1().
		Pods(namespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf(
			"list pods in namespace %s: %w",
			namespace,
			err,
		)
	}

	ownerSet := make(map[string]struct{}, len(ownerUIDs))
	for _, uid := range ownerUIDs {
		ownerSet[uid] = struct{}{}
	}

	pods := make([]domain.Pod, 0)

	for _, pod := range list.Items {
		for _, owner := range pod.OwnerReferences {
			if _, exists := ownerSet[string(owner.UID)]; !exists {
				continue
			}

			var restarts int32
			for _, status := range pod.Status.ContainerStatuses {
				restarts += status.RestartCount
			}

			ready := false
			for _, condition := range pod.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					ready = true
					break
				}
			}

			pods = append(pods, domain.Pod{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				UID:       string(pod.UID),

				OwnerUID:  string(owner.UID),
				OwnerKind: owner.Kind,
				OwnerName: owner.Name,

				Phase:    string(pod.Status.Phase),
				NodeName: pod.Spec.NodeName,
				Ready:    ready,
				Restarts: restarts,
			})

			break
		}
	}

	return pods, nil
}