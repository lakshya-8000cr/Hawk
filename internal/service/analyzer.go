package service
// main analyzer which will
import (
	"context"
	"fmt"

	"hawk/internal/domain"
	"hawk/internal/repositry"
)

type Analyzer struct {
	deployments repository.DeploymentRepository
}

func NewAnalyzer(
	deployments repository.DeploymentRepository,
) *Analyzer {
	return &Analyzer{
		deployments: deployments,
	}
}

func (a *Analyzer) AnalyzeDeployment(
	ctx context.Context,
	namespace string,
	name string,
) (*domain.Deployment, error) {
	deployment, err := a.deployments.Get(
		ctx,
		namespace,
		name,
	)

	if err != nil {
		return nil, fmt.Errorf("analyze deployment: %w", err)
	}

	return deployment, nil
}