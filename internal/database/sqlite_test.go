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
		db        *sql.DB
		tmpDir    string
		sqliteDB  *SQLiteDB
		originCwd string
	)

	BeforeEach(func() {
		sqliteDB = NewSQLiteDB()

		var err error
		tmpDir, err = os.MkdirTemp("", "sqlite-test-*")
		Expect(err).NotTo(HaveOccurred())

		// Change to temp directory so InitDB creates kpi-collector/ there
		originCwd, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		err = os.Chdir(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		db, err = sqliteDB.InitDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(db).NotTo(BeNil())
	})

	AfterEach(func() {
		if db != nil {
			err := db.Close()
			Expect(err).NotTo(HaveOccurred())
		}
		if originCwd != "" {
			err := os.Chdir(originCwd)
			Expect(err).NotTo(HaveOccurred())
		}
		if tmpDir != "" {
			err := os.RemoveAll(tmpDir)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("SQLite-Specific Features", func() {
		It("should create the database and required tables", func() {
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

		It("should create the data directory", func() {
			dataDir := filepath.Join(tmpDir, DefaultDataDir)
			_, err := os.Stat(dataDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create the database file", func() {
			dbFile := filepath.Join(tmpDir, DefaultDataDir, DefaultDBFileName)
			_, err := os.Stat(dbFile)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	RunDatabaseInterfaceTests(func() (Database, *sql.DB) { return sqliteDB, db })
})
