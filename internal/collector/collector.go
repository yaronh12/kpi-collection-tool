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

	// Start all KPI goroutines grouped by frequency
	cancel, wg := startKPIGoroutines(kpis, flags)
	defer cancel()

	// Main goroutine only handles duration timer and interrupts
	select {
	case <-durationTimer.C:
		log.Printf("Duration timer expired")
		shutdown(cancel, wg)
		return

	case <-interruptChan:
		log.Printf("Program interrupted")
		shutdown(cancel, wg)
		return
	}
}

// groupKPIsByFrequency groups KPIs by their effective sampling frequency
// This includes both default and custom frequency KPIs
func groupKPIsByFrequency(kpis config.KPIs, defaultFreq int) map[int]config.KPIs {
	kpisByFreq := make(map[int]config.KPIs)

	for _, kpi := range kpis.Queries {
		effectiveFreq := kpi.GetEffectiveFrequency(defaultFreq)

		group := kpisByFreq[effectiveFreq]
		group.Queries = append(group.Queries, kpi)
		kpisByFreq[effectiveFreq] = group
	}

	if len(kpisByFreq) > 0 {
		log.Printf("Grouped all KPIs into %d unique frequency groups", len(kpisByFreq))
		for freq, group := range kpisByFreq {
			log.Printf("  Frequency %ds: %d KPIs", freq, len(group.Queries))
		}
	}

	return kpisByFreq
}

// startKPIGoroutines starts one goroutine per unique frequency for all KPIs
func startKPIGoroutines(kpis config.KPIs, flags config.InputFlags) (context.CancelFunc, *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Group ALL KPIs by their sampling frequency (including default frequency)
	kpisByFreq := groupKPIsByFrequency(kpis, flags.SamplingFreq)

	// Start one goroutine per unique frequency
	for freq, kpisForFreq := range kpisByFreq {
		wg.Add(1)
		go func(frequency int, kpiGroup config.KPIs) {
			defer wg.Done()
			runKPIGroupLoop(ctx, kpiGroup, frequency, flags)
		}(freq, kpisForFreq)
	}

	return cancel, &wg
}

func shutdown(cancel context.CancelFunc, wg *sync.WaitGroup) {
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
