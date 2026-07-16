package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hawk/internal/kube"
	"hawk/internal/repositry" // Fixed typo in import path from "repositry"
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

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := analyzer.AnalyzeDeployment(ctx, namespace, resourceName)
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

		// --- Directly owned output ---
		fmt.Println("Directly owned:")
		if len(result.ReplicaSets) == 0 {
			fmt.Println("  No ReplicaSets found")
		} else {
			for _, rs := range result.ReplicaSets {
				fmt.Printf(
					"  ├── ReplicaSet/%s  replicas(%d) ready(%d)\n",
					rs.Name,
					rs.Replicas,
					rs.ReadyCount,
				)

				podFound := false
				for _, pod := range result.Pods {
					if pod.OwnerUID != rs.UID {
						continue
					}
					podFound = true
					fmt.Printf(
						"  │   └── Pod/%s  phase(%s) ready(%t) restarts(%d)\n",
						pod.Name,
						pod.Phase,
						pod.Ready,
						pod.Restarts,
					)
				}

				if !podFound {
					fmt.Println("  │   └── No active Pods")
				}
			}
		}

		fmt.Println()

		// --- Referenced by Services output ---
		fmt.Println("Referenced by:")
		if len(result.Services) == 0 {
			fmt.Println("  No Services currently select these Pods")
		} else {
			for _, svc := range result.Services {
				fmt.Printf(
					"  └── Service/%s  type(%s) clusterIP(%s)\n",
					svc.Name,
					svc.Type,
					svc.ClusterIP,
				)
				fmt.Printf("      selector: %v\n", svc.Selector)
			}
		}

		// --- Ingress impact output ---
		fmt.Println()
		fmt.Println("Traffic exposure:")
		if len(result.Ingresses) == 0 {
			fmt.Println("  No Ingress routes detected")
		} else {
			for _, ingress := range result.Ingresses {
				fmt.Printf(
					"  └── Ingress/%s  class(%s)\n",
					ingress.Name,
					valueOrDefault(ingress.ClassName, "default"),
				)

				for _, backend := range ingress.Backends {
					for _, svc := range result.Services {
						if backend.ServiceName != svc.Name {
							continue
						}

						host := valueOrDefault(backend.Host, "*")
						path := valueOrDefault(backend.Path, "/")

						fmt.Printf(
							"      %s%s → Service/%s:%s\n",
							host,
							path,
							backend.ServiceName,
							backend.ServicePort,
						)
					}
				}

				if len(ingress.TLSHosts) > 0 {
					fmt.Printf("      TLS hosts: %v\n", ingress.TLSHosts)
				}

				if len(ingress.LoadBalancerAddresses) > 0 {
					fmt.Printf("      Addresses: %v\n", ingress.LoadBalancerAddresses)
				}
			}
		}

		fmt.Println()
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
