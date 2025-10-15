package main

import "net/http"

// tokenRoundTripper adds Bearer token authentication to HTTP requests
type tokenRoundTripper struct {
	token string
	rt    http.RoundTripper
}

// InputFlags holds all command line flag values
type InputFlags struct {
	BearerToken string
	ThanosURL   string
	Kubeconfig  string
	ClusterName string
	InsecureTLS bool
}

// KPIs represents the structure of the kpis.json file containing
// the list of KPI queries to be executed against Prometheus/Thanos
type KPIs struct {
	Queries []struct {
		ID        string `json:"id"`
		PromQuery string `json:"promquery"`
	} `json:"kpis"`
}

// Route represents the OpenShift route object structure
type Route struct {
	Spec struct {
		Host string `json:"host"`
	} `json:"spec"`
}
