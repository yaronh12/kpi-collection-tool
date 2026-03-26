package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
	"github.com/prometheus/common/model"
)

const (
	// DefaultDataDir is the artifact directory created in the user's working directory
	DefaultDataDir = "kpi-collector-artifacts"
	// DefaultDBFileName is the SQLite database file name
	DefaultDBFileName = "kpi_metrics.db"
)

type SQLiteDB struct{}

// NewSQLiteDB creates a new SQLite database instance
func NewSQLiteDB() *SQLiteDB {
	return &SQLiteDB{}
}

// InitDB initializes the SQLite database and creates required tables.
// The database is stored in ./kpi-collector-artifacts/kpi_metrics.db (relative to CWD).
func (sqlite_db *SQLiteDB) InitDB() (*sql.DB, error) {
	dbPath := filepath.Join(DefaultDataDir, DefaultDBFileName)

	if err := os.MkdirAll(DefaultDataDir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	schema := `
    CREATE TABLE IF NOT EXISTS clusters (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        cluster_name TEXT UNIQUE NOT NULL,
		cluster_type TEXT,
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
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_query_results_dedup
	ON query_results(kpi_id, cluster_id, timestamp_value, metric_labels)
    `

	_, err = db.Exec(schema)
	return db, err
}

// getOrCreateCluster gets existing cluster ID or creates a new cluster record
func (sqlite_db *SQLiteDB) GetOrCreateCluster(db *sql.DB, clusterName string, clusterType string) (int64, error) {
	var clusterID int64
	err := db.QueryRow("SELECT id FROM clusters WHERE cluster_name = ?", clusterName).Scan(&clusterID)
	if err == nil {
		if clusterType != "" {
			_, updateErr := db.Exec("UPDATE clusters SET cluster_type = ? WHERE id = ?", clusterType, clusterID)
			if updateErr != nil {
				return clusterID, updateErr
			}
		}
		return clusterID, nil
	}

	result, err := db.Exec("INSERT INTO clusters (cluster_name, cluster_type) VALUES (?, ?)", clusterName, clusterType)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// increments the error count for a given KPI ID in the query_errors table.
func (sqlite_db *SQLiteDB) IncrementQueryError(db *sql.DB, kpiID string) error {
	_, err := db.Exec(`
        INSERT INTO query_errors (kpi_id, errors) VALUES (?, 1)
        ON CONFLICT(kpi_id) DO UPDATE SET errors = errors + 1
    `, kpiID)
	return err
}

// returns the error count for a given KPI ID.
func (sqlite_db *SQLiteDB) GetQueryErrorCount(db *sql.DB, kpiID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT errors FROM query_errors WHERE kpi_id = ?", kpiID).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// StoreQueryResults stores the results of a Prometheus query in the database
func (sqlite_db *SQLiteDB) StoreQueryResults(db *sql.DB, clusterID int64, queryID string, result model.Value) error {
	switch values := result.(type) {
	case model.Vector:
		return sqlite_db.storeVectorResults(db, clusterID, queryID, values)
	case model.Matrix:
		return sqlite_db.storeMatrixResults(db, clusterID, queryID, values)
	default:
		return fmt.Errorf("unsupported Prometheus result type for KPI '%s': %T", queryID, result)
	}
}

func (sqlite_db *SQLiteDB) storeVectorResults(db *sql.DB, clusterID int64, queryID string, vector model.Vector) error {
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
            VALUES (?, ?, ?, ?, ?)
            ON CONFLICT(kpi_id, cluster_id, timestamp_value, metric_labels) DO NOTHING`,
			queryID, value, timestamp, clusterID, string(labelsJSON),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sqlite_db *SQLiteDB) storeMatrixResults(db *sql.DB, clusterID int64, queryID string, matrix model.Matrix) error {
	for _, stream := range matrix {
		metric := stream.Metric
		labelsJSON, err := json.Marshal(metric)
		if err != nil {
			return err
		}

		for _, samplePair := range stream.Values {
			value := float64(samplePair.Value)
			timestamp := float64(samplePair.Timestamp) / 1000

			_, execErr := db.Exec(`
                INSERT INTO query_results 
                (kpi_id, metric_value, timestamp_value, cluster_id, metric_labels)
                VALUES (?, ?, ?, ?, ?)
                ON CONFLICT(kpi_id, cluster_id, timestamp_value, metric_labels) DO NOTHING`,
				queryID, value, timestamp, clusterID, string(labelsJSON),
			)
			if execErr != nil {
				return execErr
			}
		}
	}

	return nil
}
