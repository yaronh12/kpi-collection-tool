package database

import (
	"database/sql"

	"github.com/prometheus/common/model"
)

// Database defines the interface that all database implementations must satisfy
type Database interface {
	// InitDB initializes the database and creates required tables
	InitDB() (*sql.DB, error)

	// GetOrCreateCluster gets existing cluster ID or creates a new cluster record
	GetOrCreateCluster(db *sql.DB, clusterName string) (int64, error)

	// IncrementQueryError increments the error count for a given KPI ID
	IncrementQueryError(db *sql.DB, kpiID string) error

	// GetQueryErrorCount returns the error count for a given KPI ID
	GetQueryErrorCount(db *sql.DB, kpiID string) (int, error)

	// StoreQueryResults stores the results of a Prometheus query in the database
	StoreQueryResults(db *sql.DB, clusterID int64, queryID string, result model.Value) error
}
