package config

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("KPIs Loader", func() {
	var (
		tmpDir string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "kpis-loader-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tmpDir != "" {
			err := os.RemoveAll(tmpDir)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("LoadKPIs", func() {
		Context("when loading a valid KPIs file", func() {
			It("should load KPIs without custom frequency", func() {
				kpisJSON := `{
					"kpis": [
						{
							"id": "cpu-usage",
							"promquery": "avg(rate(node_cpu_seconds_total[5m]))"
						}
					]
				}`
				kpisPath := filepath.Join(tmpDir, "kpis.json")
				err := os.WriteFile(kpisPath, []byte(kpisJSON), 0644)
				Expect(err).NotTo(HaveOccurred())

				kpis, err := LoadKPIs(kpisPath)

				Expect(err).NotTo(HaveOccurred())
				Expect(kpis.Queries).To(HaveLen(1))
				Expect(kpis.Queries[0].ID).To(Equal("cpu-usage"))
				Expect(kpis.Queries[0].PromQuery).To(Equal("avg(rate(node_cpu_seconds_total[5m]))"))
				Expect(kpis.Queries[0].SampleFrequency).To(BeNil())
			})

			It("should load KPIs with custom frequency as integer seconds", func() {
				kpisJSON := `{
					"kpis": [
						{
							"id": "memory-usage",
							"promquery": "node_memory_MemTotal_bytes",
							"sample-frequency": 120
						}
					]
				}`
				kpisPath := filepath.Join(tmpDir, "kpis.json")
				err := os.WriteFile(kpisPath, []byte(kpisJSON), 0644)
				Expect(err).NotTo(HaveOccurred())

				kpis, err := LoadKPIs(kpisPath)

				Expect(err).NotTo(HaveOccurred())
				Expect(kpis.Queries).To(HaveLen(1))
				Expect(kpis.Queries[0].ID).To(Equal("memory-usage"))
				Expect(kpis.Queries[0].SampleFrequency).NotTo(BeNil())
				Expect(kpis.Queries[0].SampleFrequency.Duration).To(Equal(120 * time.Second))
			})

			It("should load KPIs with custom frequency as duration string", func() {
				kpisJSON := `{
					"kpis": [
						{
							"id": "disk-usage",
							"promquery": "node_filesystem_avail_bytes",
							"sample-frequency": "2m30s"
						}
					]
				}`
				kpisPath := filepath.Join(tmpDir, "kpis.json")
				err := os.WriteFile(kpisPath, []byte(kpisJSON), 0644)
				Expect(err).NotTo(HaveOccurred())

				kpis, err := LoadKPIs(kpisPath)

				Expect(err).NotTo(HaveOccurred())
				Expect(kpis.Queries).To(HaveLen(1))
				Expect(kpis.Queries[0].SampleFrequency).NotTo(BeNil())
				Expect(kpis.Queries[0].SampleFrequency.Duration).To(Equal(2*time.Minute + 30*time.Second))
			})

			It("should load multiple KPIs", func() {
				kpisJSON := `{
					"kpis": [
						{"id": "kpi-1", "promquery": "query1"},
						{"id": "kpi-2", "promquery": "query2"},
						{"id": "kpi-3", "promquery": "query3", "sample-frequency": 60}
					]
				}`
				kpisPath := filepath.Join(tmpDir, "kpis.json")
				err := os.WriteFile(kpisPath, []byte(kpisJSON), 0644)
				Expect(err).NotTo(HaveOccurred())

				kpis, err := LoadKPIs(kpisPath)

				Expect(err).NotTo(HaveOccurred())
				Expect(kpis.Queries).To(HaveLen(3))
				Expect(kpis.Queries[0].ID).To(Equal("kpi-1"))
				Expect(kpis.Queries[1].ID).To(Equal("kpi-2"))
				Expect(kpis.Queries[2].ID).To(Equal("kpi-3"))
			})

			It("should load an empty KPIs array", func() {
				kpisJSON := `{"kpis": []}`
				kpisPath := filepath.Join(tmpDir, "kpis.json")
				err := os.WriteFile(kpisPath, []byte(kpisJSON), 0644)
				Expect(err).NotTo(HaveOccurred())

				kpis, err := LoadKPIs(kpisPath)

				Expect(err).NotTo(HaveOccurred())
				Expect(kpis.Queries).To(BeEmpty())
			})
		})

		Context("when the file does not exist", func() {
			It("should return an error", func() {
				nonExistentPath := filepath.Join(tmpDir, "nonexistent.json")

				kpis, err := LoadKPIs(nonExistentPath)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to open kpis file"))
				Expect(kpis.Queries).To(BeNil())
			})
		})

		Context("when the file contains invalid JSON", func() {
			It("should return an error for malformed JSON", func() {
				invalidJSON := `{"kpis": [{"id": "test"`
				kpisPath := filepath.Join(tmpDir, "invalid.json")
				err := os.WriteFile(kpisPath, []byte(invalidJSON), 0644)
				Expect(err).NotTo(HaveOccurred())

				kpis, err := LoadKPIs(kpisPath)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to decode kpis file"))
				Expect(kpis.Queries).To(BeNil())
			})

			It("should return an error for invalid duration string", func() {
				invalidDuration := `{
					"kpis": [
						{
							"id": "test",
							"promquery": "test",
							"sample-frequency": "invalid-duration"
						}
					]
				}`
				kpisPath := filepath.Join(tmpDir, "invalid-duration.json")
				err := os.WriteFile(kpisPath, []byte(invalidDuration), 0644)
				Expect(err).NotTo(HaveOccurred())

				kpis, err := LoadKPIs(kpisPath)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to decode kpis file"))
				Expect(kpis.Queries).To(BeNil())
			})
		})

		Context("when the file is empty", func() {
			It("should return an error", func() {
				emptyPath := filepath.Join(tmpDir, "empty.json")
				err := os.WriteFile(emptyPath, []byte(""), 0644)
				Expect(err).NotTo(HaveOccurred())

				kpis, err := LoadKPIs(emptyPath)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to decode kpis file"))
				Expect(kpis.Queries).To(BeNil())
			})
		})
	})

	Describe("ValidateKPIs", func() {
		Context("when all KPIs are valid", func() {
			It("should return no errors for valid PromQL queries", func() {
				kpis := KPIs{
					Queries: []Query{
						{ID: "cpu-usage", PromQuery: "avg(rate(node_cpu_seconds_total[5m]))"},
						{ID: "memory-usage", PromQuery: "node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes"},
					},
				}

				errors := ValidateKPIs(kpis)

				Expect(errors).To(BeEmpty())
			})

			It("should return no errors for empty KPIs list", func() {
				kpis := KPIs{Queries: []Query{}}

				errors := ValidateKPIs(kpis)

				Expect(errors).To(BeEmpty())
			})
		})

		Context("when KPIs have invalid PromQL syntax", func() {
			It("should return an error for invalid PromQL", func() {
				kpis := KPIs{
					Queries: []Query{
						{ID: "broken-query", PromQuery: "rate(metric[5m]"},
					},
				}

				errors := ValidateKPIs(kpis)

				Expect(errors).To(HaveLen(1))
				Expect(errors[0].Error()).To(ContainSubstring("broken-query"))
				Expect(errors[0].Error()).To(ContainSubstring("invalid PromQL syntax"))
			})

			It("should return multiple errors for multiple invalid queries", func() {
				kpis := KPIs{
					Queries: []Query{
						{ID: "broken-1", PromQuery: "rate(metric[5m]"},
						{ID: "valid", PromQuery: "up"},
						{ID: "broken-2", PromQuery: "sum("},
					},
				}

				errors := ValidateKPIs(kpis)

				Expect(errors).To(HaveLen(2))
			})
		})

		Context("when KPIs have duplicate IDs", func() {
			It("should return an error for duplicate IDs", func() {
				kpis := KPIs{
					Queries: []Query{
						{ID: "cpu-usage", PromQuery: "up"},
						{ID: "cpu-usage", PromQuery: "down"},
					},
				}

				errors := ValidateKPIs(kpis)

				Expect(errors).To(HaveLen(1))
				Expect(errors[0].Error()).To(ContainSubstring("duplicate KPI ID"))
				Expect(errors[0].Error()).To(ContainSubstring("cpu-usage"))
			})

			It("should detect multiple duplicates", func() {
				kpis := KPIs{
					Queries: []Query{
						{ID: "kpi-a", PromQuery: "up"},
						{ID: "kpi-b", PromQuery: "up"},
						{ID: "kpi-a", PromQuery: "up"},
						{ID: "kpi-b", PromQuery: "up"},
					},
				}

				errors := ValidateKPIs(kpis)

				Expect(errors).To(HaveLen(2))
			})
		})

		Context("when KPIs have empty fields", func() {
			It("should return an error for empty KPI ID", func() {
				kpis := KPIs{
					Queries: []Query{
						{ID: "", PromQuery: "up"},
					},
				}

				errors := ValidateKPIs(kpis)

				Expect(errors).To(HaveLen(1))
				Expect(errors[0].Error()).To(ContainSubstring("empty ID"))
			})

			It("should return an error for whitespace-only KPI ID", func() {
				kpis := KPIs{
					Queries: []Query{
						{ID: "   ", PromQuery: "up"},
					},
				}

				errors := ValidateKPIs(kpis)

				Expect(errors).To(HaveLen(1))
				Expect(errors[0].Error()).To(ContainSubstring("empty ID"))
			})

			It("should return an error for empty PromQL query", func() {
				kpis := KPIs{
					Queries: []Query{
						{ID: "empty-query", PromQuery: ""},
					},
				}

				errors := ValidateKPIs(kpis)

				Expect(errors).To(HaveLen(1))
				Expect(errors[0].Error()).To(ContainSubstring("empty-query"))
				Expect(errors[0].Error()).To(ContainSubstring("empty PromQL query"))
			})

			It("should return an error for whitespace-only PromQL query", func() {
				kpis := KPIs{
					Queries: []Query{
						{ID: "whitespace-query", PromQuery: "   "},
					},
				}

				errors := ValidateKPIs(kpis)

				Expect(errors).To(HaveLen(1))
				Expect(errors[0].Error()).To(ContainSubstring("empty PromQL query"))
			})
		})

		Context("when KPIs have multiple validation issues", func() {
			It("should return all errors", func() {
				kpis := KPIs{
					Queries: []Query{
						{ID: "valid", PromQuery: "up"},
						{ID: "broken", PromQuery: "rate(metric[5m]"},
						{ID: "valid", PromQuery: "up"},
						{ID: "empty-query", PromQuery: ""},
					},
				}

				errors := ValidateKPIs(kpis)

				Expect(errors).To(HaveLen(3))
			})
		})
	})
})
