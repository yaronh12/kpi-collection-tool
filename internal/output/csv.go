package output

import (
	"encoding/csv"
	"encoding/json"
	"strconv"
)

func (p *Printer) printKPIsCSV(records []KPIRecord) error {
	w := csv.NewWriter(p.writer)
	defer w.Flush()

	// Write header
	if err := w.Write([]string{"id", "kpi_name", "cluster", "value", "timestamp", "execution_time", "labels"}); err != nil {
		return err
	}

	// Write records
	for _, r := range records {
		labelsJSON, _ := json.Marshal(r.Labels)
		row := []string{
			strconv.FormatInt(r.ID, 10),
			r.KPIName,
			r.Cluster,
			strconv.FormatFloat(r.Value, 'f', 6, 64),
			strconv.FormatFloat(r.Timestamp, 'f', 0, 64),
			r.ExecutionTime.Format("2006-01-02 15:04:05"),
			string(labelsJSON),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	return w.Error()
}

func (p *Printer) printClustersCSV(records []ClusterRecord) error {
	w := csv.NewWriter(p.writer)
	defer w.Flush()

	// Write header
	if err := w.Write([]string{"id", "name", "created_at", "total_metrics"}); err != nil {
		return err
	}

	// Write records
	for _, c := range records {
		row := []string{
			strconv.FormatInt(c.ID, 10),
			c.Name,
			c.CreatedAt.Format("2006-01-02 15:04:05"),
			strconv.FormatInt(c.TotalMetrics, 10),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	return w.Error()
}

func (p *Printer) printErrorsCSV(records []ErrorRecord) error {
	w := csv.NewWriter(p.writer)
	defer w.Flush()

	// Write header
	if err := w.Write([]string{"kpi_id", "error_count"}); err != nil {
		return err
	}

	// Write records
	for _, e := range records {
		row := []string{
			e.KPIID,
			strconv.Itoa(e.ErrorCount),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	return w.Error()
}

