package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadKPIs loads Prometheus queries from kpis.json file
func LoadKPIs(filepath string) (KPIs, error) {
	kpisFile, err := os.Open(filepath)
	if err != nil {
		return KPIs{}, fmt.Errorf("failed to open kpis.json: %v", err)
	}
	defer func() {
		if closeErr := kpisFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close kpis.json: %v\n", closeErr)
		}
	}()

	var kpis KPIs
	decoder := json.NewDecoder(kpisFile)
	if err := decoder.Decode(&kpis); err != nil {
		return KPIs{}, fmt.Errorf("failed to decode kpis.json: %v", err)
	}

	return kpis, nil
}
