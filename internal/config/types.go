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
	DatabaseType string // "sqlite" or "postgres"
	PostgresURL  string // PostgreSQL connection string
}

// Query represents a single KPI query configuration
type Query struct {
	ID              string `json:"id"`
	PromQuery       string `json:"promquery"`
	SampleFrequency *int   `json:"sample-frequency,omitempty"`
}

// KPIs represents the structure of the kpis.json file containing
// the list of KPI queries to be executed against Prometheus/Thanos
type KPIs struct {
	Queries []Query `json:"kpis"`
}

// GetEffectiveFrequency returns the sample frequency for this query,
// falling back to the provided default if not specified
func (q *Query) GetEffectiveFrequency(defaultFreq int) int {
	if q.SampleFrequency != nil && *q.SampleFrequency > 0 {
		return *q.SampleFrequency
	}
	return defaultFreq
}
