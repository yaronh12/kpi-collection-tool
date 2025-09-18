package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

const (
	OUTPUT_FILE = "kpi-output.json"
	// CPU_USAGE_COMMAND          = "oc rsh -n openshift-monitoring prometheus-k8s-0 curl -ks 'http://localhost:9090/api/v1/query' --data-urlencode 'query=sort_desc(rate(container_cpu_usage_seconds_total{id=~\"/system.slice/.*\"}[30m]))'"
	// OVS_CPU_USAGE_COMMAND      = "oc rsh -n openshift-monitoring prometheus-k8s-0 curl -ks 'http://localhost:9090/api/v1/query' --data-urlencode 'query=sort_desc((rate(container_cpu_usage_seconds_total{id=~\"/ovs.slice/.*\"}[30m])))'"
	// POD_CPU_USAGE_COMMAND      = "oc rsh -n openshift-monitoring prometheus-k8s-0 curl -ks 'http://localhost:9090/api/v1/query' --data-urlencode 'query=sort_desc(avg_over_time(pod:container_cpu_usage:sum[30m]))'"
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

type KPIs struct {
	Queries []struct {
		ID        string `json:"id"`
		PromQuery string `json:"promquery"`
	} `json:"queries"`
}

func main() {
	fmt.Println("RDS KPI Collector starting...")

	// Add all the command from Jose Nunez here
	// commandsToRun := []string{
	// 	CPU_USAGE_COMMAND,
	// 	OVS_CPU_USAGE_COMMAND,
	// 	POD_CPU_USAGE_COMMAND,
	// }
	// Parse command-line flags for authentication and connection options

	// Option 1: --bearer-token and --thanos-url
	// Option 2: --kubeconfig

	var (
		bearerToken = flag.String("token", "", "bearer token for thanos-queries")
		thanosURL   = flag.String("thanos-url", "", "thanos url for http requests")
		kubeconfig  = flag.String("kubeconfig", "", "kubeconfig file path")
	)

	flag.Parse()

	fmt.Println("bearer token: ", *bearerToken)
	fmt.Println("thanos url: ", *thanosURL)
	fmt.Println("kubeconfig: ", *kubeconfig)

	// Open and read the kpis.json file
	kpisFile, err := os.Open("kpis.json")
	if err != nil {
		fmt.Printf("Failed to open kpis.json: %v\n", err)
		return
	}
	defer kpisFile.Close()

	// Decode the kpis.json file into an array of KPIs
	var kpis KPIs
	decoder := json.NewDecoder(kpisFile)
	if err := decoder.Decode(&kpis); err != nil {
		fmt.Printf("Failed to decode kpis.json: %v\n", err)
		return
	}

	// Print the queries from the predefined file
	var queries []string
	fmt.Printf("kpis queries as Go struct:\n")
	for _, query := range kpis.Queries {
		queries = append(queries, query.PromQuery)
	}
	fmt.Println(queries)
	// We run the commands
	queriesResults, err := runQueries(queries)
	if err != nil {
		fmt.Printf("Failed to run commands: %v\n", err)
		return
	}

	fmt.Println(queriesResults)

	// // We save the results to a file
	// err = saveToFile(commandsResults, OUTPUT_FILE)
	// if err != nil {
	// 	fmt.Printf("Failed to save file: %v\n", err)
	// 	return
	// }

	// fmt.Printf("Done! Check %s\n", OUTPUT_FILE)

}

func runQueries(queriesToRun []string) (map[string]PrometheusResponse, error) {
	// this is a map: key is the query, output of the query is the value
	queriesResults := make(map[string]PrometheusResponse)

	for _, command := range queriesToRun {
		fmt.Printf("Running: %s\n", command)
		// We run the command and get the output

	}

	return queriesResults, nil
}

func saveToFile(data map[string]PrometheusResponse, filename string) error {
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
