package collector

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"rds-kpi-collector/internal/config"
)

var _ = Describe("Collector", func() {
	Describe("groupKPIsByFrequency", func() {
		var (
			kpis        config.KPIs
			defaultFreq int
		)

		BeforeEach(func() {
			defaultFreq = 60
		})

		Context("when all KPIs have default frequency", func() {
			BeforeEach(func() {
				kpis = config.KPIs{
					Queries: []config.Query{
						{ID: "kpi-1", PromQuery: "query1"},
						{ID: "kpi-2", PromQuery: "query2"},
					},
				}
			})

			It("should group all KPIs under default frequency", func() {
				grouped := groupKPIsByFrequency(kpis, defaultFreq)

				Expect(grouped).To(HaveLen(1))
				Expect(grouped[60].Queries).To(HaveLen(2))
				Expect(grouped[60].Queries[0].ID).To(Equal("kpi-1"))
				Expect(grouped[60].Queries[1].ID).To(Equal("kpi-2"))
			})
		})

		Context("when all KPIs have custom frequency", func() {
			BeforeEach(func() {
				freq1 := 10
				freq2 := 30
				kpis = config.KPIs{
					Queries: []config.Query{
						{ID: "kpi-1", PromQuery: "query1", SampleFrequency: &freq1},
						{ID: "kpi-2", PromQuery: "query2", SampleFrequency: &freq2},
					},
				}
			})

			It("should group KPIs by their custom frequencies", func() {
				grouped := groupKPIsByFrequency(kpis, defaultFreq)

				Expect(grouped).To(HaveLen(2))
				Expect(grouped[10].Queries).To(HaveLen(1))
				Expect(grouped[10].Queries[0].ID).To(Equal("kpi-1"))
				Expect(grouped[30].Queries).To(HaveLen(1))
				Expect(grouped[30].Queries[0].ID).To(Equal("kpi-2"))
			})
		})

		Context("when KPIs have mixed frequencies", func() {
			BeforeEach(func() {
				freq := 15
				kpis = config.KPIs{
					Queries: []config.Query{
						{ID: "default-1", PromQuery: "query1"},
						{ID: "custom-1", PromQuery: "query2", SampleFrequency: &freq},
						{ID: "default-2", PromQuery: "query3"},
					},
				}
			})

			It("should group KPIs correctly by frequency", func() {
				grouped := groupKPIsByFrequency(kpis, defaultFreq)

				Expect(grouped).To(HaveLen(2))
				Expect(grouped[60].Queries).To(HaveLen(2))
				Expect(grouped[60].Queries[0].ID).To(Equal("default-1"))
				Expect(grouped[60].Queries[1].ID).To(Equal("default-2"))
				Expect(grouped[15].Queries).To(HaveLen(1))
				Expect(grouped[15].Queries[0].ID).To(Equal("custom-1"))
			})
		})

		Context("when there are no KPIs", func() {
			BeforeEach(func() {
				kpis = config.KPIs{Queries: []config.Query{}}
			})

			It("should return empty map", func() {
				grouped := groupKPIsByFrequency(kpis, defaultFreq)

				Expect(grouped).To(BeEmpty())
			})
		})

		Context("when multiple KPIs share the same custom frequency", func() {
			BeforeEach(func() {
				freq := 30
				kpis = config.KPIs{
					Queries: []config.Query{
						{ID: "kpi-1", PromQuery: "query1", SampleFrequency: &freq},
						{ID: "kpi-2", PromQuery: "query2", SampleFrequency: &freq},
						{ID: "kpi-3", PromQuery: "query3", SampleFrequency: &freq},
					},
				}
			})

			It("should group all KPIs under the same frequency", func() {
				grouped := groupKPIsByFrequency(kpis, defaultFreq)

				Expect(grouped).To(HaveLen(1))
				Expect(grouped[30].Queries).To(HaveLen(3))
				Expect(grouped[30].Queries[0].ID).To(Equal("kpi-1"))
				Expect(grouped[30].Queries[1].ID).To(Equal("kpi-2"))
				Expect(grouped[30].Queries[2].ID).To(Equal("kpi-3"))
			})
		})
	})

	Describe("Query.GetEffectiveFrequency", func() {
		var defaultFreq int

		BeforeEach(func() {
			defaultFreq = 60
		})

		Context("when query has custom frequency", func() {
			It("should return the custom frequency", func() {
				customFreq := 30
				query := config.Query{
					ID:              "test-kpi",
					PromQuery:       "test_query",
					SampleFrequency: &customFreq,
				}

				effectiveFreq := query.GetEffectiveFrequency(defaultFreq)
				Expect(effectiveFreq).To(Equal(30))
			})
		})

		Context("when query has no custom frequency", func() {
			It("should return the default frequency", func() {
				query := config.Query{
					ID:              "test-kpi",
					PromQuery:       "test_query",
					SampleFrequency: nil,
				}

				effectiveFreq := query.GetEffectiveFrequency(defaultFreq)
				Expect(effectiveFreq).To(Equal(60))
			})
		})

		Context("when query has zero custom frequency", func() {
			It("should return the default frequency", func() {
				zeroFreq := 0
				query := config.Query{
					ID:              "test-kpi",
					PromQuery:       "test_query",
					SampleFrequency: &zeroFreq,
				}

				effectiveFreq := query.GetEffectiveFrequency(defaultFreq)
				Expect(effectiveFreq).To(Equal(60))
			})
		})

		Context("when query has negative custom frequency", func() {
			It("should return the default frequency", func() {
				negativeFreq := -10
				query := config.Query{
					ID:              "test-kpi",
					PromQuery:       "test_query",
					SampleFrequency: &negativeFreq,
				}

				effectiveFreq := query.GetEffectiveFrequency(defaultFreq)
				Expect(effectiveFreq).To(Equal(60))
			})
		})
	})

	Describe("startKPIGoroutines", func() {
		var flags config.InputFlags

		BeforeEach(func() {
			flags = config.InputFlags{
				SamplingFreq: 60,
				Duration:     1 * time.Second,
			}
		})

		Context("when there are KPIs with various frequencies", func() {
			It("should return cancel function and WaitGroup", func() {
				freq := 5
				kpis := config.KPIs{
					Queries: []config.Query{
						{ID: "kpi-1", PromQuery: "query1", SampleFrequency: &freq},
						{ID: "kpi-2", PromQuery: "query2"}, // default frequency
					},
				}

				cancel, wg := startKPIGoroutines(kpis, flags)

				Expect(cancel).NotTo(BeNil())
				Expect(wg).NotTo(BeNil())

				// Cancel and wait to clean up goroutines
				cancel()
				wg.Wait()
			})
		})

		Context("when there are only default frequency KPIs", func() {
			It("should start goroutines for default frequency", func() {
				kpis := config.KPIs{
					Queries: []config.Query{
						{ID: "kpi-1", PromQuery: "query1"},
						{ID: "kpi-2", PromQuery: "query2"},
					},
				}

				cancel, wg := startKPIGoroutines(kpis, flags)

				Expect(cancel).NotTo(BeNil())
				Expect(wg).NotTo(BeNil())

				// Cancel and wait to clean up goroutines
				cancel()
				wg.Wait()
			})
		})

		Context("when there are no KPIs", func() {
			It("should return cancel function and empty WaitGroup", func() {
				kpis := config.KPIs{Queries: []config.Query{}}

				cancel, wg := startKPIGoroutines(kpis, flags)

				Expect(cancel).NotTo(BeNil())
				Expect(wg).NotTo(BeNil())

				// Cancel and wait (should return immediately)
				cancel()
				wg.Wait()
			})
		})
	})
})
