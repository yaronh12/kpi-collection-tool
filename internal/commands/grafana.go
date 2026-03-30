package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/database"

	"github.com/spf13/cobra"
)

const grafanaContainerName = "grafana-kpi"

var grafanaCmd = &cobra.Command{
	Use:   "grafana",
	Short: "Manage Grafana dashboard for KPI visualization",
	Long: `Manage a local Grafana instance with the KPI dashboard pre-configured.
Supports both SQLite and PostgreSQL datasources.

  Use 'grafana start' to launch Grafana and 'grafana stop' to stop it.

When using SQLite, run this command from the same directory where 'kpi-collector run' was executed,
or use --artifact-dir to point to the artifact directory.`,
}

func init() {
	rootCmd.AddCommand(grafanaCmd)
}

// getGrafanaConfigDir returns the path to the grafana config directory
// <OutputDir>/grafana/
func getGrafanaConfigDir() string {
	return filepath.Join(database.OutputDir, "grafana")
}

// createGrafanaDirectories creates all necessary directories for grafana config
func createGrafanaDirectories(grafanaDir string) error {
	dirs := []string{
		grafanaDir,
		filepath.Join(grafanaDir, "datasources"),
		filepath.Join(grafanaDir, "dashboards"),
		filepath.Join(grafanaDir, "provisioning", "dashboards"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
