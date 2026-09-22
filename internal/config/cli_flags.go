package config

import (
	"fmt"
)

// ValidateFlags ensures the correct combination of flags is provided.
// Prometheus-specific fields (frequency, duration, db-type, postgres-url)
// are validated separately by ValidatePromSettings after YAML values have
// been merged into the flags.
func ValidateFlags(flags InputFlags) error {
	if flags.ClusterName == "" {
		return fmt.Errorf("cluster name is required: use --cluster-name flag")
	}

	validClusterTypes := map[string]bool{"ran": true, "core": true, "hub": true}
	if flags.ClusterType == "" {
		return fmt.Errorf("cluster-type is required: must be 'ran', 'core', or 'hub'")
	}
	if !validClusterTypes[flags.ClusterType] {
		return fmt.Errorf("invalid cluster-type '%s': must be 'ran', 'core', or 'hub'", flags.ClusterType)
	}

	if flags.InsecureTLS {
		fmt.Println("WARNING: TLS certificate verification is disabled. Use only in development environments.")
	}

	validAuthCombo := (flags.BearerToken != "" && flags.ThanosURL != "" && flags.Kubeconfig == "") ||
		(flags.BearerToken == "" && flags.ThanosURL == "" && flags.Kubeconfig != "")

	if !validAuthCombo {
		return fmt.Errorf("invalid flag combination: either provide --token and --thanos-url, or provide --kubeconfig")
	}

	if flags.TasksConfig != "" && flags.PromKPIsConfig != "" {
		return fmt.Errorf("--tasks and --prom-kpis-config are mutually exclusive")
	}

	if !flags.HasAnyTaskConfig() {
		return fmt.Errorf("either --tasks or --prom-kpis-config is required")
	}

	return nil
}

// ValidatePromSettings checks the Prometheus-specific fields in InputFlags.
// Called after YAML values (from tasks.yaml or KPI YAML) have been applied.
func ValidatePromSettings(flags InputFlags) error {
	if flags.SamplingFreq <= 0 {
		return fmt.Errorf("sampling frequency must be greater than 0")
	}

	if flags.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}

	if flags.DatabaseType != "sqlite" && flags.DatabaseType != "postgres" {
		return fmt.Errorf("invalid db-type: must be 'sqlite' or 'postgres'")
	}

	if flags.DatabaseType == "postgres" && flags.PostgresURL == "" {
		return fmt.Errorf("postgres-url is required when db-type=postgres")
	}

	return nil
}

// ApplyPromTaskConfig copies Prometheus settings from a tasks.yaml prometheus
// section into InputFlags. Used with --tasks where CLI prom flags are rejected
// and the YAML is the sole source.
func ApplyPromTaskConfig(flags *InputFlags, p *PrometheusTaskConfig) {
	if p == nil {
		return
	}
	if p.Frequency != nil {
		flags.SamplingFreq = p.Frequency.Duration
	}
	if p.Duration != nil {
		flags.Duration = p.Duration.Duration
	}
	if p.DBType != "" {
		flags.DatabaseType = p.DBType
	}
	if p.PostgresURL != "" {
		flags.PostgresURL = p.PostgresURL
	}
	if p.Once != nil {
		flags.SingleRun = *p.Once
	}
}

// ApplyKPIsDefaults copies Prometheus settings from a KPI YAML file into
// InputFlags, but only for fields not already set by CLI (indicated by
// changedFlags). Used with --prom-kpis-config where CLI flags override YAML.
func ApplyKPIsDefaults(flags *InputFlags, kpis KPIs, changedFlags map[string]bool) {
	if kpis.Frequency != nil && !changedFlags["frequency"] {
		flags.SamplingFreq = kpis.Frequency.Duration
	}
	if kpis.Duration != nil && !changedFlags["duration"] {
		flags.Duration = kpis.Duration.Duration
	}
	if kpis.DBType != "" && !changedFlags["db-type"] {
		flags.DatabaseType = kpis.DBType
	}
	if kpis.PostgresURL != "" && !changedFlags["postgres-url"] {
		flags.PostgresURL = kpis.PostgresURL
	}
	if kpis.Once != nil && !changedFlags["once"] {
		flags.SingleRun = *kpis.Once
	}
}
