package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

// ExecuteCommand runs the given shell command and parses its output as JSON.
// Returns the parsed JSON or an error if the command fails or output is invalid.
func ExecuteCommand(command string) (interface{}, error) {
	log.Printf("Executing command: %s\n", command)
	output, err := exec.Command("sh", "-c", command).Output()
	if err != nil {
		return nil, fmt.Errorf("command failed: %w", err)
	}

	var jsonData interface{}
	if err := json.Unmarshal(output, &jsonData); err != nil {
		return nil, fmt.Errorf("JSON parse failed: %w", err)
	}

	return jsonData, nil
}

// SaveJSON writes the given data map to a JSON file with indentation.
// Returns an error if marshaling or writing the file fails.
func SaveJSON(data map[string]interface{}, fileName string) error {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(fileName, file, 0644); err != nil {
		return fmt.Errorf("failed to save file '%s': %w", fileName, err)
	}

	log.Printf("Results saved to %s\n", fileName)
	return nil
}

// RunKPICollector runs the main collection loop for KPIs.
// samplingFrequency is in seconds, duration specifies total collection time,
// and outputFile is the path where results are saved.
func RunKPICollector(samplingFrequency int, duration time.Duration, outputFile string) error {
	log.Printf("Collector started: frequency=%ds, duration=%s, output=%s\n",
		samplingFrequency, duration, outputFile)
	fmt.Printf("Starting KPI collector with sampling frequency %d seconds and duration %s\n", samplingFrequency, duration)

	startTime := time.Now()
	commandsResults := make(map[string]interface{})
	sampleNumber := 1

	for time.Since(startTime) < duration {
		log.Printf("Running sample %d\n", sampleNumber)
		fmt.Printf("Running sample %d\n", sampleNumber)

		// Build Prometheus query
		query := fmt.Sprintf("sort_desc(rate(container_cpu_usage_seconds_total{id=~\"/system.slice/.*\"}[%dm]))", samplingFrequency)
		command := fmt.Sprintf(
			"oc rsh -n openshift-monitoring prometheus-k8s-0 curl -ks 'http://localhost:9090/api/v1/query' --data-urlencode 'query=%s' | jq .",
			query,
		)

		// Execute the command and parse JSON output
		result, err := ExecuteCommand(command)
		if err != nil {
			log.Printf("Sample %d failed: %v\n", sampleNumber, err)
			commandsResults[fmt.Sprintf("sample_%d", sampleNumber)] = map[string]string{"error": err.Error()}
			fmt.Printf("Sample %d failed (see log for details)\n", sampleNumber)
		} else {
			log.Printf("Sample %d success\n", sampleNumber)
			commandsResults[fmt.Sprintf("sample_%d", sampleNumber)] = result
			fmt.Printf("Sample %d success\n", sampleNumber)
		}

		// Save results after each sample
		if err := SaveJSON(commandsResults, outputFile); err != nil {
			log.Printf("Failed to save JSON after sample %d: %v\n", sampleNumber, err)
			return err
		}

		sampleNumber++
		time.Sleep(time.Duration(samplingFrequency) * time.Second)
	}

	log.Printf("Collector finished successfully. Output: %s\n", outputFile)
	fmt.Printf("Done! Check %s\n", outputFile)
	return nil
}
