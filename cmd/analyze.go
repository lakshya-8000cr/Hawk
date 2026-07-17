package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hawk/internal/kube"
	"hawk/internal/repositry" // Fixed typo from repositry -> repository
	"hawk/internal/service"

	"github.com/spf13/cobra"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorCyan   = "\033[36m"
	colorBlue   = "\033[34m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
)

var namespace string

var analyzeCmd = &cobra.Command{
	Use:   "analyze deployment <name>",
	Short: "Analyze a Kubernetes resource blast radius",
	Args:  cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		resourceType := strings.ToLower(args[0])
		resourceName := args[1]

		if resourceType != "deployment" &&
			resourceType != "deploy" &&
			resourceType != "deployments" {
			return fmt.Errorf(
				"unsupported resource type %q; currently supported: deployment",
				resourceType,
			)
		}

		client, err := kube.NewClient()
		if err != nil {
			return err
		}

		deploymentRepo := repository.NewKubernetesDeploymentRepository(client)
		replicaSetRepo := repository.NewKubernetesReplicaSetRepository(client)
		podRepo := repository.NewKubernetesPodRepository(client)
		serviceRepo := repository.NewKubernetesServiceRepository(client)
		ingressRepo := repository.NewKubernetesIngressRepository(client)

		analyzer := service.NewAnalyzer(
			deploymentRepo,
			replicaSetRepo,
			podRepo,
			serviceRepo,
			ingressRepo,
		)

		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		result, err := analyzer.AnalyzeDeployment(
			ctx,
			namespace,
			resourceName,
		)
		if err != nil {
			return err
		}

		printAnalysis(result)
		return nil
	},
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func init() {
	analyzeCmd.Flags().StringVarP(
		&namespace,
		"namespace",
		"n",
		"default",
		"Kubernetes namespace",
	)

	rootCmd.AddCommand(analyzeCmd)
}
