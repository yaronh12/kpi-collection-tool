package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	fmt.Println("RDS KPI Collector starting...")

	// Setup flags
	flags, err := setupFlags()
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
		flags.ThanosURL, flags.BearerToken, err = setupKubeconfigAuth(flags.Kubeconfig)
		if err != nil {
			fmt.Printf("Failed to setup kubeconfig auth: %v\n", err)
			return
		}
		fmt.Printf("Discovered Thanos URL: %s\n", flags.ThanosURL)
		fmt.Printf("Created service account token!\n")
	}

	// Run queries
	err = runQueries(kpis, flags)
	if err != nil {
		fmt.Printf("Failed to run queries: %v\n", err)
		return
	}

	fmt.Println("All queries completed successfully!")
}

// loadKPIs loads Prometheus queries from kpis.json file
func loadKPIs() (KPIs, error) {
	kpisFile, err := os.Open("kpis.json")
	if err != nil {
		return KPIs{}, fmt.Errorf("failed to open kpis.json: %v", err)
	}
	defer func() {
		if closeErr := kpisFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close kpis.json: %v\n", closeErr)
		}
	}()

	var kpis KPIs
	decoder := json.NewDecoder(kpisFile)
	if err := decoder.Decode(&kpis); err != nil {
		return KPIs{}, fmt.Errorf("failed to decode kpis.json: %v", err)
	}

	return kpis, nil
}
