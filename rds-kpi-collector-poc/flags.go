package main

import (
	"flag"
	"fmt"
)

// SetupFlags parses and validates command line flags, returns InputFlags struct
func SetupFlags() (InputFlags, error) {
	var flags InputFlags

	flag.StringVar(&flags.BearerToken, "token", "", "bearer token for thanos-queries")
	flag.StringVar(&flags.ThanosURL, "thanos-url", "", "thanos url for http requests")
	flag.StringVar(&flags.Kubeconfig, "kubeconfig", "", "kubeconfig file path")
	flag.StringVar(&flags.ClusterName, "cluster-name", "", "cluster name (required)")
	flag.BoolVar(&flags.InsecureTLS, "insecure-tls", false, "skip TLS certificate verification")

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

	if (flags.BearerToken != "" && flags.ThanosURL != "" && flags.Kubeconfig == "") ||
		(flags.BearerToken == "" && flags.ThanosURL == "" && flags.Kubeconfig != "") {
		return nil
	} else {
		return fmt.Errorf("invalid flag combination: either provide --token and --thanos-url, or provide --kubeconfig")
	}
}
