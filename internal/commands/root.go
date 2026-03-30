package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/database"

	"github.com/spf13/cobra"
)

var artifactDirFlag string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kpi-collector",
	Short: "KPI Collection and Visualization Tool",
	Long: `A tool to automate metrics gathering and visualization for KPIs 
in disconnected environments. Supports Kubernetes auto-discovery, 
Prometheus/Thanos integration, and multiple database backends.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if artifactDirFlag != "" {
			database.OutputDir = filepath.Join(artifactDirFlag, database.DefaultOutputDir)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&artifactDirFlag, "artifact-dir", "",
		"parent directory for the kpi-collector-artifacts folder (default: current directory)")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
