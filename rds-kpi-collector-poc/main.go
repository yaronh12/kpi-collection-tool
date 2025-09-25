package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
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

type QueryResponse struct {
	PrometheusResponse PrometheusResponse `json:"prometheus_response"`
	ErrorMsg           string             `json:"error"`
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
		bearerToken string
		thanosURL   string
		kubeconfig  string
	)

	flag.StringVar(&bearerToken, "token", "", "bearer token for thanos-queries")
	flag.StringVar(&thanosURL, "thanos-url", "", "thanos url for http requests")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig file path")
	// validateFlags checks that either both token and thanos-url are provided,
	// or kubeconfig is provided, but not both scenarios

	flag.Parse()

	fmt.Println("bearer token: ", bearerToken)
	fmt.Println("thanos url: ", thanosURL)
	fmt.Println("kubeconfig: ", kubeconfig)

	// Validate that either both token and thanos-url are provided,
	// or kubeconfig is provided, but not both scenarios
	err := validateFlags(bearerToken, thanosURL, kubeconfig)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Open and read the kpis.json file
	// kspis.json file contains predefined KPI queries with their IDs and PromQL queries
	// that will be executed against the Prometheus/Thanos endpoint
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

	// Extract PromQL queries from the KPIs structure into a slice of strings
	// This prepares the queries to be executed against the Prometheus/Thanos endpoint
	var queries []string
	for _, query := range kpis.Queries {
		queries = append(queries, query.PromQuery)
	}

	// We run the queries
	queriesResults, err := runQueries(queries, thanosURL, bearerToken)
	if err != nil {
		fmt.Printf("Failed to run commands: %v\n", err)
		return
	}

	fmt.Printf("\n\n\n\n%v", queriesResults)

	// // We save the results to a file
	// err = saveToFile(commandsResults, OUTPUT_FILE)
	// if err != nil {
	// 	fmt.Printf("Failed to save file: %v\n", err)
	// 	return
	// }

	// fmt.Printf("Done! Check %s\n", OUTPUT_FILE)

}

func validateFlags(token string, url string, kubeconfig string) error {
	if (token != "" && url != "" && kubeconfig == "") ||
		(token == "" && url == "" && kubeconfig != "") {
		return nil
	} else {
		return fmt.Errorf("invalid flag combination: either provide --token and --thanos-url, or provide --kubeconfig")
	}
}

func runQueries(queriesToRun []string, thanosURL string, bearerToken string) (map[string]QueryResponse, error) {

	// tr := &http.Transport{
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	// client := &http.Client{Transport: tr}

	// Create Prometheus client
	client, err := api.NewClient(api.Config{
		Address: "https://" + thanosURL, // Creation of the URL
		RoundTripper: &tokenRoundTripper{
			token: bearerToken, // Adding the bearer token to http header
			rt: &http.Transport{
				// NOTE: InsecureSkipVerify is set to true for development purposes only.
				// In production environments, this should be false and proper certificate
				// validation should be implemented. This is needed here because the local
				// development machine doesn't trust OpenShift's self-signed certificates.
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	// Create Prometheus v1 API client for executing queries
	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// this is a map: key is the query, output of the query is the value
	queriesResults := make(map[string]QueryResponse)

	for _, query := range queriesToRun {
		fmt.Println("------------------------")
		fmt.Printf("Running: %s\n", query)
		// Create POST request with form data
		// data := url.Values{}
		// data.Set("query", query)
		// req, err := http.NewRequest("POST", "https://"+(thanosURL)+"/api/v1/query", strings.NewReader(data.Encode()))
		// if err != nil {
		// 	continue
		// }

		// Execute query using the client library
		result, warnings, err := v1api.Query(ctx, query, time.Now())
		if err != nil {
			fmt.Println("query failed: ", err)
			queriesResults[query] = QueryResponse{ErrorMsg: fmt.Sprintf("query failed: %v", err)}
			continue
		}

		if len(warnings) > 0 {
			fmt.Printf("Warnings: %v\n", warnings)
		}

		fmt.Println("RESULT:", result)

		// // Add bearer token
		// req.Header.Set("Authorization", "Bearer "+(bearerToken))
		// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// resp, err := client.Do(req)
		// if err != nil {
		// 	fmt.Printf("Request failed: %v\n", err)
		// 	continue
		// }
		// defer resp.Body.Close()

		// // Read the response body
		// body, err := io.ReadAll(resp.Body)
		// if err != nil {
		// 	fmt.Printf("Failed to read response: %v\n", err)
		// 	continue
		// }

		// // Print the actual JSON
		// var prettyJSON interface{}
		// if err := json.Unmarshal(body, &prettyJSON); err == nil {
		// 	formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
		// 	fmt.Printf("Response JSON:\n%s\n", string(formatted))
		// } else {
		// 	fmt.Printf("Raw response: %s\n", string(body))
		// }
		//fmt.Println(resp)
		fmt.Println("------------------------")

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

// Custom RoundTripper to add Bearer token
type tokenRoundTripper struct {
	token string
	rt    http.RoundTripper
}

func (t *tokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.rt.RoundTrip(req)
}

//TOKEN = eyJhbGciOiJSUzI1NiIsImtpZCI6Im5mOTl3a05TN2dCQ19Zdl8yOExsc1hBUE9vR2pIR19fZnQxemY5RHk2Y0UifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjIl0sImV4cCI6MTc1ODIyNzM5NiwiaWF0IjoxNzU4MTkxMzk2LCJpc3MiOiJodHRwczovL2t1YmVybmV0ZXMuZGVmYXVsdC5zdmMiLCJqdGkiOiI4MjZkMWQyOS1lZTkwLTQ2Y2QtYjgxNy04ZDExOGY1OTY5ODUiLCJrdWJlcm5ldGVzLmlvIjp7Im5hbWVzcGFjZSI6Im9wZW5zaGlmdC1tb25pdG9yaW5nIiwic2VydmljZWFjY291bnQiOnsibmFtZSI6InRlbGVtZXRlci1jbGllbnQiLCJ1aWQiOiJiNDM3ODYzMS1mNWQ3LTQxMGYtYWI5OS1kOTM4ZDgyMGY3ZGQifX0sIm5iZiI6MTc1ODE5MTM5Niwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50Om9wZW5zaGlmdC1tb25pdG9yaW5nOnRlbGVtZXRlci1jbGllbnQifQ.HnFq-3x9MMuHrL7OAOZzsWUxhCNM1k9ULvJxMunMmVt8hWKwOd6O_gBNM6EYxDWvWFoqxPIq0ItpwoXY6oESMBv7BXMYCkiby5JvKJYVdqJmrkJuW59LPA8IRGWMBkQHRFu7yph8LMAKfOwzjfzm9Cwnwj3WzmZMkhFWJmvl1YxvOopTxHdWvLPddioCcLtXBqY6d0rwWyS2hCIFkAQt29kPuCT2IY1FYZvhhuA8J-Yhg8IuR52dmvWlS7tjFxW5O6WqpTv5RNZK2jHFPDCCp1SI70iqjOzkdjEleyPZdTbAaro3nPXf6BAK1a3y7WDnKtlV64NbZrhxUxadTan_en5p4TKfHpIuTSOiq6XqbF4w4TBhTHnb0aWClhch5LKFNTwViwmgabL7g0vJJPG05zNIHtKK5xfVC0Uls9MuBKIlBpdjh8d9MEK7us1qt49Tguby6OPxgQjsbb3riF95LpuOapM1lPzCQVnkP9NqChbPUEo0U0ox6bbmikObXya7-Z5_E1dGF4lEn6iphaQ-TerNLLCp0ZYo46S55bCt8zncWuBSRkEmdl1U18lo-6Mo2otjOybjQP49bj9FfCvJv_Uk_wTps7e8pkvUPgqJhAg11au1kMlWzUQYUu0p6ekzFp42DuP6wsErPAggh3RMJVsJcn3q1sMXl6qi5RgSxOs
//thanor querier url = thanos-querier-openshift-monitoring.apps.clus0.t5g.lab.eng.rdu2.redhat.com

// Command to get the token for telemeter-client:
// TOKEN=$(oc create token telemeter-client -n openshift-monitoring --duration=10h)

// Command to get to thanos URL:
//oc get route thanos-querier -n openshift-monitoring -o jsonpath='{.spec.host}'
