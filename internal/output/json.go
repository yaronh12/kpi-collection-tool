package output

import (
	"encoding/json"
)

func (p *Printer) printKPIsJSON(records []KPIRecord) error {
	encoder := json.NewEncoder(p.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(records)
}
