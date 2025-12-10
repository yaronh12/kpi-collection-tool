package config

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration is a wrapper around time.Duration that supports JSON unmarshaling
// from both duration strings (e.g., "30s", "2m", "1h") and integer seconds
type Duration struct {
	time.Duration
}

// UnmarshalJSON implements json.Unmarshaler for Duration
// Supports both string format ("30s", "2m", "1h") and integer seconds (60)
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		// JSON numbers are float64, treat as seconds for backward compatibility
		d.Duration = time.Duration(value) * time.Second
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration string %q: %w", value, err)
		}
		return nil
	default:
		return fmt.Errorf("invalid duration type: %T", v)
	}
}

// MarshalJSON implements json.Marshaler for Duration
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// InputFlags holds all command line flag values
type InputFlags struct {
	BearerToken  string
	ThanosURL    string
	Kubeconfig   string
	ClusterName  string
	ClusterType  string
	InsecureTLS  bool
	SamplingFreq time.Duration
	Duration     time.Duration
	OutputFile   string
	LogFile      string
	DatabaseType string // "sqlite" or "postgres"
	PostgresURL  string // PostgreSQL connection string
	KPIsFile     string
}

// Query represents a single KPI query configuration
type Query struct {
	ID              string    `json:"id"`
	PromQuery       string    `json:"promquery"`
	SampleFrequency *Duration `json:"sample-frequency,omitempty"`
}

// KPIs represents the structure of the kpis.json file containing
// the list of KPI queries to be executed against Prometheus/Thanos
type KPIs struct {
	Queries []Query `json:"kpis"`
}

// GetEffectiveFrequency returns the sample frequency for this query,
// falling back to the provided default if not specified
func (q *Query) GetEffectiveFrequency(defaultFreq time.Duration) time.Duration {
	if q.SampleFrequency != nil && q.SampleFrequency.Duration > 0 {
		return q.SampleFrequency.Duration
	}
	return defaultFreq
}
