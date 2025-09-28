package main

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/common/model"
)

// initDB initializes the SQLite database and creates required tables
func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./kpi_metrics.db")
	if err != nil {
		return nil, err
	}

	schema := `
    CREATE TABLE IF NOT EXISTS clusters (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        cluster_name TEXT UNIQUE NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE TABLE IF NOT EXISTS queries (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        cluster_id INTEGER REFERENCES clusters(id),
        query_text TEXT NOT NULL,
        execution_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        status TEXT DEFAULT 'success',
        error_message TEXT
    );
    
    CREATE TABLE IF NOT EXISTS query_results (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        query_id INTEGER REFERENCES queries(id),
        metric_value REAL,
        timestamp_value REAL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        metric_labels TEXT  -- JSON string of all labels
    );
    `

	_, err = db.Exec(schema)
	return db, err
}

// getOrCreateCluster gets existing cluster ID or creates a new cluster record
func getOrCreateCluster(db *sql.DB, clusterName string) (int64, error) {
	var clusterID int64
	err := db.QueryRow("SELECT id FROM clusters WHERE cluster_name = ?", clusterName).Scan(&clusterID)
	if err == nil {
		return clusterID, nil
	}

	result, err := db.Exec("INSERT INTO clusters (cluster_name) VALUES (?)", clusterName)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// storeQueryError stores a failed query execution in the database
func storeQueryError(db *sql.DB, clusterID int64, queryText string, errorMsg string) error {
	_, err := db.Exec(
		"INSERT INTO queries (cluster_id, query_text, status, error_message) VALUES (?, ?, 'error', ?)",
		clusterID, queryText, errorMsg,
	)
	return err
}

// storeQueryExecution stores a successful query execution and returns the query ID
func storeQueryExecution(db *sql.DB, clusterID int64, queryText string) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO queries (cluster_id, query_text) VALUES (?, ?)",
		clusterID, queryText,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// storeQueryResults stores the results of a Prometheus query in the database
func storeQueryResults(db *sql.DB, queryID int64, result model.Value) error {
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
            (query_id, metric_value, timestamp_value, metric_labels)
            VALUES (?, ?, ?, ?)`,
			queryID, value, timestamp, string(labelsJSON),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
