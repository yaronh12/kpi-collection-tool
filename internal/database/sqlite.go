package database

import (
	"database/sql"
	"encoding/json"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/common/model"
)

// initDB initializes the SQLite database and creates required tables
func InitDB() (*sql.DB, error) {
	// Create collected-data directory if it doesn't exist
	if err := os.MkdirAll("./collected-data", 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", "./collected-data/kpi_metrics.db")
	if err != nil {
		return nil, err
	}

	schema := `
    CREATE TABLE IF NOT EXISTS clusters (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        cluster_name TEXT UNIQUE NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    
    
    CREATE TABLE IF NOT EXISTS query_results (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        kpi_id TEXT NOT NULL,
        metric_value REAL,
        timestamp_value REAL,
		cluster_id INTEGER NOT NULL REFERENCES clusters(id),
		execution_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        metric_labels TEXT  -- JSON string of all labels
    );

	CREATE TABLE IF NOT EXISTS query_errors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		kpi_id TEXT UNIQUE NOT NULL,
		errors INTEGER DEFAULT 0
	)
    `

	_, err = db.Exec(schema)
	return db, err
}

// getOrCreateCluster gets existing cluster ID or creates a new cluster record
func GetOrCreateCluster(db *sql.DB, clusterName string) (int64, error) {
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

// increments the error count for a given KPI ID in the query_errors table.
func IncrementQueryError(db *sql.DB, kpiID string) error {
	_, err := db.Exec(`
        INSERT INTO query_errors (kpi_id, errors) VALUES (?, 1)
        ON CONFLICT(kpi_id) DO UPDATE SET errors = errors + 1
    `, kpiID)
	return err
}

// returns the error count for a given KPI ID.
func GetQueryErrorCount(db *sql.DB, kpiID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT errors FROM query_errors WHERE kpi_id = ?", kpiID).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// storeQueryResults stores the results of a Prometheus query in the database
func StoreQueryResults(db *sql.DB, clusterID int64, queryID string, result model.Value) error {
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
            VALUES (?, ?, ?, ?, ?)`,
			queryID, value, timestamp, clusterID, string(labelsJSON),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
