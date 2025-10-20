package main

import (
	"flag"
	"fmt"
	"kpi-collection-tool/collector"
	"kpi-collection-tool/logger"
	"log"
	"os"
	"time"
)

func main() {
	fmt.Println("RDS KPI Collector starting...")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	samplingFrequency := flag.Int("frequency", 60, "Sampling frequency in seconds")
	duration := flag.Duration("duration", 45*time.Minute, "Total duration for sampling (e.g. 10s, 1m, 2h)")
	outputFile := flag.String("output", "kpi-output.json", "Output file name for results")
	logFile := flag.String("log", "kpi.log", "Log file name")

	flag.Parse()

	if *samplingFrequency <= 0 {
		fmt.Println("Error: frequency must be greater than 0")
		os.Exit(1)
	}
	if *duration <= 0 {
		fmt.Println("Error: duration must be greater than 0")
		os.Exit(1)
	}

	// Initialize logger
	logF, err := logger.InitLogger(*logFile)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := logF.Close(); err != nil {
			fmt.Printf("Failed to close log file: %v\n", err)
		}
	}()

	log.Println("RDS KPI Collector initialized.")

	// Run collector
	if err := collector.RunKPICollector(*samplingFrequency, *duration, *outputFile); err != nil {
		log.Printf("Collector error: %v\n", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	log.Println("RDS KPI Collector finished successfully.")
	fmt.Println("RDS KPI Collector finished successfully.")
}
