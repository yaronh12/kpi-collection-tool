// Package config provides configuration types, validation, and loading
// for the KPI collector. It defines input flags, KPI query structures,
// and validation logic for command-line arguments.
package config

import (
	"fmt"
	"strings"
	"time"
)

// Duration is a wrapper around time.Duration that supports YAML unmarshaling
// from both duration strings (e.g., "30s", "2m", "1h") and integer seconds
type Duration struct {
	time.Duration
}

// UnmarshalYAML implements yaml.Unmarshaler for Duration.
// Supports both string format ("30s", "2m", "1h") and integer seconds (60).
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v interface{}
	if err := unmarshal(&v); err != nil {
		return err
	}

	switch value := v.(type) {
	case int:
		d.Duration = time.Duration(value) * time.Second
		return nil
	case float64:
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

// MarshalYAML implements yaml.Marshaler for Duration
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.String(), nil
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
	DatabaseType string // "sqlite" or "postgres"
	PostgresURL  string // PostgreSQL connection string
	KPIsFile     string
	SingleRun    bool // collect metrics once and exit
}

// Query represents a single KPI query configuration
type Query struct {
	ID              string    `yaml:"id"`
	PromQuery       string    `yaml:"promquery"`
	SampleFrequency *Duration `yaml:"sample-frequency,omitempty"`
	QueryType       string    `yaml:"query-type,omitempty"`
	Step            *Duration `yaml:"step,omitempty"`
	Range           *Duration `yaml:"range,omitempty"`
	RunOnce         *bool     `yaml:"run-once,omitempty"`
}

// IsRunOnce returns true if this query is configured to run only once
func (q *Query) IsRunOnce() bool {
	return q.RunOnce != nil && *q.RunOnce
}

// KPIs represents the structure of the KPI configuration file containing
// the list of KPI queries to be executed against Prometheus/Thanos
type KPIs struct {
	Queries []Query `yaml:"kpis"`
}

// GetEffectiveFrequency returns the sample frequency for this query,
// falling back to the provided default if not specified
func (q *Query) GetEffectiveFrequency(defaultFreq time.Duration) time.Duration {
	if q.SampleFrequency != nil && q.SampleFrequency.Duration > 0 {
		return q.SampleFrequency.Duration
	}
	return defaultFreq
}

// GetEffectiveQueryType returns the query type for this query,
// defaulting to "instant" if not specified
func (q *Query) GetEffectiveQueryType() string {
	if qt := strings.TrimSpace(q.QueryType); qt != "" {
		return qt
	}
	return "instant"
}
