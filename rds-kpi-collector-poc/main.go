package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("RDS KPI Collector starting...")

	// Add all the command from Jose Nunez here
	commandsToRun := []string{
		"oc rsh -n openshift-monitoring prometheus-k8s-0 curl -ks 'http://localhost:9090/api/v1/query' --data-urlencode 'query=sort_desc(rate(container_cpu_usage_seconds_total{id=~\"/system.slice/.*\"}[30m]))' | jq .",
	}

	// this is a map: key is the command, output of the command is the value (jsonData)
	commandsResults := make(map[string]interface{})

	for _, command := range commandsToRun {
		fmt.Printf("Running: %s\n", command)
		// We run the command and get the output
		output, err := exec.Command("sh", "-c", command).Output()
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			commandsResults[command] = map[string]string{"error": err.Error()}
			continue
		}

		// We parse the output to a JSON go struct
		var jsonData interface{}
		if err := json.Unmarshal(output, &jsonData); err != nil {
			fmt.Printf("JSON parse failed: %v\n", err)
			commandsResults[command] = map[string]string{"error": "JSON parse failed"}
			continue
		}

		commandsResults[command] = jsonData // Direct assignment to the map

		fmt.Printf("Success\n")
	}

	// We save the results to a file
	file, err := json.MarshalIndent(commandsResults, "", "  ")
	if err != nil {
		fmt.Printf("Failed to create JSON: %v\n", err)
		return
	}

	err = os.WriteFile("kpi-output.json", file, 0644)
	if err != nil {
		fmt.Printf("Failed to save file: %v\n", err)
		return
	}

	fmt.Println("Done! Check kpi-output.json")

}
