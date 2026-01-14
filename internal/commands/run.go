package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/collector"
	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/config"
	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/kubernetes"
	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/logger"

	"github.com/spf13/cobra"
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
  kpi-collector collect --cluster-name prod --cluster-type ran --kubeconfig ~/.kube/config

  # Using manual credentials
  kpi-collector collect --cluster-name prod --cluster-type core --token TOKEN --thanos-url thanos.example.com

  # With PostgreSQL backend
  kpi-collector collect --cluster-name prod --cluster-type hub --kubeconfig ~/.kube/config \
    --db-type postgres --postgres-url "postgresql://user:pass@localhost/kpi`,
	RunE: runCollect,
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Authentication flags
	runCmd.Flags().StringVar(&flags.BearerToken, "token", "",
		"bearer token for Thanos authentication")
	runCmd.Flags().StringVar(&flags.ThanosURL, "thanos-url", "",
		"Thanos querier URL (without https://)")
	runCmd.Flags().StringVar(&flags.Kubeconfig, "kubeconfig", "",
		"path to kubeconfig file for auto-discovery")
	runCmd.Flags().StringVar(&flags.ClusterName, "cluster-name", "",
		"cluster name (required)")
	runCmd.Flags().StringVar(&flags.ClusterType, "cluster-type", "",
		"cluster type for categorization: ran, core, or hub")
	runCmd.Flags().BoolVar(&flags.InsecureTLS, "insecure-tls", false,
		"skip TLS certificate verification (development only)")

	// Sampling flags
	runCmd.Flags().DurationVar(&flags.SamplingFreq, "frequency", 60*time.Second,
		"sampling frequency (e.g. 30s, 1m, 2h)")
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

	runCmd.Flags().StringVar(&flags.KPIsFile, "kpis-file", "",
		"path to KPIs configuration file (required)")

	// Mark required flags
	if err := runCmd.MarkFlagRequired("cluster-name"); err != nil {
		panic(fmt.Sprintf("failed to mark cluster-name as required: %v", err))
	}
	if err := runCmd.MarkFlagRequired("kpis-file"); err != nil {
		panic(fmt.Sprintf("failed to mark kpis-file as required: %v", err))
	}

}

func runCollect(cmd *cobra.Command, args []string) error {
	fmt.Println("KPI Collector starting...")

	// Validate all flags (including cluster type)
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

	kpis, err = substituteCPUsIfNeeded(kpis, flags)
	if err != nil {
		return err
	}

	// Warn if any KPI frequency exceeds the duration
	warnFrequencyExceedsDuration(kpis, flags)

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

	return nil
}

// substituteCPUsIfNeeded checks if queries contain CPU placeholders and if so,
// fetches CPU IDs from PerformanceProfiles and substitutes them into queries
func substituteCPUsIfNeeded(kpis config.KPIs, flags config.InputFlags) (config.KPIs, error) {
	if !config.RequiresCPUSubstitution(kpis) {
		return kpis, nil
	}

	if flags.Kubeconfig == "" {
		return kpis, fmt.Errorf("queries contain CPU placeholders ({{RESERVED_CPUS}}/{{ISOLATED_CPUS}}) but no --kubeconfig provided")
	}

	reservedCPUs, isolatedCPUs, err := kubernetes.FetchCPUsFromPerformanceProfiles(flags.Kubeconfig)
	if err != nil {
		return kpis, fmt.Errorf("failed to fetch CPUs from PerformanceProfiles: %w", err)
	}

	fmt.Printf("Loaded CPU sets - Reserved: [%s], Isolated: [%s]\n", reservedCPUs, isolatedCPUs)

	cpuPlaceholders := &config.CPUPlaceholders{
		Reserved: reservedCPUs,
		Isolated: isolatedCPUs,
	}

	return config.SubstituteCPUPlaceholders(kpis, cpuPlaceholders), nil
}

// warnFrequencyExceedsDuration prints a warning if any KPI's sampling frequency
// is longer than the total duration, meaning only one sample will be collected
func warnFrequencyExceedsDuration(kpis config.KPIs, flags config.InputFlags) {
	for _, kpi := range kpis.Queries {
		effectiveFreq := kpi.GetEffectiveFrequency(flags.SamplingFreq)

		if effectiveFreq > flags.Duration {
			fmt.Printf("WARNING: KPI '%s' has frequency %s which exceeds duration %s. Only 1 sample will be collected.\n",
				kpi.ID, effectiveFreq, flags.Duration)
		}
	}

	// Also warn about the default frequency if no custom frequencies are set
	if flags.SamplingFreq > flags.Duration {
		fmt.Printf("WARNING: Default sampling frequency %s exceeds duration %s. KPIs without custom frequency will only collect 1 sample.\n",
			flags.SamplingFreq, flags.Duration)
	}
}
