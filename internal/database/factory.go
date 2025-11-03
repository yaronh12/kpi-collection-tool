package database

import (
	"database/sql"
	"fmt"
	"rds-kpi-collector/internal/config"
)

// NewDatabase creates a database instance based on the configuration
func NewDatabase(flags config.InputFlags) (Database, error) {
	switch flags.DatabaseType {
	case "sqlite":
		return NewSQLiteDB(), nil
	case "postgres":
		if flags.PostgresURL == "" {
			return nil, fmt.Errorf("postgres-url is required for postgres database type")
		}
		return NewPostgresDB(flags.PostgresURL), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", flags.DatabaseType)
	}
}

// InitDatabaseWithConfig initializes a database based on configuration flags
func InitDatabaseWithConfig(flags config.InputFlags) (*sql.DB, Database, error) {
	dbImpl, err := NewDatabase(flags)
	if err != nil {
		return nil, nil, err
	}

	db, err := dbImpl.InitDB()
	if err != nil {
		return nil, nil, err
	}

	return db, dbImpl, nil
}
