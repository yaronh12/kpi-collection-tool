package main

import (
	"fmt"
	"log"
	"os"

	"rds-kpi-collector/internal/collector"
	"rds-kpi-collector/internal/config"
	"rds-kpi-collector/internal/kubernetes"
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
	kpis, err := config.LoadKPIs("configs/kpis.json")
	if err != nil {
		log.Printf("Failed to load KPI queries: %v\n", err)
		return
	}

	// If kubeconfig is provided, discover Thanos URL and token
	if flags.Kubeconfig != "" {
		flags.ThanosURL, flags.BearerToken, err = kubernetes.SetupKubeconfigAuth(flags.Kubeconfig)
		if err != nil {
			log.Printf("Failed to setup kubeconfig auth: %v\n", err)
			return
		}
		fmt.Printf("Discovered Thanos URL: %s\n", flags.ThanosURL)
		fmt.Printf("Created service account token!\n")
	}

	// Run collection
	collector.Run(kpis, flags)

	fmt.Println("All queries completed successfully!")

}
