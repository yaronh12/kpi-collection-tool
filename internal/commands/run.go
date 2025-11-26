package commands

import (
	"fmt"
	"log"
	"time"

	"kpi-collector/internal/collector"
	"kpi-collector/internal/config"
	"kpi-collector/internal/grafana_ai"
	"kpi-collector/internal/kubernetes"
	"kpi-collector/internal/logger"

	"github.com/spf13/cobra"
)

const (
	defaultKPIsFilepath = "configs/kpis.json"
)

// Use the existing InputFlags struct directly!
var flags config.InputFlags

// runCmd represents the collect command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Collect KPI metrics from Prometheus/Thanos",
	Long: `Collect KPI metrics from Prometheus/Thanos endpoints and store them 
in a database (SQLite or PostgreSQL). Supports two authentication modes:
  1. Kubeconfig-based auto-discovery
  2. Manual bearer token and Thanos URL

The tool will continuously collect metrics at the specified frequency 
for the specified duration.`,
	Example: `  # Using kubeconfig (auto-discovery)
  kpi-collector collect --cluster-name prod --kubeconfig ~/.kube/config

  # Using manual credentials
  kpi-collector collect --cluster-name prod --token TOKEN --thanos-url thanos.example.com

  # With PostgreSQL backend
  kpi-collector collect --cluster-name prod --kubeconfig ~/.kube/config \
    --db-type postgres --postgres-url "postgresql://user:pass@localhost/kpi"

  # With Grafana AI analysis
  kpi-collector collect --cluster-name prod --kubeconfig ~/.kube/config \
    --summarize --grafana-file dashboard.json`,
	RunE: runCollect,
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Bind flags directly to config.InputFlags fields!

	// Authentication flags
	runCmd.Flags().StringVar(&flags.BearerToken, "token", "",
		"bearer token for Thanos authentication")
	runCmd.Flags().StringVar(&flags.ThanosURL, "thanos-url", "",
		"Thanos querier URL (without https://)")
	runCmd.Flags().StringVar(&flags.Kubeconfig, "kubeconfig", "",
		"path to kubeconfig file for auto-discovery")
	runCmd.Flags().StringVar(&flags.ClusterName, "cluster-name", "",
		"cluster name (required)")
	runCmd.Flags().BoolVar(&flags.InsecureTLS, "insecure-tls", false,
		"skip TLS certificate verification (development only)")

	// Sampling flags
	runCmd.Flags().IntVar(&flags.SamplingFreq, "frequency", 60,
		"sampling frequency in seconds")
	runCmd.Flags().DurationVar(&flags.Duration, "duration", 45*time.Minute,
		"total duration for sampling (e.g. 10s, 1m, 2h)")

	// Output flags
	runCmd.Flags().StringVar(&flags.OutputFile, "output", "kpi-output.json",
		"output file name for results")
	runCmd.Flags().StringVar(&flags.LogFile, "log", "kpi.log",
		"log file name")

	// Database flags
	runCmd.Flags().StringVar(&flags.DatabaseType, "db-type", "sqlite",
		"database type: sqlite or postgres")
	runCmd.Flags().StringVar(&flags.PostgresURL, "postgres-url", "",
		"PostgreSQL connection string (required if db-type=postgres)")

	// Grafana AI flags
	runCmd.Flags().StringVar(&flags.GrafanaFile, "grafana-file", "",
		"path to exported Grafana dashboard JSON to analyze")
	runCmd.Flags().BoolVar(&flags.Summarize, "summarize", false,
		"run Grafana AI summarization after KPI collection")
	runCmd.Flags().StringVar(&flags.AIModel, "ollama-model", "llama3.2:latest",
		"local Ollama model to use")
	runCmd.Flags().StringVar(&flags.KPIsFile, "kpis-file", defaultKPIsFilepath,
		"path to KPIs configuration file")

	// Mark required flags
	err := runCmd.MarkFlagRequired("cluster-name")
	if err != nil {
		panic(fmt.Sprintf("failed to mark cluster-name as required: %v", err))
	}

}

func runCollect(cmd *cobra.Command, args []string) error {
	fmt.Println("KPI Collector starting...")

	// Reuse existing validation logic!
	if err := config.ValidateFlags(flags); err != nil {
		return fmt.Errorf("invalid flags: %w", err)
	}

	fmt.Printf("Cluster: %s\n", flags.ClusterName)

	// Initialize logger
	logF, err := logger.InitLogger(flags.LogFile)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() {
		if err := logF.Close(); err != nil {
			fmt.Printf("Failed to close log file: %v\n", err)
		}
	}()

	log.Println("KPI Collector initialized.")

	// Load KPI queries
	kpis, err := config.LoadKPIs(flags.KPIsFile)
	if err != nil {
		return fmt.Errorf("failed to load KPI queries: %w", err)
	}

	// If kubeconfig is provided, discover Thanos URL and token
	if flags.Kubeconfig != "" {
		flags.ThanosURL, flags.BearerToken, err = kubernetes.SetupKubeconfigAuth(flags.Kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to setup kubeconfig auth: %w", err)
		}
		fmt.Printf("Discovered Thanos URL: %s\n", flags.ThanosURL)
		fmt.Printf("Created service account token!\n")
	}

	// Run collection
	collector.Run(kpis, flags)

	fmt.Println("All queries completed successfully!")

	// Run Grafana AI analysis if requested
	if flags.Summarize && flags.GrafanaFile != "" {
		log.Println("Starting Grafana AI Analysis...")
		if err := grafana_ai.Run(flags); err != nil {
			log.Printf("Grafana AI analysis failed: %v\n", err)
		} else {
			log.Println("Grafana AI analysis finished.")
		}
	}

	return nil
}
