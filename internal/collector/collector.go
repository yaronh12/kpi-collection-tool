package collector

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"rds-kpi-collector/internal/config"
	"rds-kpi-collector/internal/prometheus"
)

const HUNDRED_MILLIS_DURATION_BUFFER = 100 * time.Millisecond

// Run executes the KPI collection loop
func Run(kpis config.KPIs, flags config.InputFlags) {
	durationTimer := time.NewTimer(flags.Duration + HUNDRED_MILLIS_DURATION_BUFFER)
	defer durationTimer.Stop()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	log.Printf("Running for %s, deadline time: %s",
		flags.Duration.String(),
		time.Now().Add(flags.Duration).Format(time.RFC3339))

	defaultFreqKPIs, customFreqKPIs := separateByFrequency(kpis, flags.SamplingFreq)

	cancel, wg := startCustomFrequencyGoroutines(customFreqKPIs, flags)
	defer cancel()

	tickerChan, defaultSampleCount := setupDefaultKPITicker(defaultFreqKPIs, flags)

	for {
		select {
		case <-tickerChan:
			defaultSampleCount++
			runDefaultKPIs(defaultFreqKPIs, flags, defaultSampleCount)

		case <-durationTimer.C:
			log.Printf("Duration timer expired")
			shutdown(defaultFreqKPIs, defaultSampleCount, cancel, wg)
			return

		case <-interruptChan:
			log.Printf("Program interrupted")
			shutdown(defaultFreqKPIs, defaultSampleCount, cancel, wg)
			return
		}
	}
}

func separateByFrequency(kpis config.KPIs, defaultFreq int) (config.KPIs, config.KPIs) {
	var defaultFreqKPIs, customFreqKPIs config.KPIs

	for _, kpi := range kpis.Queries {
		if kpi.SampleFrequency != nil {
			customFreqKPIs.Queries = append(customFreqKPIs.Queries, kpi)
		} else {
			defaultFreqKPIs.Queries = append(defaultFreqKPIs.Queries, kpi)
		}
	}

	log.Printf("KPIs with default frequency (%ds): %d", defaultFreq, len(defaultFreqKPIs.Queries))
	log.Printf("KPIs with custom frequency: %d", len(customFreqKPIs.Queries))

	return defaultFreqKPIs, customFreqKPIs
}

func startCustomFrequencyGoroutines(kpis config.KPIs, flags config.InputFlags) (context.CancelFunc, *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	for _, kpi := range kpis.Queries {
		wg.Add(1)
		go func(k config.Query) {
			defer wg.Done()
			runKPILoop(ctx, k, flags)
		}(kpi)
	}

	return cancel, &wg
}

func setupDefaultKPITicker(kpis config.KPIs, flags config.InputFlags) (<-chan time.Time, int) {
	var tickerChan <-chan time.Time
	sampleCount := 0

	if len(kpis.Queries) > 0 {
		ticker := time.NewTicker(time.Duration(flags.SamplingFreq) * time.Second)
		tickerChan = ticker.C

		log.Printf("Starting main loop for %d KPIs with default frequency", len(kpis.Queries))

		sampleCount++
		runDefaultKPIs(kpis, flags, sampleCount)
	} else {
		tickerChan = make(<-chan time.Time)
	}

	return tickerChan, sampleCount
}

func shutdown(kpis config.KPIs, sampleCount int, cancel context.CancelFunc, wg *sync.WaitGroup) {
	if len(kpis.Queries) > 0 {
		log.Printf("Default frequency KPIs: %d samples collected", sampleCount)
	}
	cancel()
	wg.Wait()
}

func runKPILoop(ctx context.Context, kpi config.Query, flags config.InputFlags) {
	effectiveFreq := kpi.GetEffectiveFrequency(flags.SamplingFreq)
	ticker := time.NewTicker(time.Duration(effectiveFreq) * time.Second)
	defer ticker.Stop()

	sampleCount := 0
	log.Printf("Starting dedicated goroutine for KPI '%s' with custom frequency %ds", kpi.ID, effectiveFreq)

	sampleCount++
	runSingleKPI(kpi, flags, sampleCount)

	for {
		select {
		case <-ticker.C:
			sampleCount++
			runSingleKPI(kpi, flags, sampleCount)

		case <-ctx.Done():
			log.Printf("KPI '%s' stopped after %d samples", kpi.ID, sampleCount)
			return
		}
	}
}

func runDefaultKPIs(kpis config.KPIs, flags config.InputFlags, sampleCount int) {
	if len(kpis.Queries) == 0 {
		return
	}

	log.Printf("Running sample %d for default frequency KPIs", sampleCount)

	if err := prometheus.RunQueries(kpis, flags); err != nil {
		log.Printf("RunQueries failed for default frequency KPIs: %v", err)
	} else {
		log.Printf("Sample %d for default frequency KPIs completed successfully", sampleCount)
	}
}

func runSingleKPI(kpi config.Query, flags config.InputFlags, sampleCount int) {
	log.Printf("Running sample %d for KPI: %s", sampleCount, kpi.ID)

	singleKPI := config.KPIs{
		Queries: []config.Query{kpi},
	}

	if err := prometheus.RunQueries(singleKPI, flags); err != nil {
		log.Printf("RunQueries failed for KPI %s: %v", kpi.ID, err)
	} else {
		log.Printf("Sample %d for KPI %s completed successfully", sampleCount, kpi.ID)
	}
}
