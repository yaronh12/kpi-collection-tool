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

// YAMLConfig represents the YAML configuration file structure
type YAMLConfig struct {
	ClusterName       string         `yaml:"cluster_name"`
	BearerToken       string         `yaml:"bearer_token"`
	ThanosURL         string         `yaml:"thanos_url"`
	Kubeconfig        string         `yaml:"kubeconfig"`
	InsecureTLS       bool           `yaml:"insecure_tls"`
	SamplingFrequency int            `yaml:"sampling_frequency"`
	Duration          string         `yaml:"duration"`
	OutputFile        string         `yaml:"output_file"`
	LogFile           string         `yaml:"log_file"`
	Database          DatabaseConfig `yaml:"database"`
}

// DatabaseConfig holds database-specific configuration
type DatabaseConfig struct {
	Type        string `yaml:"type"`
	PostgresURL string `yaml:"postgres_url"`
}

// KPIs represents the structure of the kpis.json file containing
// the list of KPI queries to be executed against Prometheus/Thanos
type KPIs struct {
	Queries []struct {
		ID        string `json:"id"`
		PromQuery string `json:"promquery"`
	} `json:"kpis"`
}
