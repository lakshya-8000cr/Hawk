package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"hawk/internal/kube"
	"hawk/internal/repositry"
	"hawk/internal/service"

	"github.com/spf13/cobra"
)

var (
	deleteYes    bool
	deleteDryRun bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete deployment <name>",
	Short: "Analyze impact and safely delete a Kubernetes deployment",
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

		deleter := service.NewDeleter(deploymentRepo)

		ctx, cancel := context.WithTimeout(
			context.Background(),
			20*time.Second,
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

		// Same UI as analyze command.
		printAnalysis(result)

		if deleteDryRun {
			fmt.Printf(
				"%sDry run complete. No resources were deleted.%s\n",
				colorYellow,
				colorReset,
			)

			return nil
		}

		if !deleteYes {
			confirmed, err := confirmDeletion(
				result.Deployment.Name,
				cmd,
			)
			if err != nil {
				return err
			}

			if !confirmed {
				fmt.Printf(
					"\n%sDeletion cancelled.%s\n",
					colorYellow,
					colorReset,
				)

				return nil
			}
		}

		fmt.Printf(
			"\n%sDeleting Deployment/%s...%s\n",
			colorCyan,
			resourceName,
			colorReset,
		)

		if err := deleter.DeleteDeployment(
			ctx,
			namespace,
			resourceName,
		); err != nil {
			return err
		}

		fmt.Printf(
			"%s✓ Deployment/%s deleted successfully.%s\n",
			colorGreen,
			resourceName,
			colorReset,
		)

		return nil
	},
}

func confirmDeletion(
	deploymentName string,
	cmd *cobra.Command,
) (bool, error) {
	fmt.Printf(
		"%s--------------------------------------------------------%s\n",
		colorDim,
		colorReset,
	)

	fmt.Printf(
		"\n%s%sDelete Deployment/%s?%s\n",
		colorBold,
		colorRed,
		deploymentName,
		colorReset,
	)

	fmt.Print("Proceed with deletion? [y/N]: ")

	reader := bufio.NewReader(cmd.InOrStdin())

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf(
			"read confirmation: %w",
			err,
		)
	}

	input = strings.ToLower(strings.TrimSpace(input))

	return input == "y" || input == "yes", nil
}

func init() {
	deleteCmd.Flags().BoolVarP(
		&deleteYes,
		"yes",
		"y",
		false,
		"Skip deletion confirmation",
	)

	deleteCmd.Flags().BoolVar(
		&deleteDryRun,
		"dry-run",
		false,
		"Analyze impact without deleting the deployment",
	)

	deleteCmd.Flags().StringVarP(
		&namespace,
		"namespace",
		"n",
		"default",
		"Kubernetes namespace",
	)

	deleteCmd.SetIn(os.Stdin)

	rootCmd.AddCommand(deleteCmd)
}
