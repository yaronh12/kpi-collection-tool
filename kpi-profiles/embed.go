package kpiprofiles

import "embed"

//go:embed kpis-ran.yaml kpis-core.yaml kpis-hub.yaml
var FS embed.FS

const (
	RAN  = "kpis-ran.yaml"
	Core = "kpis-core.yaml"
	Hub  = "kpis-hub.yaml"
)
