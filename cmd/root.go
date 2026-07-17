package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-hawk",
	Short: "Dependency-aware Kubernetes impact analysis",
	Long: `Hawk analyzes relationships between Kubernetes resources
before potentially destructive operations are performed.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
