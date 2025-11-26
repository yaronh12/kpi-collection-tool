package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kpi-collector",
	Short: "KPI Collection and Visualization Tool",
	Long: `A tool to automate metrics gathering and visualization for KPIs 
in disconnected environments. Supports Kubernetes auto-discovery, 
Prometheus/Thanos integration, and multiple database backends.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
