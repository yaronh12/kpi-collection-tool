package collector

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestSaveJSON(t *testing.T) {
	tmpFile := "test_output.json"

	defer func() {
		if err := os.Remove(tmpFile); err != nil {
			fmt.Printf("Failed to remove temp file %s: %v\n", tmpFile, err)
		}
	}()

	data := map[string]interface{}{
		"a": 123,
	}
	if err := SaveJSON(data, tmpFile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !strings.Contains(string(content), `"a": 123`) {
		t.Fatalf("file content mismatch: %s", content)
	}
}

func TestExecuteCommand(t *testing.T) {
	// valid JSON command
	result, err := ExecuteCommand("echo '{\"key\": \"value\"}'")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data, ok := result.(map[string]interface{})
	if !ok || data["key"] != "value" {
		t.Fatalf("unexpected result: %v", result)
	}

	// invalid JSON
	_, err = ExecuteCommand("echo 'invalid-json'")
	if err == nil {
		t.Fatalf("expected error for invalid JSON")
	}

	// command fails
	_, err = ExecuteCommand("exit 1")
	if err == nil {
		t.Fatalf("expected error for failing command")
	}
}
