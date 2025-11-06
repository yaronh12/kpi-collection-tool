package database

import (
	"database/sql"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sqlite", func() {
	var (
		db       *sql.DB
		tmpDir   string
		sqliteDB *SQLiteDB
	)

	// Runs before and after each test (It section)
	// To provide clean, isolated environment for each test
	BeforeEach(func() {
		// Create SQLiteDB instance
		sqliteDB = NewSQLiteDB()

		// Create a temporary directory for test database
		var err error
		tmpDir, err = os.MkdirTemp("", "sqlite-test-*")
		Expect(err).NotTo(HaveOccurred())

		// Change to temp directory for database creation
		originalDir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		err = os.Chdir(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		// Initialize database
		db, err = sqliteDB.InitDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(db).NotTo(BeNil())

		// Change back to original directory
		err = os.Chdir(originalDir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if db != nil {
			err := db.Close()
			Expect(err).NotTo(HaveOccurred())
		}
		// Clean up temporary directory
		if tmpDir != "" {
			err := os.RemoveAll(tmpDir)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("SQLite-Specific Features", func() {
		It("should create the database and required tables", func() {
			// sqlite_master = special system table in SQLite that contains
			// metadata about all the database objects (tables, indexes, views, triggers)
			// in your SQLite database.
			var tableName string
			err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='clusters'").Scan(&tableName)
			Expect(err).NotTo(HaveOccurred())
			Expect(tableName).To(Equal("clusters"))

			err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='query_results'").Scan(&tableName)
			Expect(err).NotTo(HaveOccurred())
			Expect(tableName).To(Equal("query_results"))

			err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='query_errors'").Scan(&tableName)
			Expect(err).NotTo(HaveOccurred())
			Expect(tableName).To(Equal("query_errors"))
		})

		It("should create the collected-data directory", func() {
			dbPath := filepath.Join(tmpDir, "collected-data")
			_, err := os.Stat(dbPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create the database file", func() {
			dbFile := filepath.Join(tmpDir, "collected-data", "kpi_metrics.db")
			_, err := os.Stat(dbFile)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	// Run all the shared interface tests
	RunDatabaseInterfaceTests(func() (Database, *sql.DB) { return sqliteDB, db })
})
