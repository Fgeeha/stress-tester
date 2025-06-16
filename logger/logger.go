package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"
)

var (
	logFileMutex sync.Mutex
	logFilePath  = "stress_test.log"
)

// Event types for communication between test logic and UI
const (
	EventLog      = iota // Regular log message
	EventError           // Error occurred
	EventProgress        // Progress update
)

// Event represents a message sent from the test logic to the UI
type Event struct {
	Type       int     // Event type (EventLog, EventError, EventProgress)
	Message    string  // Message content
	ErrorCount int     // Number of errors (for EventError)
	Coverage   float64 // Memory coverage percentage (for EventProgress)
}

// Initialize the log file
func InitLogFile() error {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	// Clear the file by opening with O_TRUNC
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create/open log file: %v", err)
	}
	file.Close()
	return nil
}

// Write a line to the log file (with timestamp)
func LogToFile(message string) {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	file.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
}

// AppendLogLine adds a line to the UI log widget
func AppendLogLine(logEntry *widget.Entry, message string) {
	// Clear the log if it's empty
	if logEntry.Text == "" {
		logEntry.SetText(message)
	} else {
		logEntry.SetText(logEntry.Text + "\n" + message)
	}
	logEntry.Refresh()
}

// LogEvent sends an event to the UI and log file
func LogEvent(message string, eventType int, eventsChan chan Event) {
	LogToFile(message)
	event := Event{
		Type:    eventType,
		Message: message,
	}
	select {
	case eventsChan <- event:
	default:
		LogToFile("Warning: Event channel is full, message dropped: " + message)
	}
}

// LogError sends an error event to the UI and log file
func LogError(message string, errorCount int, eventsChan chan Event) {
	LogToFile(message)
	event := Event{
		Type:       EventError,
		Message:    message,
		ErrorCount: errorCount,
	}
	eventsChan <- event
}

// LogProgress sends a progress event to the UI
func LogProgress(coverage float64, eventsChan chan Event) {
	event := Event{
		Type:     EventProgress,
		Coverage: coverage,
	}
	eventsChan <- event
}
