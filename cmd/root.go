package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "v1.0.0-dev"

var rootCmd = &cobra.Command{
	Use:   "kubectl-hawk",
	Short: "Dependency-aware Kubernetes impact analysis",
	Long: `Hawk analyzes relationships between Kubernetes resources
before potentially destructive operations are performed.`,
}


//to show the version 
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Hawk version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Hawk %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}