package database

import (
	"database/sql"
	"fmt"
)

// NewDatabase creates a database instance based on the configuration
func NewDatabase(databaseType string, postgresURL string) (Database, error) {
	switch databaseType {
	case "sqlite":
		return NewSQLiteDB(), nil
	case "postgres":
		if postgresURL == "" {
			return nil, fmt.Errorf("postgres-url is required for postgres database type")
		}
		return NewPostgresDB(postgresURL), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", databaseType)
	}
}

// InitDatabaseWithConfig initializes a database based on configuration flags
func InitDatabaseWithConfig(databaseType string, postgresURL string) (*sql.DB, Database, error) {
	dbImpl, err := NewDatabase(databaseType, postgresURL)
	if err != nil {
		return nil, nil, err
	}

	db, err := dbImpl.InitDB()
	if err != nil {
		return nil, nil, err
	}

	return db, dbImpl, nil
}
