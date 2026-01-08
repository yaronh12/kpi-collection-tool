package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/prometheus/prometheus/promql/parser"
)

// LoadKPIs loads Prometheus queries from kpis file
func LoadKPIs(filepath string) (KPIs, error) {
	kpisFile, err := os.Open(filepath)
	if err != nil {
		return KPIs{}, fmt.Errorf("failed to open kpis file: %v", err)
	}
	defer func() {
		if closeErr := kpisFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close kpis file: %v\n", closeErr)
		}
	}()

	var kpis KPIs
	decoder := json.NewDecoder(kpisFile)
	if err := decoder.Decode(&kpis); err != nil {
		return KPIs{}, fmt.Errorf("failed to decode kpis file: %v", err)
	}

	return kpis, nil
}

// ValidateKPIs checks all KPI queries for syntax errors and configuration issues.
// Returns a slice of errors found during validation.
func ValidateKPIs(kpis KPIs) []error {
	var errors []error
	seenIDs := make(map[string]bool)

	for _, kpi := range kpis.Queries {
		// Check for empty KPI ID
		if strings.TrimSpace(kpi.ID) == "" {
			errors = append(errors, fmt.Errorf("KPI has empty ID"))
			continue
		}

		// Check for duplicate IDs
		if seenIDs[kpi.ID] {
			errors = append(errors, fmt.Errorf("duplicate KPI ID: %s", kpi.ID))
		}
		seenIDs[kpi.ID] = true

		// Check for empty queries
		if strings.TrimSpace(kpi.PromQuery) == "" {
			errors = append(errors, fmt.Errorf("KPI '%s': empty PromQL query", kpi.ID))
			continue
		}

		// Validate PromQL syntax
		if _, err := parser.ParseExpr(kpi.PromQuery); err != nil {
			errors = append(errors, fmt.Errorf("KPI '%s': invalid PromQL syntax - %w", kpi.ID, err))
		}
	}

	return errors
}
