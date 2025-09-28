package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

const (
	OUTPUT_FILE                = "kpi-output.json"
	CPU_USAGE_COMMAND          = "oc rsh -n openshift-monitoring prometheus-k8s-0 curl -ks 'http://localhost:9090/api/v1/query' --data-urlencode 'query=sort_desc(rate(container_cpu_usage_seconds_total{id=~\"/system.slice/.*\"}[30m]))'"
	OVS_CPU_USAGE_COMMAND      = "oc rsh -n openshift-monitoring prometheus-k8s-0 curl -ks 'http://localhost:9090/api/v1/query' --data-urlencode 'query=sort_desc((rate(container_cpu_usage_seconds_total{id=~\"/ovs.slice/.*\"}[30m])))'"
	POD_CPU_USAGE_COMMAND      = "oc rsh -n openshift-monitoring prometheus-k8s-0 curl -ks 'http://localhost:9090/api/v1/query' --data-urlencode 'query=sort_desc(avg_over_time(pod:container_cpu_usage:sum[30m]))'"
	USER_READ_WRITE_PERMISSION = 0644
)

type PrometheusResponse struct {
	Data struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  []any             `json:"value"`
		} `json:"result"`
		ResultType string `json:"resultType"`
	} `json:"data"`
	Status string `json:"status"`
}

type QueryResponse struct {
	PrometheusResponse PrometheusResponse `json:"prometheus_response"`
	ErrorMsg           string             `json:"error"`
}

func main() {
	fmt.Println("RDS KPI Collector starting...")

	// Add all the command from Jose Nunez here
	commandsToRun := []string{
		CPU_USAGE_COMMAND,
		OVS_CPU_USAGE_COMMAND,
		POD_CPU_USAGE_COMMAND,
	}

	// We run the commands
	commandsResults := runCommands(commandsToRun)

	// We save the results to a file
	err := saveToFile(commandsResults, OUTPUT_FILE)
	if err != nil {
		fmt.Printf("Failed to save file: %v\n", err)
		return
	}

	fmt.Printf("Done! Check %s\n", OUTPUT_FILE)

}

func saveToFile(data map[string]QueryResponse, filename string) error {
	// We create the JSON file
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create JSON: %v", err)
	}

	// We save the file
	err = os.WriteFile(filename, file, USER_READ_WRITE_PERMISSION)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}
	return nil
}

func runCommands(commandsToRun []string) map[string]QueryResponse {
	// this is a map: key is the command, output of the command is the value (jsonData)
	commandsResults := make(map[string]QueryResponse)

	for _, command := range commandsToRun {
		fmt.Printf("Running: %s\n", command)
		// We run the command and get the output
		output, err := exec.Command("sh", "-c", command).Output()
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			commandsResults[command] = QueryResponse{ErrorMsg: fmt.Sprintf("command execution failed: %v", err)}
			continue
		}

		// We parse the output to a JSON go struct
		var jsonData PrometheusResponse
		if err := json.Unmarshal(output, &jsonData); err != nil {
			fmt.Printf("JSON parse failed: %v\n", err)
			commandsResults[command] = QueryResponse{
				PrometheusResponse: jsonData,
				ErrorMsg:           fmt.Sprintf("JSON parsing failed: %v", err),
			}
			continue
		}

		//We save the data in the map
		commandsResults[command] = QueryResponse{PrometheusResponse: jsonData}

		// We print the JSON data to the console
		prettyJSON, _ := json.MarshalIndent(jsonData, "", "  ")
		fmt.Printf("JSON data: %s\n", prettyJSON)

		fmt.Printf("Success\n")
	}
	return commandsResults
}
