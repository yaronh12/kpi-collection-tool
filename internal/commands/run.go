package commands

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/collector"
	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/config"
	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/database"
	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/kubernetes"
	"github.com/redhat-best-practices-for-k8s/kpi-collection-tool/internal/logger"

	"github.com/spf13/cobra"
)

const defaultLogFile = "kpi.log"

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
for the specified duration.

For more usage options, see the docs/ directory.`,
	Example: `  # Using kubeconfig (auto-discovery of Thanos URL and token)
  kpi-collector run --cluster-name prod --cluster-type ran \
    --kubeconfig ~/.kube/config --kpis-file kpis.json

  # Using manual credentials
  kpi-collector run --cluster-name prod --cluster-type core \
    --token $TOKEN --thanos-url thanos.example.com --kpis-file kpis.json

  # Collect all KPIs once and exit
  kpi-collector run --cluster-name prod --cluster-type ran \
    --kubeconfig ~/.kube/config --kpis-file kpis.json --once

  # Custom sampling: every 30s for 2 hours
  kpi-collector run --cluster-name prod --cluster-type ran \
    --kubeconfig ~/.kube/config --kpis-file kpis.json \
    --frequency 30s --duration 2h

  # With PostgreSQL backend
  kpi-collector run --cluster-name prod --cluster-type hub \
    --kubeconfig ~/.kube/config --kpis-file kpis.json \
    --db-type postgres --postgres-url "postgresql://user:pass@localhost:5432/kpi"`,
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

	// Database flags
	runCmd.Flags().StringVar(&flags.DatabaseType, "db-type", "sqlite",
		"database type: sqlite or postgres")
	runCmd.Flags().StringVar(&flags.PostgresURL, "postgres-url", "",
		"PostgreSQL connection string (required if db-type=postgres)")

	runCmd.Flags().StringVar(&flags.KPIsFile, "kpis-file", "",
		"path to KPIs configuration file (required)")

	// Single-run mode
	runCmd.Flags().BoolVar(&flags.SingleRun, "once", false,
		"collect all KPI metrics once and exit (ignores --frequency and --duration)")

	// Mark required flags
	if err := runCmd.MarkFlagRequired("cluster-name"); err != nil {
		panic(fmt.Sprintf("failed to mark cluster-name as required: %v", err))
	}
	if err := runCmd.MarkFlagRequired("kpis-file"); err != nil {
		panic(fmt.Sprintf("failed to mark kpis-file as required: %v", err))
	}

	// --once is mutually exclusive with --frequency and --duration
	runCmd.MarkFlagsMutuallyExclusive("once", "frequency")
	runCmd.MarkFlagsMutuallyExclusive("once", "duration")

}

func runCollect(cmd *cobra.Command, args []string) error {
	fmt.Println("KPI Collector starting...")

	// Validate all flags (including cluster type)
	if err := config.ValidateFlags(flags); err != nil {
		return fmt.Errorf("invalid flags: %w", err)
	}

	fmt.Printf("Cluster: %s (%s)\n", flags.ClusterName, flags.ClusterType)

	// Initialize logger
	logF, err := logger.InitLogger(defaultLogFile)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() {
		if err := logF.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close log file: %v\n", err)
		}
	}()
	fmt.Printf("Log file: %s\n", defaultLogFile)
	fmt.Printf("Database: %s\n", databaseLocation(flags))

	log.Println("KPI Collector initialized")

	// Load KPI queries
	kpis, err := config.LoadKPIs(flags.KPIsFile)
	if err != nil {
		return fmt.Errorf("failed to load KPI queries: %w", err)
	}
	log.Printf("Loaded KPIs from %s", flags.KPIsFile)

	// Validate KPI configurations (syntax, duplicates, etc.)
	if validationErrors := config.ValidateKPIs(kpis); len(validationErrors) > 0 {
		fmt.Println("KPI validation errors:")
		for _, e := range validationErrors {
			fmt.Printf("  ✗ %v\n", e)
		}
		return fmt.Errorf("found %d KPI validation error(s)", len(validationErrors))
	}
	fmt.Printf("✓ Validated %d KPI(s)\n", len(kpis.Queries))

	kpis, err = substituteCPUsIfNeeded(kpis, flags)
	if err != nil {
		return err
	}

	if !flags.SingleRun {
		warnFrequencyExceedsDuration(kpis, flags)
	}

	// Validate range query frequency/range mismatches
	if err := validateRangeFrequency(kpis, flags); err != nil {
		return err
	}

	// If kubeconfig is provided, discover Thanos URL and token
	if flags.Kubeconfig != "" {
		log.Printf("Using kubeconfig authentication: %s", flags.Kubeconfig)
		flags.ThanosURL, flags.BearerToken, err = kubernetes.SetupKubeconfigAuth(flags.Kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to setup kubeconfig auth: %w", err)
		}
		fmt.Printf("Discovered Thanos URL: %s\n", flags.ThanosURL)
		fmt.Println("Created service account token")
	} else {
		log.Printf("Using manual token authentication for %s", flags.ThanosURL)
	}

	// Run collection
	if flags.SingleRun {
		collector.RunOnce(kpis, flags)
	} else {
		collector.Run(kpis, flags)
	}

	return nil
}

func databaseLocation(flags config.InputFlags) string {
	if flags.DatabaseType == "postgres" {
		return "postgres (external)"
	}
	return fmt.Sprintf("sqlite (%s)", database.GetSQLiteDBPath())
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

// validateRangeFrequency checks range queries for frequency/range mismatches.
// Returns an error if frequency exceeds range (data gaps), and prints a warning for heavy overlap.
func validateRangeFrequency(kpis config.KPIs, flags config.InputFlags) error {
	for _, kpi := range kpis.Queries {
		if kpi.GetEffectiveQueryType() != "range" || kpi.Range == nil {
			continue
		}

		freq := kpi.GetEffectiveFrequency(flags.SamplingFreq)
		queryRange := kpi.Range.Duration

		if freq > queryRange {
			return fmt.Errorf("KPI '%s' has frequency %s > range %s — this creates gaps where no data is collected",
				kpi.ID, freq, queryRange)
		}

		if freq < queryRange/2 {
			overlapPercent := 100 - (100*freq)/queryRange
			fmt.Printf("WARNING: KPI '%s' has frequency %s with range %s — ~%d%% of each query overlaps the previous one.\n",
				kpi.ID, freq, queryRange, overlapPercent)
		}
	}

	return nil
}

// warnFrequencyExceedsDuration prints a warning if any KPI's sampling frequency
// is longer than the total duration, meaning only one sample will be collected
func warnFrequencyExceedsDuration(kpis config.KPIs, flags config.InputFlags) {
	for _, kpi := range kpis.Queries {
		if kpi.IsRunOnce() {
			continue
		}

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
