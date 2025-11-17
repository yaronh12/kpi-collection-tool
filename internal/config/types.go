package config

import "time"

// InputFlags holds all command line flag values
type InputFlags struct {
	BearerToken  string
	ThanosURL    string
	Kubeconfig   string
	ClusterName  string
	InsecureTLS  bool
	SamplingFreq int
	Duration     time.Duration
	OutputFile   string
	LogFile      string
	DatabaseType string
	PostgresURL  string
	KPIsFile     string
}

// KPIs represents the structure of the kpis.json file containing
// the list of KPI queries to be executed against Prometheus/Thanos
type KPIs struct {
	Queries []struct {
		ID        string `json:"id"`
		PromQuery string `json:"promquery"`
	} `json:"kpis"`
}
