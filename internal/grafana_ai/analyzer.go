package grafana_ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"kpi-collector/internal/config"
)

func Run(flags config.InputFlags) error {
	// Parse the Grafana dashboard JSON
	dash, err := ParseGrafanaFile(flags.GrafanaFile)
	if err != nil {
		return fmt.Errorf("parse grafana file: %w", err)
	}

	// Extract basic statistics from the dashboard
	stats := ExtractBasicStats(dash)

	// Build prompt for Ollama AI
	prompt := BuildPrompt(dash, stats, nil)

	model := flags.AIModel
	if model == "" {
		model = "llama3.2:latest"
	}

	// Run Ollama locally
	cmd := exec.Command("ollama", "run", model)
	cmd.Stdin = bytes.NewBufferString(prompt)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ollama run failed: %w (output: %s)", err, string(out))
	}

	// Ensure output directory exists
	outDir := "out"
	_ = os.MkdirAll(outDir, 0755)

	// Save AI summary text file
	ts := time.Now().Format("20060102_150405")
	txtPath := filepath.Join(outDir, fmt.Sprintf("summary_%s.txt", ts))
	_ = os.WriteFile(txtPath, out, 0644)

	// Save meta JSON
	meta := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"dashboard": map[string]interface{}{
			"title":   dash.Title,
			"panels":  stats.PanelCount,
			"queries": stats.QueryCount,
		},
		"ollama_model": model,
	}
	if jb, err := json.MarshalIndent(meta, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(outDir, fmt.Sprintf("summary_%s.meta.json", ts)), jb, 0644)
	}

	fmt.Println("===== Grafana AI Analysis =====")
	fmt.Println(string(out))
	fmt.Printf("\nSaved: %s\n", txtPath)

	return nil
}
