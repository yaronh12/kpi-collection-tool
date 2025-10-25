package main

import (
	"encoding/json"
	"fmt"
	"os"
	"log"

	"rds-kpi-collector/internal/config"
	"rds-kpi-collector/internal/kubernetes"
	"rds-kpi-collector/internal/prometheus"
	"rds-kpi-collector/internal/logger"
	"rds-kpi-collector/internal/collector"
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


	// Initialize logger
	logF, err := logger.InitLogger(flags.LogFile)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := logF.Close(); err != nil {
			fmt.Printf("Failed to close log file: %v\n", err)
		}
	}()

	log.Println("RDS KPI Collector initialized.")

	// Load KPI queries
	kpis, err := loadKPIs()
	if err != nil {
		log.Printf("Failed to load KPI queries: %v\n", err)
		fmt.Printf("Failed to load KPI queries: %v\n", err)
		return
	}

	// If kubeconfig is provided, discover Thanos URL and token
	if flags.Kubeconfig != "" {
		flags.ThanosURL, flags.BearerToken, err = kubernetes.SetupKubeconfigAuth(flags.Kubeconfig)
		if err != nil {
			log.Printf("Failed to setup kubeconfig auth: %v\n", err)
			fmt.Printf("Failed to setup kubeconfig auth: %v\n", err)
			return
		}
		fmt.Printf("Discovered Thanos URL: %s\n", flags.ThanosURL)
		fmt.Printf("Created service account token!\n")
	}

	// Run queries
	err = prometheus.RunQueries(kpis, flags)
	if err != nil {
		log.Printf("Failed to run queries: %v\n", err)
		fmt.Printf("Failed to run queries: %v\n", err)
		return
	}

	fmt.Println("All queries completed successfully!")

	// Run collector
	if err := collector.RunKPICollector(flags.SamplingFreq, flags.Duration, flags.OutputFile); err != nil {
		log.Printf("Collector error: %v\n", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	log.Println("RDS KPI Collector finished successfully.")
	fmt.Println("RDS KPI Collector finished successfully.")

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
