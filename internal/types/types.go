package types

import "net/http"

// tokenRoundTripper adds Bearer token authentication to HTTP requests
type TokenRoundTripper struct {
	Token string
	RT    http.RoundTripper
}

func (t *TokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.Token)
	return t.RT.RoundTrip(req)
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
