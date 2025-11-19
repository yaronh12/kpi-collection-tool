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
			runKPIs(defaultFreqKPIs, flags, defaultSampleCount, flags.SamplingFreq)
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
		effectiveFreq := kpi.GetEffectiveFrequency(defaultFreq)

		if effectiveFreq == defaultFreq {
			defaultFreqKPIs.Queries = append(defaultFreqKPIs.Queries, kpi)
		} else {
			customFreqKPIs.Queries = append(customFreqKPIs.Queries, kpi)
		}
	}

	log.Printf("KPIs with default frequency (%ds): %d", defaultFreq, len(defaultFreqKPIs.Queries))
	log.Printf("KPIs with custom frequency: %d", len(customFreqKPIs.Queries))

	return defaultFreqKPIs, customFreqKPIs
}

// groupKPIsByFrequency groups KPIs by their effective sampling frequency
// Excludes KPIs that use the default frequency (those are handled by the main goroutine)
func groupKPIsByFrequency(kpis config.KPIs, defaultFreq int) map[int]config.KPIs {
	// key - frequency, value - kpi
	kpisByFreq := make(map[int]config.KPIs)

	for _, kpi := range kpis.Queries {
		effectiveFreq := kpi.GetEffectiveFrequency(defaultFreq)

		// Skip KPIs with default frequency - they're handled by the main goroutine
		if effectiveFreq == defaultFreq {
			continue
		}

		group := kpisByFreq[effectiveFreq]
		group.Queries = append(group.Queries, kpi)
		kpisByFreq[effectiveFreq] = group
	}

	if len(kpisByFreq) > 0 {
		log.Printf("Grouped custom frequency KPIs into %d unique frequency groups", len(kpisByFreq))
		for freq, group := range kpisByFreq {
			log.Printf("  Frequency %ds: %d KPIs", freq, len(group.Queries))
		}
	}

	return kpisByFreq
}

func startCustomFrequencyGoroutines(kpis config.KPIs, flags config.InputFlags) (context.CancelFunc, *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	defaultFrequency := flags.SamplingFreq
	// Group KPIs by their sampling frequency (excluding default frequency)
	kpisByFreq := groupKPIsByFrequency(kpis, defaultFrequency)

	// Start one goroutine per unique frequency (only for non-default frequencies)
	for freq, kpisForFreq := range kpisByFreq {
		wg.Add(1)
		go func(frequency int, kpiGroup config.KPIs) {
			defer wg.Done()
			runKPIGroupLoop(ctx, kpiGroup, frequency, flags)
		}(freq, kpisForFreq)
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
		runKPIs(kpis, flags, sampleCount, flags.SamplingFreq)
	} else {
		// Return a channel that never sends (blocks forever)
		// This is safe because it's used in a select statement with other cases
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

// runKPIGroupLoop runs a group of KPIs that share the same sampling frequency
func runKPIGroupLoop(ctx context.Context, kpis config.KPIs, frequency int, flags config.InputFlags) {
	ticker := time.NewTicker(time.Duration(frequency) * time.Second)
	defer ticker.Stop()

	sampleCount := 0
	log.Printf("Starting goroutine for %d KPIs with frequency %ds", len(kpis.Queries), frequency)

	// Run immediately on start
	sampleCount++
	runKPIs(kpis, flags, sampleCount, frequency)

	for {
		select {
		case <-ticker.C:
			sampleCount++
			runKPIs(kpis, flags, sampleCount, frequency)

		case <-ctx.Done():
			log.Printf("KPI group (frequency %ds) stopped after %d samples", frequency, sampleCount)
			return
		}
	}
}

// runKPIs executes a group of KPIs and logs the results
func runKPIs(kpis config.KPIs, flags config.InputFlags, sampleCount int, frequency int) {
	if len(kpis.Queries) == 0 {
		return
	}

	log.Printf("Running sample %d for %d KPIs with frequency %ds", sampleCount, len(kpis.Queries), frequency)

	if err := prometheus.RunQueries(kpis, flags); err != nil {
		log.Printf("RunQueries failed for frequency %ds KPIs: %v", frequency, err)
	} else {
		log.Printf("Sample %d for frequency %ds KPIs completed successfully", sampleCount, frequency)
	}
}
