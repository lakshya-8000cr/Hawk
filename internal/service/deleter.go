package service

import (
	"context"
	"fmt"

	"hawk/internal/repositry" // Fixed typo completely
)

type Deleter struct {
	deployments repository.DeploymentRepository
}

func NewDeleter(
	deployments repository.DeploymentRepository,
) *Deleter {
	return &Deleter{
		deployments: deployments,
	}
}

func (d *Deleter) DeleteDeployment(
	ctx context.Context,
	namespace string,
	name string,
) error {
	if err := d.deployments.Delete(
		ctx,
		namespace,
		name,
	); err != nil {
		return fmt.Errorf(
			"delete deployment: %w",
			err,
		)
	}

	return nil
}