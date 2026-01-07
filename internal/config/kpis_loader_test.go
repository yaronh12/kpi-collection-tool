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
})
