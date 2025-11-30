package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var grafanaFlags struct {
	datasource  string
	postgresURL string
	port        int
}

var grafanaCmd = &cobra.Command{
	Use:   "grafana",
	Short: "Launch Grafana dashboard with KPI data",
	Long: `Launch a local Grafana instance with the KPI dashboard pre-configured.
Supports both SQLite and PostgreSQL datasources.`,
	Example: `  # Using SQLite
  kpi-collector grafana --datasource=sqlite

  # Using PostgreSQL
  kpi-collector grafana --datasource=postgres --postgres-url "postgresql://user:pass@host:5432/dbname"

  # Custom port
  kpi-collector grafana --datasource=sqlite --port 3001`,
	RunE: runGrafana,
}

func init() {
	rootCmd.AddCommand(grafanaCmd)

	grafanaCmd.Flags().StringVar(&grafanaFlags.datasource, "datasource", "",
		"datasource type: sqlite or postgres (required)")
	grafanaCmd.Flags().StringVar(&grafanaFlags.postgresURL, "postgres-url", "",
		"PostgreSQL connection string (required if datasource=postgres)")
	grafanaCmd.Flags().IntVar(&grafanaFlags.port, "port", 3000,
		"Grafana port (default: 3000)")

	if err := grafanaCmd.MarkFlagRequired("datasource"); err != nil {
		panic(fmt.Sprintf("failed to mark datasource as required: %v", err))
	}
}

func runGrafana(cmd *cobra.Command, args []string) error {
	if grafanaFlags.datasource != "sqlite" && grafanaFlags.datasource != "postgres" {
		return fmt.Errorf("datasource must be 'sqlite' or 'postgres', got: %s", grafanaFlags.datasource)
	}

	if grafanaFlags.datasource == "postgres" && grafanaFlags.postgresURL == "" {
		return fmt.Errorf("--postgres-url is required when datasource is 'postgres'")
	}

	fmt.Printf("Launching Grafana with %s datasource...\n", grafanaFlags.datasource)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create datasource config
	if err := createDatasourceConfig(cwd); err != nil {
		return fmt.Errorf("failed to create datasource config: %w", err)
	}

	// Call make install-grafana with appropriate variables
	if err := runMakeInstallGrafana(); err != nil {
		return fmt.Errorf("failed to launch Grafana: %w", err)
	}

	fmt.Println("✅ Grafana is starting up...")
	fmt.Printf("🌐 Open http://localhost:%d in your browser\n", grafanaFlags.port)
	fmt.Println("👤 Default login: admin/admin")
	fmt.Println("\n💡 To stop Grafana, run: docker stop grafana-kpi")

	return nil
}

func createDatasourceConfig(workDir string) error {
	datasourceDir := filepath.Join(workDir, "grafana", "datasource")

	if err := os.MkdirAll(datasourceDir, 0755); err != nil {
		return fmt.Errorf("failed to create datasource directory: %w", err)
	}

	var configContent string

	if grafanaFlags.datasource == "sqlite" {
		configContent = "apiVersion: 1\n\n" +
			"datasources:\n" +
			"  - name: KPI-SQLite\n" +
			"    type: frser-sqlite-datasource\n" +
			"    uid: kpi-datasource\n" +
			"    access: proxy\n" +
			"    jsonData:\n" +
			"      path: /var/lib/grafana/kpi_metrics.db\n" +
			"    isDefault: true\n" +
			"    editable: true\n"
	} else {
		// Parse PostgreSQL URL
		// Expected format: postgresql://user:password@host:port/database
		pgURL := grafanaFlags.postgresURL
		pgURL = strings.TrimPrefix(pgURL, "postgresql://")
		pgURL = strings.TrimPrefix(pgURL, "postgres://")

		var user, password, host, port, database string

		// Extract user and password
		atIndex := strings.Index(pgURL, "@")
		if atIndex > 0 {
			userPass := pgURL[:atIndex]
			pgURL = pgURL[atIndex+1:]

			colonIndex := strings.Index(userPass, ":")
			if colonIndex > 0 {
				user = userPass[:colonIndex]
				password = userPass[colonIndex+1:]
			} else {
				user = userPass
			}
		}

		// Extract database
		slashIndex := strings.Index(pgURL, "/")
		if slashIndex > 0 {
			database = pgURL[slashIndex+1:]
			// Remove query parameters if any
			if qIndex := strings.Index(database, "?"); qIndex > 0 {
				database = database[:qIndex]
			}
			pgURL = pgURL[:slashIndex]
		}

		// Extract host and port
		colonIndex := strings.LastIndex(pgURL, ":")
		if colonIndex > 0 {
			host = pgURL[:colonIndex]
			port = pgURL[colonIndex+1:]
		} else {
			host = pgURL
			port = "5432"
		}

		configContent = "apiVersion: 1\n\n" +
			"datasources:\n" +
			"  - name: KPI-PostgreSQL\n" +
			"    type: postgres\n" +
			"    uid: kpi-datasource\n" +
			"    access: proxy\n" +
			"    url: " + host + ":" + port + "\n" +
			"    database: " + database + "\n" +
			"    user: " + user + "\n"

		if password != "" {
			configContent += "    secureJsonData:\n" +
				"      password: " + password + "\n"
		}

		configContent += "    isDefault: true\n" +
			"    editable: true\n" +
			"    jsonData:\n" +
			"      sslmode: disable\n" +
			"      maxOpenConns: 10\n" +
			"      maxIdleConns: 10\n" +
			"      connMaxLifetime: 14400\n"
	}

	configPath := filepath.Join(datasourceDir, "datasource.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write datasource config: %w", err)
	}

	return nil
}

func runMakeInstallGrafana() error {
	// Determine which dashboard file to use
	var dashboardFile string
	if grafanaFlags.datasource == "sqlite" {
		dashboardFile = "sqlite-dashboard.json"
	} else {
		dashboardFile = "postgres-dashboard.json"
	}

	makeCmd := exec.Command("make", "install-grafana",
		fmt.Sprintf("DB_TYPE=%s", grafanaFlags.datasource),
		fmt.Sprintf("GRAFANA_PORT=%d", grafanaFlags.port),
		fmt.Sprintf("DASHBOARD_FILE=%s", dashboardFile),
	)

	// Pass postgres URL if needed (though not used by Makefile currently)
	if grafanaFlags.datasource == "postgres" {
		makeCmd.Env = append(os.Environ(), fmt.Sprintf("POSTGRES_URL=%s", grafanaFlags.postgresURL))
	}

	output, err := makeCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("make install-grafana failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
