package cmd

import (
	"fmt"
	"strings"
	"time"

	"hawk/internal/domain"
)

func printAnalysis(result *domain.DeploymentAnalysis) {
	deployment := result.Deployment

	// --- 1. PRO-LOOK GEOMETRIC HAWK BIRD IN FLIGHT ASCII BANNER ---
	fmt.Println()
	fmt.Printf("%s                /\\   \n", colorCyan)
	fmt.Printf("%s|\\ _          _/  \\_           ____/|   %s  _    _    _     __  __ %s\n", colorCyan, colorBold, colorReset)
	fmt.Printf("%s \\_ \\__       \\    /       ___/ ___/    %s | |  | |  / \\    | |/ / %s.io\n", colorCyan, colorBold, colorReset)
	fmt.Printf("%s   \\__ \\____   \\  /   ____/ ___/        %s | |__| | / _ \\   | ' /  %s\n", colorCyan, colorBold, colorReset)
	fmt.Printf("%s      \\____    /  \\    ____/            %s |  __  |/ ___ \\  | . \\  %s\n", colorDim, colorBold, colorReset)
	fmt.Printf("%s           \\__/    \\__/                 %s |_|  |_/_/   \\_\\|_|\\_\\ %s\n", colorDim, colorBold, colorReset)
	fmt.Printf("%s              |    |                %s\n", colorDim, colorReset)
	fmt.Printf("%s              |    |         %s\n", colorDim, colorReset)
	fmt.Printf("%s               \\   /\n", colorDim)
	fmt.Printf("%s                \\ /\n", colorDim)
	fmt.Println()

	// --- 2. EXECUTION CONTEXT BLOCK ---
	fmt.Printf("%sexecution:%s clusterscope\n", colorCyan, colorReset)
	fmt.Printf("   %starget:%s Deployment/%s\n", colorDim, colorReset, deployment.Name)
	fmt.Printf("%snamespace:%s %s\n", colorDim, colorReset, deployment.Namespace)
	fmt.Printf("     %stime:%s %s\n", colorDim, colorReset, time.Now().Format(time.RFC1123))
	fmt.Println()
	fmt.Printf("%s--------------------------------------------------------%s\n", colorDim, colorReset)
	fmt.Println()

	// --- 3. DIRECTLY OWNED INFRASTRUCTURE ---
	fmt.Printf("%s%s[1] Directly Owned Resources%s\n", colorBold, colorCyan, colorReset)

	if len(result.ReplicaSets) == 0 {
		fmt.Printf(
			"  %sNo ReplicaSets managed by this deployment%s\n",
			colorDim,
			colorReset,
		)
	} else {
		for _, rs := range result.ReplicaSets {
			fmt.Printf(
				"  ├── %sReplicaSet/%s%s  replicas(%s%d%s) ready(%s%d%s)\n",
				colorBold,
				rs.Name,
				colorReset,
				colorYellow,
				rs.Replicas,
				colorReset,
				colorGreen,
				rs.ReadyCount,
				colorReset,
			)

			podFound := false

			for _, pod := range result.Pods {
				if pod.OwnerUID != rs.UID {
					continue
				}

				podFound = true

				phaseColor := colorGreen
				if pod.Phase != "Running" {
					phaseColor = colorYellow
				}

				fmt.Printf(
					"  │   └── %sPod/%s%s  phase(%s%s%s) ready(%t) restarts(%s%d%s)\n",
					colorDim,
					pod.Name,
					colorReset,
					phaseColor,
					pod.Phase,
					colorReset,
					pod.Ready,
					colorYellow,
					pod.Restarts,
					colorReset,
				)
			}

			if !podFound {
				fmt.Printf(
					"  │   └── %sNo active Pods discovered%s\n",
					colorDim,
					colorReset,
				)
			}
		}
	}

	fmt.Println()

	// --- 4. NETROUTING REFERENCE BY SERVICES ---
	fmt.Printf("%s%s[2] Service Mesh Routing%s\n", colorBold, colorCyan, colorReset)

	if len(result.Services) == 0 {
		fmt.Printf(
			"  %sNo Core Services targeting these pods%s\n",
			colorDim,
			colorReset,
		)
	} else {
		for _, svc := range result.Services {
			fmt.Printf(
				"  └── %sService/%s%s  type(%s%s%s) clusterIP(%s%s%s)\n",
				colorBold,
				svc.Name,
				colorReset,
				colorCyan,
				svc.Type,
				colorReset,
				colorDim,
				svc.ClusterIP,
				colorReset,
			)

			fmt.Printf(
				"      %sselector: %v%s\n",
				colorDim,
				svc.Selector,
				colorReset,
			)
		}
	}

	// --- 5. EDGE TRAFFIC EXPOSURE ---
	fmt.Println()
	fmt.Printf("%s%s[3] Ingress Traffic Footprint%s\n", colorBold, colorCyan, colorReset)

	if len(result.Ingresses) == 0 {
		fmt.Printf(
			"  %sNo Ingress endpoints routes mapped%s\n",
			colorDim,
			colorReset,
		)
	} else {
		for _, ingress := range result.Ingresses {
			fmt.Printf(
				"  └── %sIngress/%s%s  class(%s%s%s)\n",
				colorBold,
				ingress.Name,
				colorReset,
				colorCyan,
				valueOrDefault(ingress.ClassName, "default"),
				colorReset,
			)

			for _, backend := range ingress.Backends {
				for _, svc := range result.Services {
					if backend.ServiceName != svc.Name {
						continue
					}

					host := valueOrDefault(backend.Host, "*")
					path := valueOrDefault(backend.Path, "/")

					fmt.Printf(
						"      %s%s%s%s%s → %sService/%s:%s%s\n",
						colorGreen,
						host,
						colorReset,
						colorBold,
						path,
						colorReset,
						colorDim,
						backend.ServiceName,
						backend.ServicePort,
						colorReset,
					)
				}
			}

			if len(ingress.TLSHosts) > 0 {
				fmt.Printf(
					"      %sTLS hosts: %v%s\n",
					colorDim,
					ingress.TLSHosts,
					colorReset,
				)
			}

			if len(ingress.LoadBalancerAddresses) > 0 {
				fmt.Printf(
					"      %sEndpoints: %v%s\n",
					colorDim,
					ingress.LoadBalancerAddresses,
					colorReset,
				)
			}
		}
	}


	// --- 6. CONFIGURATION DEPENDENCIES ---
fmt.Println()
fmt.Printf("%s%s[4] Configuration Dependencies%s\n", colorBold, colorCyan, colorReset)

if len(result.ConfigMaps) == 0 {
	fmt.Printf(
		"  %sNo ConfigMap dependencies detected%s\n",
		colorDim,
		colorReset,
	)
} else {
	for _, configMap := range result.ConfigMaps {
		fmt.Printf(
			"  └── %sConfigMap/%s%s\n",
			colorBold,
			configMap.Name,
			colorReset,
		)
	}
}

// --- 7. SECRET DEPENDENCIES ---
fmt.Println()
fmt.Printf("%s%s[5] Secret Dependencies%s\n", colorBold, colorCyan, colorReset)

if len(result.Secrets) == 0 {
	fmt.Printf(
		"  %sNo Secret dependencies detected%s\n",
		colorDim,
		colorReset,
	)
} else {
	for _, secret := range result.Secrets {
		fmt.Printf(
			"  └── %sSecret/%s%s\n",
			colorBold,
			secret.Name,
			colorReset,
		)
	}
}

	// --- 6. DYNAMIC BLAST RADIUS RADAR SYSTEM ---
	fmt.Println()
	fmt.Printf("%s--------------------------------------------------------%s\n", colorDim, colorReset)
	fmt.Println()
	fmt.Printf("%s%s[!] Blast Radius Radar%s\n", colorBold, colorRed, colorReset)
	fmt.Println()

	riskColor := colorGreen
	riskText := strings.ToUpper(string(result.Impact.Risk))

	if riskText == "HIGH" || riskText == "CRITICAL" {
		riskColor = colorRed
	} else if riskText == "MEDIUM" || riskText == "WARNING" {
		riskColor = colorYellow
	}

	fmt.Printf(
		"  Risk Profile : %s%s%s%s\n",
		colorBold,
		riskColor,
		riskText,
		colorReset,
	)

	fmt.Printf("  Summary      : %s\n", result.Impact.Summary)
	fmt.Println()
	fmt.Printf("  %sImpact Cascading Tree:%s\n", colorBold, colorReset)

	if len(result.Impact.Affected) == 0 {
		fmt.Printf(
			"    %sIsolated workload. No cascading impact risks detected.%s\n",
			colorDim,
			colorReset,
		)
	} else {
		for _, resource := range result.Impact.Affected {
			fmt.Printf(
				"    ├── %s%s/%s%s  relation(%s%s%s)\n",
				colorBold,
				resource.Kind,
				resource.Name,
				colorReset,
				colorYellow,
				resource.Relationship,
				colorReset,
			)
		}
	}

	fmt.Println()
}
