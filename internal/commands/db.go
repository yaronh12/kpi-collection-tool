package commands

import (
	"database/sql"
	"fmt"
	"os"

	"kpi-collector/internal/database"

	"github.com/spf13/cobra"
)

// dbFlags holds database connection flags for db commands
var dbFlags struct {
	DatabaseType string
	PostgresURL  string
}

// dbCmd represents the db command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Query and manage collected KPI data",
	Long: `Direct access to query and manage the collected KPI data stored in the database.

Supports querying KPI metrics, listing clusters, viewing errors, and cleaning up data.
Works with both SQLite (default) and PostgreSQL databases.

Database connection can be specified via:
  1. CLI flags: --db-type and --postgres-url
  2. Environment variables: KPI_COLLECTOR_DB_TYPE and KPI_COLLECTOR_DB_URL
  3. Default: SQLite at ~/.local/share/kpi-collector/kpi_metrics.db`,
	Example: `  # Using SQLite (default)
  kpi-collector db show clusters
  
  # Using PostgreSQL (via flags)
  kpi-collector db show kpis --name="cpu-system" \
    --db-type=postgres --postgres-url="postgresql://user:pass@localhost:5432/kpi"
  
  # Using PostgreSQL (via environment variables)
  export KPI_COLLECTOR_DB_TYPE=postgres
  export KPI_COLLECTOR_DB_URL="postgresql://user:pass@localhost:5432/kpi"
  kpi-collector db show clusters`,
}

func init() {
	// Register the db command as a subcommand of root
	rootCmd.AddCommand(dbCmd)

	// Add persistent flags that apply to all db subcommands
	dbCmd.PersistentFlags().StringVar(&dbFlags.DatabaseType, "db-type", "",
		"database type: sqlite (default) or postgres")
	dbCmd.PersistentFlags().StringVar(&dbFlags.PostgresURL, "postgres-url", "",
		"PostgreSQL connection string")
}

// connectToDB establishes a database connection using flags or environment variables
func connectToDB() (*sql.DB, database.Database, error) {
	// Priority 1: CLI flags
	dbType := dbFlags.DatabaseType
	postgresURL := dbFlags.PostgresURL

	// Priority 2: Environment variables (if flags not provided)
	if dbType == "" {
		dbType = os.Getenv("KPI_COLLECTOR_DB_TYPE")
	}
	if postgresURL == "" {
		postgresURL = os.Getenv("KPI_COLLECTOR_DB_URL")
	}

	// Priority 3: Default to SQLite
	if dbType == "" {
		dbType = "sqlite"
	}

	// Validate PostgreSQL URL if needed
	if dbType == "postgres" && postgresURL == "" {
		return nil, nil, fmt.Errorf("PostgreSQL connection URL is required.\n" +
			"Provide via --postgres-url flag or KPI_COLLECTOR_DB_URL environment variable")
	}

	// Create database implementation
	var dbImpl database.Database
	switch dbType {
	case "postgres":
		dbImpl = database.NewPostgresDB(postgresURL)
	case "sqlite":
		dbImpl = database.NewSQLiteDB()
	default:
		return nil, nil, fmt.Errorf("invalid database type: %s (must be 'sqlite' or 'postgres')", dbType)
	}

	// Initialize database connection
	db, err := dbImpl.InitDB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, dbImpl, nil
}
