package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hawk/internal/kube"
	"hawk/internal/repositry"
	"hawk/internal/service"

	"github.com/spf13/cobra"
)

var namespace string

var analyzeCmd = &cobra.Command{
	Use:   "analyze deployment <name>",
	Short: "Analyze a Kubernetes resource",
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

deploymentRepo :=
	repository.NewKubernetesDeploymentRepository(client)

replicaSetRepo :=
	repository.NewKubernetesReplicaSetRepository(client)

analyzer := service.NewAnalyzer(
	deploymentRepo,
	replicaSetRepo,
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

deployment := result.Deployment

fmt.Println()
fmt.Println("HAWK   Impact Analysis")
fmt.Println()
fmt.Printf("Target: Deployment/%s\n", deployment.Name)
fmt.Printf("Namespace: %s\n", deployment.Namespace)
fmt.Println()

fmt.Println("Directly owned:")

if len(result.ReplicaSets) == 0 {
	fmt.Println("  No ReplicaSets found")
} else {
	for _, rs := range result.ReplicaSets {
		fmt.Printf(
			"  └── ReplicaSet/%s  replicas(%d) ready(%d)\n",
			rs.Name,
			rs.Replicas,
			rs.ReadyCount,
		)
	}
}

fmt.Println()

		return nil
	},
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