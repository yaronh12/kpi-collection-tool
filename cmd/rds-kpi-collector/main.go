package main

import (
	"encoding/json"
	"fmt"
	"os"
	"log"
	"time"

	"rds-kpi-collector/internal/config"
	"rds-kpi-collector/internal/kubernetes"
	"rds-kpi-collector/internal/prometheus"
	"rds-kpi-collector/internal/logger"
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

	// Calculate number of runs based on sampling frequency and duration
	numRuns := int(flags.Duration.Seconds()) / flags.SamplingFreq

	for i := 1; i <= numRuns; i++ {
		log.Printf("Running sample %d/%d\n", i, numRuns)
		fmt.Printf("Running sample %d/%d\n", i, numRuns)

		// Run Prometheus queries
		if err := prometheus.RunQueries(kpis, flags); err != nil {
			log.Printf("RunQueries failed on sample %d: %v\n", i, err)
			fmt.Printf("RunQueries failed on sample %d: %v\n", i, err)
		} else {
			log.Printf("Sample %d completed successfully\n", i)
			fmt.Printf("Sample %d completed successfully\n", i)
		}

		// Sleep between samples
		time.Sleep(time.Duration(flags.SamplingFreq) * time.Second)
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
