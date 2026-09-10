package output

import (
	"fmt"
	"log"
	"os"
	"time"
)

const waitHeartbeatMinGap = 60 * time.Second

func taskLine(task, msg string) string {
	return fmt.Sprintf("[%s] %s", task, msg)
}

// PrintTaskStart announces a task run on stdout and in the log file.
func PrintTaskStart(task string) {
	printMutex.Lock()
	defer printMutex.Unlock()
	line := taskLine(task, "starting")
	fmt.Println(line)
	log.Print(line)
}

// PrintTaskProgress prints a milestone to stdout only.
func PrintTaskProgress(task, msg string) {
	printMutex.Lock()
	defer printMutex.Unlock()
	fmt.Println(taskLine(task, msg))
}

// PrintTaskDone announces successful completion on stdout and in the log file.
func PrintTaskDone(task string) {
	printMutex.Lock()
	defer printMutex.Unlock()
	line := taskLine(task, "completed")
	fmt.Println(line)
	log.Print(line)
}

// PrintTaskFailed announces failure on stderr and in the log file.
func PrintTaskFailed(task string, err error) {
	printMutex.Lock()
	defer printMutex.Unlock()
	line := taskLine(task, fmt.Sprintf("failed: %v", err))
	fmt.Fprintln(os.Stderr, line)
	log.Print(line)
}

// WaitReporter emits phase milestones and throttled heartbeats during pod waits.
type WaitReporter struct {
	task      string
	lastPhase string
	lastOut   time.Time
}

// NewWaitReporter creates a reporter for a single wait loop.
func NewWaitReporter(task string) *WaitReporter {
	return &WaitReporter{task: task, lastOut: time.Now()}
}

// OnPhase is called on each wait poll with the current phase and elapsed time.
func (w *WaitReporter) OnPhase(phase string, elapsed time.Duration) {
	if phase != "" && phase != w.lastPhase {
		PrintTaskProgress(w.task, "pod phase: "+phase)
		w.lastPhase = phase
		w.lastOut = time.Now()
		return
	}
	if time.Since(w.lastOut) < waitHeartbeatMinGap {
		return
	}
	PrintTaskProgress(w.task, fmt.Sprintf("still waiting (%s elapsed)", formatElapsed(elapsed)))
	w.lastOut = time.Now()
}

func formatElapsed(d time.Duration) string {
	d = d.Round(time.Second)
	if d >= time.Minute {
		return fmt.Sprintf("%dm", d/time.Minute)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

