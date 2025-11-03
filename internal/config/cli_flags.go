package config

import (
	"flag"
	"fmt"
	"time"
)

// setupFlags parses and validates command line flags, returns InputFlags struct
func SetupFlags() (InputFlags, error) {
	var flags InputFlags

	flag.StringVar(&flags.BearerToken, "token", "", "bearer token for thanos-queries")
	flag.StringVar(&flags.ThanosURL, "thanos-url", "", "thanos url for http requests")
	flag.StringVar(&flags.Kubeconfig, "kubeconfig", "", "kubeconfig file path")
	flag.StringVar(&flags.ClusterName, "cluster-name", "", "cluster name (required)")
	flag.BoolVar(&flags.InsecureTLS, "insecure-tls", false, "skip TLS certificate verification")

	flag.IntVar(&flags.SamplingFreq, "frequency", 60, "sampling frequency in seconds")
	flag.DurationVar(&flags.Duration, "duration", 45*time.Minute, "total duration for sampling (e.g. 10s, 1m, 2h)")
	flag.StringVar(&flags.OutputFile, "output", "kpi-output.json", "output file name for results")
	flag.StringVar(&flags.LogFile, "log", "kpi.log", "log file name")
	flag.StringVar(&flags.DatabaseType, "db-type", "sqlite", "database type: sqlite or postgres (default: sqlite)")
	flag.StringVar(&flags.PostgresURL, "postgres-url", "", "PostgreSQL connection string (required if db-type=postgres)")
	flag.Parse()

	err := validateFlags(flags)
	return flags, err
}

// validateFlags ensures the correct combination of flags is provided
func validateFlags(flags InputFlags) error {
	if flags.ClusterName == "" {
		return fmt.Errorf("cluster name is required: use --cluster-name flag")
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

	return nil
}
