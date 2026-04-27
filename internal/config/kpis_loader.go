package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/prometheus/prometheus/promql/parser"
	"gopkg.in/yaml.v3"
)

// LoadKPIs loads Prometheus queries from a YAML kpis file
func LoadKPIs(filepath string) (KPIs, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return KPIs{}, fmt.Errorf("failed to open kpis file: %v", err)
	}

	var kpis KPIs
	if err := yaml.Unmarshal(data, &kpis); err != nil {
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

		errors = append(errors, validateQueryType(kpi)...)
	}

	return errors
}

func validateQueryType(kpi Query) []error {
	var errors []error

	switch kpi.GetEffectiveQueryType() {
	case "instant":
		if kpi.Step != nil {
			errors = append(errors, fmt.Errorf("KPI '%s': step can only be set when query-type is 'range'", kpi.ID))
		}
		if kpi.Range != nil {
			errors = append(errors, fmt.Errorf("KPI '%s': range can only be set when query-type is 'range'", kpi.ID))
		}
	case "range":
		if kpi.Step == nil || kpi.Step.Duration <= 0 {
			errors = append(errors, fmt.Errorf("KPI '%s': step is required and must be > 0 when query-type is 'range'", kpi.ID))
		}
		if kpi.Range == nil || kpi.Range.Duration <= 0 {
			errors = append(errors, fmt.Errorf("KPI '%s': range is required and must be > 0 when query-type is 'range'", kpi.ID))
		}
		if kpi.Step != nil && kpi.Range != nil && kpi.Step.Duration > kpi.Range.Duration {
			errors = append(errors, fmt.Errorf("KPI '%s': step must be less than or equal to range", kpi.ID))
		}
	default:
		errors = append(errors, fmt.Errorf("KPI '%s': invalid query-type '%s' (must be 'instant' or 'range')", kpi.ID, kpi.QueryType))
	}

	return errors
}
