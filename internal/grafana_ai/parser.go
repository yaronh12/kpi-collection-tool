package grafana_ai

import (
	"encoding/json"
	"fmt"
	"os"
)

type Panel struct {
	Title      string   `json:"title"`
	Type       string   `json:"type"`
	Datasource string   `json:"datasource,omitempty"`
	RawSQL     []string `json:"raw_sql,omitempty"`
}

type Dashboard struct {
	Title   string                 `json:"title"`
	Panels  []Panel                `json:"panels"`
	RawJSON map[string]interface{} `json:"-"`
}

func ParseGrafanaFile(path string) (*Dashboard, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read grafana file: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal grafana json: %w", err)
	}

	d := &Dashboard{
		Title:   "",
		RawJSON: raw,
	}

	if t, ok := raw["title"].(string); ok {
		d.Title = t
	}

	if panelsRaw, ok := raw["panels"].([]interface{}); ok {
		for _, p := range panelsRaw {
			pm, ok := p.(map[string]interface{})
			if !ok {
				continue
			}
			panel := Panel{}
			if v, ok := pm["title"].(string); ok {
				panel.Title = v
			}
			if v, ok := pm["type"].(string); ok {
				panel.Type = v
			}

			if targets, ok := pm["targets"].([]interface{}); ok {
				for _, t := range targets {
					if tm, ok := t.(map[string]interface{}); ok {
						if ds, ok := tm["datasource"].(string); ok {
							panel.Datasource = ds
						} else if dsObj, ok := tm["datasource"].(map[string]interface{}); ok {
							if n, ok := dsObj["type"].(string); ok {
								panel.Datasource = n
							}
						}
						if rawSql, ok := tm["rawSql"].(string); ok {
							panel.RawSQL = append(panel.RawSQL, rawSql)
						}
					}
				}
			}

			d.Panels = append(d.Panels, panel)
		}
	}

	return d, nil
}

