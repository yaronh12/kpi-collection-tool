package config

import (
	"fmt"
)

// validateFlags ensures the correct combination of flags is provided
func ValidateFlags(flags InputFlags) error {
	if flags.ClusterName == "" {
		return fmt.Errorf("cluster name is required: use --cluster-name flag")
	}

	if flags.ClusterType != "" {
		validTypes := map[string]bool{"ran": true, "core": true, "hub": true}
		if !validTypes[flags.ClusterType] {
			return fmt.Errorf("invalid cluster-type: must be 'ran', 'core', or 'hub'")
		}
	}

	if flags.InsecureTLS {
		fmt.Println("WARNING: TLS certificate verification is disabled. Use only in development environments.")
	}

	// Validate flag combinations for authentication
	validAuthCombo := (flags.BearerToken != "" && flags.ThanosURL != "" && flags.Kubeconfig == "") ||
		(flags.BearerToken == "" && flags.ThanosURL == "" && flags.Kubeconfig != "")

	if !validAuthCombo {
		return fmt.Errorf("invalid flag combination: either provide --token and --thanos-url, or provide --kubeconfig")
	}

	if flags.SamplingFreq <= 0 {
		return fmt.Errorf("sampling frequency must be greater than 0")
	}

	if flags.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}

	if flags.OutputFile == "" {
		return fmt.Errorf("output file must be specified")
	}

	if flags.LogFile == "" {
		return fmt.Errorf("log file must be specified")
	}

	if flags.DatabaseType != "sqlite" && flags.DatabaseType != "postgres" {
		return fmt.Errorf("invalid db-type: must be 'sqlite' or 'postgres'")
	}

	if flags.DatabaseType == "postgres" && flags.PostgresURL == "" {
		return fmt.Errorf("postgres-url is required when db-type=postgres")
	}

	if flags.KPIsFile == "" {
		return fmt.Errorf("kpis-file must be specified")
	}

	return nil
}
