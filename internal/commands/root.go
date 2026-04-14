package commands

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/database"

	"github.com/spf13/cobra"
)

// Set at build time via -ldflags "-X ...commands.gitRelease=..." and "-X ...commands.gitCommit=..."
// Falls back to Go module build info for `go install` users.
var (
	gitCommit  = "unknown"
	gitRelease = "dev"
)

func init() {
	if gitRelease != "dev" {
		return
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if info.Main.Version != "" {
		gitRelease = info.Main.Version
	}
}

var artifactsDirFlag string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "kpi-collector",
	Version: gitRelease,
	Short:   "KPI Collection and Visualization Tool",
	Long: `A tool to automate metrics gathering and visualization for KPIs 
in disconnected environments. Supports Kubernetes auto-discovery, 
Prometheus/Thanos integration, and multiple database backends.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if artifactsDirFlag != "" {
			database.OutputDir = artifactsDirFlag
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and commit information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kpi-collector %s (%s)\n", gitRelease, gitCommit)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&artifactsDirFlag, "artifacts-dir", "",
		"directory for storing artifacts: database, logs, and Grafana config (default: ./kpi-collector-artifacts/)")
	rootCmd.AddCommand(versionCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
