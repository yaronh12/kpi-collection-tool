package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"rds-kpi-collector/internal/config"
	"rds-kpi-collector/internal/kubernetes"
	"rds-kpi-collector/internal/logger"
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
	kpis, err := loadKPIs(flags.KPIsFile)
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

	// Run the collection loop
	runCollectionLoop(kpis, flags)

	fmt.Println("All queries completed successfully!")

}

// runCollectionLoop runs the KPI collection with timer and ticker
func runCollectionLoop(kpis config.KPIs, flags config.InputFlags) {
	durationTimer := time.NewTimer(flags.Duration)
	defer durationTimer.Stop()

	sampleTicker := time.NewTicker(time.Duration(flags.SamplingFreq) * time.Second)
	defer sampleTicker.Stop()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	log.Printf("Running for %s, deadline time: %s\n",
		flags.Duration.String(),
		time.Now().Add(flags.Duration).Format(time.RFC3339))

	// Run the first sample immediately
	runSample(kpis, flags, 1)
	sampleCount := 2

	// Then continue with the ticker-based loop
	for {
		select {
		case <-sampleTicker.C:
			runSample(kpis, flags, sampleCount)
			sampleCount++

		case <-durationTimer.C:
			log.Printf("Duration timer expired after %d samples", sampleCount)
			return

		case <-interruptChan:
			log.Printf("Program interrupted after %d samples", sampleCount)
			return
		}
	}
}

// runSample executes a single KPI collection sample
func runSample(kpis config.KPIs, flags config.InputFlags, sampleCount int) {
	log.Printf("Running sample %d", sampleCount)

	if err := prometheus.RunQueries(kpis, flags); err != nil {
		log.Printf("RunQueries failed: %v\n", err)
	} else {
		log.Printf("Sample %d completed successfully", sampleCount)
	}
}

// loadKPIs loads Prometheus queries from kpis.json file
func loadKPIs(kpisFilePath string) (config.KPIs, error) {
	kpisFile, err := os.Open(kpisFilePath)
	if err != nil {
		return config.KPIs{}, fmt.Errorf("failed to open kpis file %s: %v", kpisFilePath, err)
	}
	defer func() {
		if closeErr := kpisFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close kpis.json: %v\n", closeErr)
		}
	}()

	var kpis config.KPIs
	decoder := json.NewDecoder(kpisFile)
	if err := decoder.Decode(&kpis); err != nil {
		return config.KPIs{}, fmt.Errorf("failed to decode kpis file %s: %v", kpisFilePath, err)
	}

	return kpis, nil
}
