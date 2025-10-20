package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"rds-kpi-collector/database"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

const (
	FIVE_SECONDS = 5 * time.Second
)

func (t *tokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.rt.RoundTrip(req)
}

// setupPromClient creates and configures a Prometheus API client
func setupPromClient(thanosURL, bearerToken string, insecureTLS bool) (promv1.API, error) {
	client, err := api.NewClient(api.Config{
		Address: "https://" + thanosURL,
		RoundTripper: &tokenRoundTripper{
			token: bearerToken,
			rt: &http.Transport{
				// NOTE: InsecureSkipVerify is set to true for development purposes only.
				// In production environments, this should be false and proper certificate
				// validation should be implemented.
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureTLS},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client: %v", err)
	}

	return promv1.NewAPI(client), nil
}

// runQueries executes all Prometheus queries and stores results in database
func runQueries(kpisToRun KPIs, flags InputFlags) error {
	// Initialize Database
	db, err := database.InitDB()
	if err != nil {
		return fmt.Errorf("failed to init database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	// Get or create cluster in DB
	clusterID, err := database.GetOrCreateCluster(db, flags.ClusterName)
	if err != nil {
		return fmt.Errorf("failed to get cluster ID: %v", err)
	}

	// Create Prometheus client
	v1api, err := setupPromClient(flags.ThanosURL, flags.BearerToken, flags.InsecureTLS)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	// Execute queries
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(len(kpisToRun.Queries))*FIVE_SECONDS)
	defer cancel()

	for _, query := range kpisToRun.Queries {
		err := executeQuery(ctx, v1api, db, clusterID, query.ID, query.PromQuery)
		if err != nil {
			fmt.Printf("Query execution failed: %v\n", err)
			// Continue with next query even if one fails
		}
	}

	return nil
}

// executeQuery executes a single Prometheus query and handles the result
func executeQuery(ctx context.Context, v1api promv1.API, db *sql.DB, clusterID int64, queryID string, queryText string) error {
	fmt.Println("------------------------")
	fmt.Printf("Running: %s\n", queryText)

	// Execute query using the Prometheus client library
	result, warnings, err := v1api.Query(ctx, queryText, time.Now())
	if err != nil {
		fmt.Println("query failed: ", err)
		if storeErr := database.IncrementQueryError(db, queryID); storeErr != nil {
			fmt.Printf("Failed to increment error count: %v\n", storeErr)
		}
		return fmt.Errorf("query execution failed: %v", err)
	}

	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	// Store results
	err = database.StoreQueryResults(db, clusterID, queryID, result)
	if err != nil {
		fmt.Printf("Failed to store results: %v\n", err)
		return fmt.Errorf("failed to store results: %v", err)
	}

	fmt.Println("Results stored successfully in database")
	fmt.Println("------------------------")

	return nil
}
