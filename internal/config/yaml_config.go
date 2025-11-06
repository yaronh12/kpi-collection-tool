package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadConfigFromYAML loads configuration from a YAML file
func LoadConfigFromYAML(filePath string) (InputFlags, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return InputFlags{}, fmt.Errorf("failed to read config file: %v", err)
	}

	// Start with defaults
	yamlConfig := DefaultValuesYAMLConfig()

	// Unmarshal will override only the fields present in YAML
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return InputFlags{}, fmt.Errorf("failed to parse YAML config: %v", err)
	}

	// Convert YAML config to InputFlags
	flags, err := yamlConfigToInputFlags(yamlConfig)
	if err != nil {
		return InputFlags{}, err
	}

	// Validate the loaded configuration
	if err := validateFlags(flags); err != nil {
		return InputFlags{}, err
	}

	return flags, nil
}

// yamlConfigToInputFlags converts YAMLConfig to InputFlags
func yamlConfigToInputFlags(yamlConfig YAMLConfig) (InputFlags, error) {
	// Parse duration string
	duration, err := time.ParseDuration(yamlConfig.Duration)
	if err != nil {
		return InputFlags{}, fmt.Errorf("invalid duration format: %v", err)
	}

	return InputFlags{
		BearerToken:  yamlConfig.BearerToken,
		ThanosURL:    yamlConfig.ThanosURL,
		Kubeconfig:   yamlConfig.Kubeconfig,
		ClusterName:  yamlConfig.ClusterName,
		InsecureTLS:  yamlConfig.InsecureTLS,
		SamplingFreq: yamlConfig.SamplingFrequency,
		Duration:     duration,
		OutputFile:   yamlConfig.OutputFile,
		LogFile:      yamlConfig.LogFile,
		DatabaseType: yamlConfig.Database.Type,
		PostgresURL:  yamlConfig.Database.PostgresURL,
	}, nil
}

// DefaultValuesYAMLConfig returns a YAMLConfig with default values
func DefaultValuesYAMLConfig() YAMLConfig {
	return YAMLConfig{
		InsecureTLS:       false,
		SamplingFrequency: 60,
		Duration:          "45m",
		OutputFile:        "kpi-output.json",
		LogFile:           "kpi.log",
		Database: DatabaseConfig{
			Type: "sqlite",
		},
	}
}
