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

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	OUTPUT_FILE                = "kpi-output.json"
	USER_READ_WRITE_PERMISSION = 0644
	THANOS_ROUTE_API_PATH      = "/apis/route.openshift.io/v1/namespaces/openshift-monitoring/routes/thanos-querier"
)

// NOTE: Currently, QueryResponse and PrometheusResponse structs are not in good use
// because we are not saving any results to a JSON file yet. These are placeholder
// structures for future implementation when we add JSON output functionality.

// PrometheusResponse represents the JSON response structure returned by Prometheus/Thanos queries
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

// QueryResponse represents the result of executing a Prometheus query,
// containing both the response data and any error that occurred
type QueryResponse struct {
	PrometheusResponse PrometheusResponse `json:"prometheus_response"`
	ErrorMsg           string             `json:"error"`
}

// KPIs represents the structure of the kpis.json file containing
// the list of KPI queries to be executed against Prometheus/Thanos
type KPIs struct {
	Queries []struct {
		ID        string `json:"id"`
		PromQuery string `json:"promquery"`
	} `json:"queries"`
}

// Define a struct to represent the OpenShift route object
type Route struct {
	Spec struct {
		Host string `json:"host"`
	} `json:"spec"`
}

func main() {
	fmt.Println("RDS KPI Collector starting...")
	// Option 1: --bearer-token and --thanos-url
	// Option 2: --kubeconfig

	// Setup and validate flags
	bearerToken, thanosURL, kubeconfig, clusterName, err := setupFlags()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("bearer token: ", bearerToken)
	fmt.Println("thanos url: ", thanosURL)
	fmt.Println("kubeconfig: ", kubeconfig)
	fmt.Println("Cluster: ", clusterName)

	// Load KPI queries
	queries, err := loadKPIQueries()
	if err != nil {
		fmt.Printf("Failed to load KPI queries: %v\n", err)
		return
	}

	// If kubeconfig is provided, discover Thanos URL and create service account token
	if kubeconfig != "" {
		thanosURL, bearerToken, err = setupKubeconfigAuth(kubeconfig)
		if err != nil {
			fmt.Printf("Failed to setup kubeconfig auth: %v\n", err)
			return
		}
		fmt.Printf("Discovered Thanos URL: %s\n", thanosURL)
		fmt.Printf("Created service account token!\n")
	}

	// We run the queries
	err = runQueries(queries, thanosURL, bearerToken, clusterName)
	if err != nil {
		fmt.Printf("Failed to run commands: %v\n", err)
		return
	}

	//fmt.Printf("\n\n\n\n%v", queriesResults)
	fmt.Println("All queries completed successfully!")

}

func setupFlags() (string, string, string, string, error) {
	var bearerToken, thanosURL, kubeconfig, clusterName string
	// Parse command line flags for authentication options
	flag.StringVar(&bearerToken, "token", "", "bearer token for thanos-queries")
	flag.StringVar(&thanosURL, "thanos-url", "", "thanos url for http requests")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig file path")
	flag.StringVar(&clusterName, "cluster-name", "", "cluster name (required)")

	flag.Parse()
	// Validate flags
	err := validateFlags(bearerToken, thanosURL, kubeconfig, clusterName)
	return bearerToken, thanosURL, kubeconfig, clusterName, err
}

func loadKPIQueries() ([]string, error) {
	// Open the kpis.json file
	kpisFile, err := os.Open("kpis.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open kpis.json: %v", err)
	}
	defer kpisFile.Close()

	// Parse the JSON file into KPIs struct using a JSON decoder
	var kpis KPIs
	decoder := json.NewDecoder(kpisFile)
	if err := decoder.Decode(&kpis); err != nil {
		return nil, fmt.Errorf("failed to decode kpis.json: %v", err)
	}

	// Extract prometheus query strings from the parsed KPI configuration
	var queries []string
	for _, query := range kpis.Queries {
		queries = append(queries, query.PromQuery)
	}

	return queries, nil
}

func setupKubeconfigAuth(kubeconfig string) (string, string, error) {
	// Set up authentication and connection to Kubernetes cluster
	// using the provided kubeconfig file.
	clientset, err := setupKubernetesClient(kubeconfig)
	if err != nil {
		return "", "", fmt.Errorf("failed to setup Kubernetes client: %v", err)
	}

	// Discover the Thanos querier URL from the OpenShift cluster
	thanosURL, err := getThanosURL(clientset)
	if err != nil {
		return "", "", fmt.Errorf("failed to get Thanos URL: %v", err)
	}

	// Create a service account token for authenticating with Thanos
	bearerToken, err := createServiceAccountToken(clientset)
	if err != nil {
		return "", "", fmt.Errorf("failed to create service account token: %v", err)
	}

	return thanosURL, bearerToken, nil
}

func setupKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	return clientset, nil
}

func getThanosURL(clientset *kubernetes.Clientset) (string, error) {
	// Equivalent to: oc get route thanos-querier -n openshift-monitoring -o jsonpath='{.spec.host}'
	// This function retrieves the Thanos querier route hostname from OpenShift monitoring namespace
	// by making a REST API call to get the route object and extracting the host from its spec
	routes, err := clientset.RESTClient().
		Get().
		AbsPath(THANOS_ROUTE_API_PATH).
		DoRaw(context.TODO())

	if err != nil {
		return "", fmt.Errorf("failed to get thanos-querier route: %v", err)
	}

	// Parse route response to get hostname
	var route Route
	if err := json.Unmarshal(routes, &route); err != nil {
		return "", fmt.Errorf("failed to parse route response: %v", err)
	}

	// Extract the host field from the route specification
	// The route spec contains the hostname where the Thanos querier is accessible
	if route.Spec.Host != "" {
		return route.Spec.Host, nil
	}

	return "", fmt.Errorf("failed to extract host from route spec")
}

func createServiceAccountToken(clientset *kubernetes.Clientset) (string, error) {
	// Equivalent to: oc create token telemeter-client -n openshift-monitoring --duration=10h
	// Creates a service account token for the telemeter-client service account
	// in the openshift-monitoring namespace with a 10-hour expiration time

	tokenRequest := &authv1.TokenRequest{
		Spec: authv1.TokenRequestSpec{
			ExpirationSeconds: int64Ptr(36000), // 10 hours = 36000 seconds
		},
	}
	// Create the token using the Kubernetes API by calling CreateToken on the service account
	result, err := clientset.CoreV1().ServiceAccounts("openshift-monitoring").
		CreateToken(context.TODO(), "telemeter-client", tokenRequest, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create service account token: %v", err)
	}

	return result.Status.Token, nil
}

// Helper function to create int64 pointer
func int64Ptr(i int64) *int64 {
	return &i
}

func validateFlags(token string, url string, kubeconfig string, clusterName string) error {
	// Check if cluster name is provided
	if clusterName == "" {
		return fmt.Errorf("cluster name is required: use --cluster-name flag")
	}

	// Check authantication flags
	if (token != "" && url != "" && kubeconfig == "") ||
		(token == "" && url == "" && kubeconfig != "") {
		return nil
	} else {
		return fmt.Errorf("invalid flag combination: either provide --token and --thanos-url, or provide --kubeconfig")
	}
}

func runQueries(queriesToRun []string, thanosURL string, bearerToken string, clusterName string) error {

	// Initialize Database
	db, err := initDB()
	if err != nil {
		return fmt.Errorf("failed to init database: %v", err)
	}
	// Get or create cluster in DB
	clusterID, err := getOrCreateCluster(db, clusterName)
	if err != nil {
		return fmt.Errorf("failed to get cluster ID: %v", err)
	}

	// Create Prometheus client
	v1api, err := setupPromClient(thanosURL, bearerToken)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	// Create Prometheus v1 API client for executing queries
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, query := range queriesToRun {
		fmt.Println("------------------------")
		fmt.Printf("Running: %s\n", query)
		// Execute query using the client library
		result, warnings, err := v1api.Query(ctx, query, time.Now())
		if err != nil {
			fmt.Println("query failed: ", err)
			if storeErr := storeQueryError(db, clusterID, query, err.Error()); storeErr != nil {
				fmt.Printf("Failed to store error: %v\n", storeErr)
			}
			continue
		}

		if len(warnings) > 0 {
			fmt.Printf("Warnings: %v\n", warnings)
		}

		// Store successful query execution
		queryID, err := storeQueryExecution(db, clusterID, query)
		if err != nil {
			fmt.Printf("Failed to store query: %v\n", err)
			continue
		}

		// Store results
		err = storeQueryResults(db, queryID, result)
		if err != nil {
			fmt.Printf("Failed to store results: %v\n", err)
		} else {
			fmt.Println("Results stored successfuly in database")
		}

		fmt.Println("------------------------")

	}

	return nil
}

func setupPromClient(thanosURL, bearerToken string) (v1.API, error) {
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
		return nil, fmt.Errorf("failed to create prometheus client: %v", err)
	}

	// Create Prometheus v1 API client for executing queries
	v1api := v1.NewAPI(client)
	return v1api, nil
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

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./kpi_metrics.db")
	if err != nil {
		return nil, err
	}

	// Create tables with your exact schema
	schema := `
    CREATE TABLE IF NOT EXISTS clusters (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        cluster_name TEXT UNIQUE NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE TABLE IF NOT EXISTS queries (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        cluster_id INTEGER REFERENCES clusters(id),
        query_text TEXT NOT NULL,
        execution_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        status TEXT DEFAULT 'success',
        error_message TEXT
    );
    
    CREATE TABLE IF NOT EXISTS query_results (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        query_id INTEGER REFERENCES queries(id),
        metric_value REAL,
        timestamp_value REAL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        
        -- Store ALL metric labels as JSON (this handles any structure)
        metric_labels TEXT  -- JSON string of all labels
    );
    
    `

	_, err = db.Exec(schema)
	return db, err
}

func getOrCreateCluster(db *sql.DB, clusterName string) (int64, error) {
	// Try to get existing cluster
	var clusterID int64
	err := db.QueryRow("SELECT id FROM clusters WHERE cluster_name = ?", clusterName).Scan(&clusterID)
	if err == nil { // Cluster exists! returning the cluster ID
		return clusterID, nil
	}

	//Create new cluster if not exists
	result, err := db.Exec("INSERT INTO clusters (cluster_name) VALUES (?)", clusterName)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func storeQueryError(db *sql.DB, clusterID int64, queryText string, errorMsg string) error {
	_, err := db.Exec(
		"INSERTS INTO queries (cluster_id, query_text, status, error_message) VALUES (?, ?, 'error', ?)",
		clusterID, queryText, errorMsg,
	)
	return err
}

func storeQueryExecution(db *sql.DB, clusterID int64, queryText string) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO queries (cluster_id, query_text) VALUES (?, ?)",
		clusterID, queryText,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func storeQueryResults(db *sql.DB, queryID int64, result model.Value) error {
	// Iterate through the vector results (assuming result is always a vector)
	vector := result.(model.Vector)
	for _, sample := range vector {
		metric := sample.Metric
		value := float64(sample.Value)
		timestamp := float64(sample.Timestamp) / 1000

		// Convert labels to JSON
		labelsJSON, err := json.Marshal(metric)
		if err != nil {
			return err
		}

		_, err = db.Exec(`
            INSERT INTO query_results 
            (query_id, metric_value, timestamp_value, metric_labels)
            VALUES (?, ?, ?, ?)`,
			queryID, value, timestamp, string(labelsJSON),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
