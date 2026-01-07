// Package logger provides file-based logging initialization for the KPI collector.
// It configures the standard log package to write to a specified file with
// timestamps and source file information.
package logger

import (
	"fmt"
	"log"
	"os"
)

// InitLogger initializes a logger that writes to a file and sets log format.
func InitLogger(logFile string) (*os.File, error) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	return file, nil
}