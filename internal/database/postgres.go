package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/prometheus/common/model"
)

// PostgresDB implements the Database interface for PostgreSQL
type PostgresDB struct {
	ConnectionURL string
}

// NewPostgresDB creates a new PostgreSQL database instance
func NewPostgresDB(connectionURL string) *PostgresDB {
	return &PostgresDB{ConnectionURL: connectionURL}
}

// InitDB initializes the PostgreSQL database and creates required tables
func (p *PostgresDB) InitDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", p.ConnectionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %v", err)
	}

	schema := `
    CREATE TABLE IF NOT EXISTS clusters (
        id SERIAL PRIMARY KEY,
        cluster_name TEXT UNIQUE NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE TABLE IF NOT EXISTS query_results (
        id SERIAL PRIMARY KEY,
        kpi_id TEXT NOT NULL,
        metric_value DOUBLE PRECISION,
        timestamp_value DOUBLE PRECISION,
        cluster_id INTEGER NOT NULL REFERENCES clusters(id),
        execution_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        metric_labels JSONB
    );

    CREATE TABLE IF NOT EXISTS query_errors (
        id SERIAL PRIMARY KEY,
        kpi_id TEXT UNIQUE NOT NULL,
        errors INTEGER DEFAULT 0
    );

    -- Create indexes for better query performance
    CREATE INDEX IF NOT EXISTS idx_query_results_cluster_id ON query_results(cluster_id);
    CREATE INDEX IF NOT EXISTS idx_query_results_kpi_id ON query_results(kpi_id);
    CREATE INDEX IF NOT EXISTS idx_query_results_created_at ON query_results(created_at);
    CREATE INDEX IF NOT EXISTS idx_query_results_labels ON query_results USING GIN(metric_labels);
    `

	_, err = db.Exec(schema)
	return db, err
}

// GetOrCreateCluster gets existing cluster ID or creates a new cluster record
func (p *PostgresDB) GetOrCreateCluster(db *sql.DB, clusterName string) (int64, error) {
	var clusterID int64

	// Try to get existing cluster
	err := db.QueryRow("SELECT id FROM clusters WHERE cluster_name = $1", clusterName).Scan(&clusterID)
	if err == nil {
		return clusterID, nil
	}

	// Insert new cluster and return ID
	err = db.QueryRow(
		"INSERT INTO clusters (cluster_name) VALUES ($1) ON CONFLICT (cluster_name) DO UPDATE SET cluster_name = EXCLUDED.cluster_name RETURNING id",
		clusterName,
	).Scan(&clusterID)

	return clusterID, err
}

// IncrementQueryError increments the error count for a given KPI ID
func (p *PostgresDB) IncrementQueryError(db *sql.DB, kpiID string) error {
	_, err := db.Exec(`
        INSERT INTO query_errors (kpi_id, errors) VALUES ($1, 1)
        ON CONFLICT(kpi_id) DO UPDATE SET errors = query_errors.errors + 1
    `, kpiID)
	return err
}

// GetQueryErrorCount returns the error count for a given KPI ID
func (p *PostgresDB) GetQueryErrorCount(db *sql.DB, kpiID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT errors FROM query_errors WHERE kpi_id = $1", kpiID).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// StoreQueryResults stores the results of a Prometheus query in the database
func (p *PostgresDB) StoreQueryResults(db *sql.DB, clusterID int64, queryID string, result model.Value) error {
	vector := result.(model.Vector)
	for _, sample := range vector {
		metric := sample.Metric
		value := float64(sample.Value)
		timestamp := float64(sample.Timestamp) / 1000

		labelsJSON, err := json.Marshal(metric)
		if err != nil {
			return err
		}

		_, err = db.Exec(`
            INSERT INTO query_results 
            (kpi_id, metric_value, timestamp_value, cluster_id, metric_labels)
            VALUES ($1, $2, $3, $4, $5)`,
			queryID, value, timestamp, clusterID, string(labelsJSON),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
