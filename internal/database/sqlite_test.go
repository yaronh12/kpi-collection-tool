package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/model"
)

var _ = Describe("Sqlite", func() {
	var (
		db     *sql.DB
		tmpDir string
	)

	// Runs before and after each test (It section)
	// To provide clean, isolated environment for each test
	BeforeEach(func() {
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
		db, err = InitDB()
		Expect(err).NotTo(HaveOccurred())
		Expect(db).NotTo(BeNil())

		// Change back to original directory
		err = os.Chdir(originalDir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if db != nil {
			db.Close()
		}
		// Clean up temporary directory
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	Describe("InitDB - Testing the initialization of the DB ", func() {
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

	Describe("GetOrCreateCluster - Testing the creation of new cluster entry", func() {
		Context("when cluster does not exist", func() {
			It("should create a new cluster and return its ID", func() {
				By("Calls GetOrCreateCluster() with a cluster name that doesn't exist and Verifies a positive ID is returned")
				clusterID, err := GetOrCreateCluster(db, "test-cluster")
				Expect(err).NotTo(HaveOccurred())
				Expect(clusterID).To(BeNumerically(">", 0))

				By("Queries the database to confirm the cluster name was stored")
				var clusterName string
				err = db.QueryRow("SELECT cluster_name FROM clusters WHERE id = ?", clusterID).Scan(&clusterName)
				Expect(err).NotTo(HaveOccurred())
				Expect(clusterName).To(Equal("test-cluster"))
			})

			It("should set created_at timestamp", func() {
				By("Creates a cluster, queries the created_at aolumn and verifies its not empty.")
				clusterID, err := GetOrCreateCluster(db, "test-cluster")
				Expect(err).NotTo(HaveOccurred())

				var createdAt string
				err = db.QueryRow("SELECT created_at FROM clusters WHERE id = ?", clusterID).Scan(&createdAt)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdAt).NotTo(BeEmpty())
			})

			It("should create different IDs for different clusters", func() {
				By("Creates two clusters with different names and verifies they get different IDs")
				clusterID1, err := GetOrCreateCluster(db, "cluster-1")
				Expect(err).NotTo(HaveOccurred())

				clusterID2, err := GetOrCreateCluster(db, "cluster-2")
				Expect(err).NotTo(HaveOccurred())

				Expect(clusterID1).NotTo(Equal(clusterID2))
			})
		})

		Context("when cluster already exists", func() {
			var existingClusterID int64

			BeforeEach(func() {
				// Creates a cluster named "existing-cluster" and stores its ID
				var err error
				existingClusterID, err = GetOrCreateCluster(db, "existing-cluster")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return the existing cluster ID", func() {
				By("Calls GetOrCreateCluster() again with the same name and verifies it returns the same ID")
				clusterID, err := GetOrCreateCluster(db, "existing-cluster")
				Expect(err).NotTo(HaveOccurred())
				Expect(clusterID).To(Equal(existingClusterID))
			})

			It("should not create a duplicate entry", func() {
				By("Calls GetOrCreateCluster() again, Counts how many rows exist with that cluster name and verifies count is exactly 1")
				GetOrCreateCluster(db, "existing-cluster")

				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM clusters WHERE cluster_name = ?", "existing-cluster").Scan(&count)
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(1))
			})
		})
	})

	Describe("IncrementQueryError - Testing incremantation of error in kpi entries", func() {
		Context("when KPI has no previous errors", func() {
			It("should create a new error record with count 1", func() {
				By("Calls IncrementQueryError() for a new KPI ID, reads back the countand verifies it's 1")
				err := IncrementQueryError(db, "test-kpi-1")
				Expect(err).NotTo(HaveOccurred())

				count, err := GetQueryErrorCount(db, "test-kpi-1")
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(1))
			})
		})

		Context("when KPI already has errors", func() {
			BeforeEach(func() {
				// Creates one error (count = 1)
				err := IncrementQueryError(db, "existing-kpi")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should increment the existing error count", func() {
				By("Increments twice more (count should become 3), reads back the count and verifies it's 3")
				err := IncrementQueryError(db, "existing-kpi")
				Expect(err).NotTo(HaveOccurred())

				err = IncrementQueryError(db, "existing-kpi")
				Expect(err).NotTo(HaveOccurred())

				count, err := GetQueryErrorCount(db, "existing-kpi")
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(3))
			})
		})

		Context("with multiple different KPI IDs", func() {
			It("should handle them independently", func() {
				By("Increments kpi-a twice (count = 2), Increments kpi-b once (count = 1) and Verifies each has the correct count")
				err := IncrementQueryError(db, "kpi-a")
				Expect(err).NotTo(HaveOccurred())
				err = IncrementQueryError(db, "kpi-a")
				Expect(err).NotTo(HaveOccurred())

				err = IncrementQueryError(db, "kpi-b")
				Expect(err).NotTo(HaveOccurred())

				countA, err := GetQueryErrorCount(db, "kpi-a")
				Expect(err).NotTo(HaveOccurred())
				Expect(countA).To(Equal(2))

				countB, err := GetQueryErrorCount(db, "kpi-b")
				Expect(err).NotTo(HaveOccurred())
				Expect(countB).To(Equal(1))
			})
		})
	})

	Describe("GetQueryErrorCount - Testing retrieval of error count", func() {
		Context("when KPI ID does not exist", func() {
			It("should return 0", func() {
				By("Queries a KPI ID that was never created, Verifies it returns 0 (not an error)")
				count, err := GetQueryErrorCount(db, "non-existent-kpi")
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(0))
			})
		})

		Context("when KPI ID exists", func() {
			BeforeEach(func() {
				for i := 0; i < 5; i++ {
					err := IncrementQueryError(db, "test-kpi")
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("should return the correct count", func() {
				By("Reads the amount, verifies its 5")
				count, err := GetQueryErrorCount(db, "test-kpi")
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(5))
			})
		})
	})

	Describe("StoreQueryResults - Testing kpi metric creation", func() {
		var clusterID int64

		BeforeEach(func() {
			var err error
			clusterID, err = GetOrCreateCluster(db, "test-cluster")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("with a single metric", func() {
			It("should store the metric value correctly", func() {
				By("Creates a Prometheus Vector with one sample, Sample has metric name, labels, value (42.5), and timestamp")
				vector := model.Vector{
					&model.Sample{
						Metric: model.Metric{
							"__name__": "test_metric",
							"label1":   "value1",
							"label2":   "value2",
						},
						Value:     model.SampleValue(42.5),
						Timestamp: model.Time(time.Now().Unix() * 1000),
					},
				}

				By("Stores it in the database")
				err := StoreQueryResults(db, clusterID, "test-query-1", vector)
				Expect(err).NotTo(HaveOccurred())

				By("Verifies one row was created and Verifies all values are correct")
				var count int
				err = db.QueryRow("SELECT COUNT(*) FROM query_results WHERE kpi_id = ?", "test-query-1").Scan(&count)
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(1))

				var metricValue, timestampValue float64
				var metricLabels string
				var storedClusterID int64
				err = db.QueryRow(`
					SELECT metric_value, timestamp_value, cluster_id, metric_labels 
					FROM query_results 
					WHERE kpi_id = ?
				`, "test-query-1").Scan(&metricValue, &timestampValue, &storedClusterID, &metricLabels)
				Expect(err).NotTo(HaveOccurred())
				Expect(metricValue).To(Equal(42.5))
				Expect(storedClusterID).To(Equal(clusterID))
				Expect(metricLabels).To(ContainSubstring("test_metric"))
				Expect(metricLabels).To(ContainSubstring("label1"))
				Expect(metricLabels).To(ContainSubstring("value1"))
			})

			It("should store metric labels as JSON", func() {
				By("Creates a sample with complex labels (namespace, pod, container)")
				vector := model.Vector{
					&model.Sample{
						Metric: model.Metric{
							"__name__":  "cpu_usage",
							"namespace": "default",
							"pod":       "my-pod",
							"container": "my-container",
						},
						Value:     model.SampleValue(75.5),
						Timestamp: model.Time(time.Now().Unix() * 1000),
					},
				}
				By("Stores it")
				err := StoreQueryResults(db, clusterID, "test-query-3", vector)
				Expect(err).NotTo(HaveOccurred())
				By("Verifies the JSON contains all label keys and values")
				var metricLabels string
				err = db.QueryRow("SELECT metric_labels FROM query_results WHERE kpi_id = ?", "test-query-3").Scan(&metricLabels)
				Expect(err).NotTo(HaveOccurred())
				Expect(metricLabels).To(ContainSubstring("cpu_usage"))
				Expect(metricLabels).To(ContainSubstring("namespace"))
				Expect(metricLabels).To(ContainSubstring("default"))
				Expect(metricLabels).To(ContainSubstring("pod"))
				Expect(metricLabels).To(ContainSubstring("my-pod"))
			})

			It("should set execution_time and created_at timestamps", func() {
				vector := model.Vector{
					&model.Sample{
						Metric:    model.Metric{"__name__": "test"},
						Value:     model.SampleValue(1.0),
						Timestamp: model.Time(time.Now().Unix() * 1000),
					},
				}
				By("Stores a metric, Queries the timestamp columns and verifies they are not empty.")
				err := StoreQueryResults(db, clusterID, "test-query-4", vector)
				Expect(err).NotTo(HaveOccurred())

				var executionTime, createdAt string
				err = db.QueryRow(`
					SELECT execution_time, created_at 
					FROM query_results 
					WHERE kpi_id = ?
				`, "test-query-4").Scan(&executionTime, &createdAt)
				Expect(err).NotTo(HaveOccurred())
				Expect(executionTime).NotTo(BeEmpty())
				Expect(createdAt).NotTo(BeEmpty())
			})
		})

		Context("with multiple metrics", func() {
			It("should store all metrics from the vector", func() {
				By("Creates a vector with 3 samples and stores them")
				vector := model.Vector{
					&model.Sample{
						Metric:    model.Metric{"__name__": "metric1"},
						Value:     model.SampleValue(10.0),
						Timestamp: model.Time(time.Now().Unix() * 1000),
					},
					&model.Sample{
						Metric:    model.Metric{"__name__": "metric2"},
						Value:     model.SampleValue(20.0),
						Timestamp: model.Time(time.Now().Unix() * 1000),
					},
					&model.Sample{
						Metric:    model.Metric{"__name__": "metric3"},
						Value:     model.SampleValue(30.0),
						Timestamp: model.Time(time.Now().Unix() * 1000),
					},
				}

				err := StoreQueryResults(db, clusterID, "test-query-2", vector)
				Expect(err).NotTo(HaveOccurred())
				By("Counts rows in database and verifies 3 rows were created")
				var count int
				err = db.QueryRow("SELECT COUNT(*) FROM query_results WHERE kpi_id = ?", "test-query-2").Scan(&count)
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(3))
			})
		})

		Context("with multiple clusters", func() {
			var clusterID2 int64

			BeforeEach(func() {
				var err error
				clusterID2, err = GetOrCreateCluster(db, "another-cluster")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should associate results with the correct cluster", func() {
				By("Creates 'another-cluster' in addition to 'test-cluster'")
				vector1 := model.Vector{
					&model.Sample{
						Metric:    model.Metric{"__name__": "metric"},
						Value:     model.SampleValue(1.0),
						Timestamp: model.Time(time.Now().Unix() * 1000),
					},
				}

				vector2 := model.Vector{
					&model.Sample{
						Metric:    model.Metric{"__name__": "metric"},
						Value:     model.SampleValue(2.0),
						Timestamp: model.Time(time.Now().Unix() * 1000),
					},
				}
				By("Stores metrics for cluster 1 with query ID 'query-cluster-1'")
				By("Stores metrics for cluster 2 with query ID 'query-cluster-2'")
				By("Verifies each result is associated with the correct cluster ID")
				err := StoreQueryResults(db, clusterID, "query-cluster-1", vector1)
				Expect(err).NotTo(HaveOccurred())

				err = StoreQueryResults(db, clusterID2, "query-cluster-2", vector2)
				Expect(err).NotTo(HaveOccurred())

				var storedClusterID int64
				err = db.QueryRow("SELECT cluster_id FROM query_results WHERE kpi_id = ?", "query-cluster-1").Scan(&storedClusterID)
				Expect(err).NotTo(HaveOccurred())
				Expect(storedClusterID).To(Equal(clusterID))

				err = db.QueryRow("SELECT cluster_id FROM query_results WHERE kpi_id = ?", "query-cluster-2").Scan(&storedClusterID)
				Expect(err).NotTo(HaveOccurred())
				Expect(storedClusterID).To(Equal(clusterID2))
			})
		})
	})
})
