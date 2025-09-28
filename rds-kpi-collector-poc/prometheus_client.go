package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func (t *tokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.rt.RoundTrip(req)
}

// setupPromClient creates and configures a Prometheus API client
func setupPromClient(thanosURL, bearerToken string) (v1.API, error) {
	client, err := api.NewClient(api.Config{
		Address: "https://" + thanosURL,
		RoundTripper: &tokenRoundTripper{
			token: bearerToken,
			rt: &http.Transport{
				// NOTE: InsecureSkipVerify is set to true for development purposes only.
				// In production environments, this should be false and proper certificate
				// validation should be implemented.
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client: %v", err)
	}

	v1api := v1.NewAPI(client)
	return v1api, nil
}

// runQueries executes all Prometheus queries and stores results in database
func runQueries(queriesToRun []string, thanosURL string, bearerToken string, clusterName string) error {
	// Initialize Database
	db, err := initDB()
	if err != nil {
		return fmt.Errorf("failed to init database: %v", err)
	}
	defer db.Close()

	// Get or create cluster in DB
	clusterID, err := getOrCreateCluster(db, clusterName)
	if err != nil {
		return fmt.Errorf("failed to get cluster ID: %v", err)
	}

	// Create Prometheus client
	v1api, err := setupPromClient(thanosURL, bearerToken)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	// Execute queries
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, query := range queriesToRun {
		err := executeQuery(ctx, v1api, db, clusterID, query)
		if err != nil {
			fmt.Printf("Query execution failed: %v\n", err)
			// Continue with next query even if one fails
		}
	}

	return nil
}

// executeQuery executes a single Prometheus query and handles the result
func executeQuery(ctx context.Context, v1api v1.API, db *sql.DB, clusterID int64, query string) error {
	fmt.Println("------------------------")
	fmt.Printf("Running: %s\n", query)

	// Execute query using the Prometheus client library
	result, warnings, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		fmt.Println("query failed: ", err)
		if storeErr := storeQueryError(db, clusterID, query, err.Error()); storeErr != nil {
			fmt.Printf("Failed to store error: %v\n", storeErr)
		}
		return fmt.Errorf("query execution failed: %v", err)
	}

	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	// Store successful query execution
	queryID, err := storeQueryExecution(db, clusterID, query)
	if err != nil {
		fmt.Printf("Failed to store query: %v\n", err)
		return fmt.Errorf("failed to store query: %v", err)
	}

	// Store results
	err = storeQueryResults(db, queryID, result)
	if err != nil {
		fmt.Printf("Failed to store results: %v\n", err)
		return fmt.Errorf("failed to store results: %v", err)
	}

	fmt.Println("Results stored successfully in database")
	fmt.Println("------------------------")

	return nil
}
