package logger

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestInitLogger(t *testing.T) {
	tmpFile := "test_log.log"
	defer func() {
		if err := os.Remove(tmpFile); err != nil {
			log.Printf("Warning: failed to remove temp file %s: %v", tmpFile, err)
		}
	}()

	f, err := InitLogger(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Warning: failed to close file: %v", err)
		}
	}()

	// Check the file actually exists
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	if info.Size() != 0 {
		t.Fatalf("expected empty file, got size %d", info.Size())
	}

	// Write to log to verify logging works
	log.Println("test log entry")
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if !strings.Contains(string(content), "test log entry") {
		t.Fatalf("log content missing: %s", content)
	}
}
