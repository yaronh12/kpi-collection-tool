package collector

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"kpi-collector/internal/config"
)

var _ = Describe("Collector", func() {
	Describe("separateByFrequency", func() {
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

			It("should put all KPIs in default frequency group", func() {
				defaultKPIs, customKPIs := separateByFrequency(kpis, defaultFreq)

				Expect(defaultKPIs.Queries).To(HaveLen(2))
				Expect(customKPIs.Queries).To(BeEmpty())
				Expect(defaultKPIs.Queries[0].ID).To(Equal("kpi-1"))
				Expect(defaultKPIs.Queries[1].ID).To(Equal("kpi-2"))
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

			It("should put all KPIs in custom frequency group", func() {
				defaultKPIs, customKPIs := separateByFrequency(kpis, defaultFreq)

				Expect(defaultKPIs.Queries).To(BeEmpty())
				Expect(customKPIs.Queries).To(HaveLen(2))
				Expect(customKPIs.Queries[0].ID).To(Equal("kpi-1"))
				Expect(customKPIs.Queries[1].ID).To(Equal("kpi-2"))
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

			It("should separate KPIs correctly", func() {
				defaultKPIs, customKPIs := separateByFrequency(kpis, defaultFreq)

				Expect(defaultKPIs.Queries).To(HaveLen(2))
				Expect(customKPIs.Queries).To(HaveLen(1))
				Expect(defaultKPIs.Queries[0].ID).To(Equal("default-1"))
				Expect(defaultKPIs.Queries[1].ID).To(Equal("default-2"))
				Expect(customKPIs.Queries[0].ID).To(Equal("custom-1"))
			})
		})

		Context("when there are no KPIs", func() {
			BeforeEach(func() {
				kpis = config.KPIs{Queries: []config.Query{}}
			})

			It("should return empty groups", func() {
				defaultKPIs, customKPIs := separateByFrequency(kpis, defaultFreq)

				Expect(defaultKPIs.Queries).To(BeEmpty())
				Expect(customKPIs.Queries).To(BeEmpty())
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

	Describe("setupDefaultKPITicker", func() {
		var flags config.InputFlags

		BeforeEach(func() {
			flags = config.InputFlags{
				SamplingFreq: 10,
				Duration:     30 * time.Second,
			}
		})

		Context("when there are default frequency KPIs", func() {
			It("should return a ticker channel and initial count of 1", func() {
				kpis := config.KPIs{
					Queries: []config.Query{
						{ID: "kpi-1", PromQuery: "query1"},
					},
				}

				tickerChan, sampleCount := setupDefaultKPITicker(kpis, flags)

				Expect(tickerChan).NotTo(BeNil())
				Expect(sampleCount).To(Equal(1))
			})
		})

		Context("when there are no default frequency KPIs", func() {
			It("should return a dummy channel and count of 0", func() {
				kpis := config.KPIs{Queries: []config.Query{}}

				tickerChan, sampleCount := setupDefaultKPITicker(kpis, flags)

				Expect(tickerChan).NotTo(BeNil())
				Expect(sampleCount).To(Equal(0))
			})
		})
	})

	Describe("startCustomFrequencyGoroutines", func() {
		var flags config.InputFlags

		BeforeEach(func() {
			flags = config.InputFlags{
				SamplingFreq: 60,
				Duration:     1 * time.Second,
			}
		})

		Context("when there are custom frequency KPIs", func() {
			It("should return cancel function and WaitGroup", func() {
				freq := 5
				kpis := config.KPIs{
					Queries: []config.Query{
						{ID: "kpi-1", PromQuery: "query1", SampleFrequency: &freq},
					},
				}

				cancel, wg := startCustomFrequencyGoroutines(kpis, flags)

				Expect(cancel).NotTo(BeNil())
				Expect(wg).NotTo(BeNil())

				// Cancel and wait to clean up goroutines
				cancel()
				wg.Wait()
			})
		})

		Context("when there are no custom frequency KPIs", func() {
			It("should return cancel function and empty WaitGroup", func() {
				kpis := config.KPIs{Queries: []config.Query{}}

				cancel, wg := startCustomFrequencyGoroutines(kpis, flags)

				Expect(cancel).NotTo(BeNil())
				Expect(wg).NotTo(BeNil())

				// Cancel and wait (should return immediately)
				cancel()
				wg.Wait()
			})
		})
	})
})
