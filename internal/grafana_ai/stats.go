package grafana_ai

type DashboardStats struct {
	PanelCount    int               `json:"panel_count"`
	QueryCount    int               `json:"query_count"`
	DatasourceMap map[string]int    `json:"datasource_usage"`
	KPIsDetected  []string          `json:"kpis_detected,omitempty"`
	ErrorsPresent bool              `json:"errors_present"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

func ExtractBasicStats(d *Dashboard) DashboardStats {
	st := DashboardStats{
		PanelCount:    len(d.Panels),
		DatasourceMap: map[string]int{},
		Extra:         map[string]interface{}{},
	}

	qc := 0
	for _, p := range d.Panels {
		if len(p.RawSQL) > 0 {
			qc += len(p.RawSQL)
		}
		if p.Datasource != "" {
			st.DatasourceMap[p.Datasource]++
		}
	}
	st.QueryCount = qc
	return st
}
