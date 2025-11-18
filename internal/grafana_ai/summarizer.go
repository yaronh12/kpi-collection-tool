package grafana_ai

import (
	"fmt"
	"strings"
)

func BuildPrompt(d *Dashboard, stats DashboardStats, kpiData map[string]interface{}) string {
	var b strings.Builder

	b.WriteString("You are an offline data analyst specialized in Grafana dashboards for KPI monitoring.\n")
	b.WriteString("Context: This dashboard uses SQLite datasource with query_results, query_errors, clusters.\n\n")

	b.WriteString("Dashboard metadata:\n")
	b.WriteString(fmt.Sprintf("- Title: %s\n", d.Title))
	b.WriteString(fmt.Sprintf("- Panels: %d\n", stats.PanelCount))
	b.WriteString(fmt.Sprintf("- Queries (estimated): %d\n", stats.QueryCount))
	b.WriteString(fmt.Sprintf("- Datasources: %v\n\n", stats.DatasourceMap))

	b.WriteString("Panels detail:\n")
	for i, p := range d.Panels {
		b.WriteString(fmt.Sprintf("%d) %s [%s] datasource=%s\n", i+1, p.Title, p.Type, p.Datasource))
		for _, sql := range p.RawSQL {
			sn := sql
			if len(sn) > 200 {
				sn = sn[:200] + "..."
			}
			b.WriteString(fmt.Sprintf("  - SQL: %s\n", sn))
		}
	}

	if kpiData != nil {
		b.WriteString("\nProvided sample KPI data (summary):\n")
		for k, v := range kpiData {
			switch vv := v.(type) {
			case []interface{}:
				b.WriteString(fmt.Sprintf("- %s : %d samples\n", k, len(vv)))
			default:
				b.WriteString(fmt.Sprintf("- %s : type=%T\n", k, vv))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\nTasks:\n")
	b.WriteString("1) Provide a concise high-level summary of what this dashboard monitors.\n")
	b.WriteString("2) Highlight potential issues suggested by queries or panel types.\n")
	b.WriteString("3) Identify unstable KPIs or KPIs with high execution time (if KPI data provided).\n")
	b.WriteString("4) Provide recommended next steps (alerts, aggregation, SQL improvements).\n")
	b.WriteString("\nResponse format: Summary, Insights, Detected Issues, Suggested Actions.\n")

	return b.String()
}
