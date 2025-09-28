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
	queries, err := loadKPIQueries()
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
	err = runQueries(queries, flags.ThanosURL, flags.BearerToken, flags.ClusterName)
	if err != nil {
		fmt.Printf("Failed to run queries: %v\n", err)
		return
	}

	fmt.Println("All queries completed successfully!")
}

// loadKPIQueries loads Prometheus queries from kpis.json file
func loadKPIQueries() ([]string, error) {
	kpisFile, err := os.Open("kpis.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open kpis.json: %v", err)
	}
	defer kpisFile.Close()

	var kpis KPIs
	decoder := json.NewDecoder(kpisFile)
	if err := decoder.Decode(&kpis); err != nil {
		return nil, fmt.Errorf("failed to decode kpis.json: %v", err)
	}

	var queries []string
	for _, query := range kpis.Queries {
		queries = append(queries, query.PromQuery)
	}

	return queries, nil
}
