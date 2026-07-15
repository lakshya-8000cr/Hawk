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

		analyzer := service.NewAnalyzer(deploymentRepo)

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

		fmt.Println()
		fmt.Println("HAWK   Deployment Analysis")
		fmt.Println()
		fmt.Println("Name:               ", result.Name)
		fmt.Println("Namespace:          ", result.Namespace)
		fmt.Println("UID:                ", result.UID)
		fmt.Println("Desired replicas:   ", result.DesiredReplicas)
		fmt.Println("Available replicas: ", result.AvailableReplicas)
		fmt.Println("Selector:           ", result.Selector)
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