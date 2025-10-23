package main

import (
	"encoding/json"
	"fmt"
	"os"
	"rds-kpi-collector/internal/config"
	"rds-kpi-collector/internal/kubernetes"
	"rds-kpi-collector/internal/prometheus"
)

func main() {
	fmt.Println("RDS KPI Collector starting...")

	// Setup flags
	flags, err := config.SetupFlags()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Cluster: %s\n", flags.ClusterName)

	// Load KPI queries
	kpis, err := loadKPIs()
	if err != nil {
		fmt.Printf("Failed to load KPI queries: %v\n", err)
		return
	}

	// If kubeconfig is provided, discover Thanos URL and token
	if flags.Kubeconfig != "" {
		flags.ThanosURL, flags.BearerToken, err = kubernetes.SetupKubeconfigAuth(flags.Kubeconfig)
		if err != nil {
			fmt.Printf("Failed to setup kubeconfig auth: %v\n", err)
			return
		}
		fmt.Printf("Discovered Thanos URL: %s\n", flags.ThanosURL)
		fmt.Printf("Created service account token!\n")
	}

	// Run queries
	err = prometheus.RunQueries(kpis, flags)
	if err != nil {
		fmt.Printf("Failed to run queries: %v\n", err)
		return
	}

	fmt.Println("All queries completed successfully!")
}

// loadKPIs loads Prometheus queries from kpis.json file
func loadKPIs() (config.KPIs, error) {
	kpisFile, err := os.Open("configs/kpis.json")
	if err != nil {
		return config.KPIs{}, fmt.Errorf("failed to open kpis.json: %v", err)
	}
	defer func() {
		if closeErr := kpisFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close kpis.json: %v\n", closeErr)
		}
	}()

	var kpis config.KPIs
	decoder := json.NewDecoder(kpisFile)
	if err := decoder.Decode(&kpis); err != nil {
		return config.KPIs{}, fmt.Errorf("failed to decode kpis.json: %v", err)
	}

	return kpis, nil
}
