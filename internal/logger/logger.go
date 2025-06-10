// File: internal/logger/logger.go
package logger

import (
	"html/template"
	"os"
	"sync"
	"time"
)

// LogEntry represents a log event
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Source    string
	Message   string
}

// Service collects logs and writes reports
type Service struct {
	entries []LogEntry
	mu      sync.Mutex
	file    *os.File
}

// NewService initializes the logger
func NewService() *Service {
	f, _ := os.Create("stress-report.html")
	return &Service{file: f}
}

// Log records an entry
func (s *Service) Log(level, source, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := LogEntry{time.Now(), level, source, msg}
	s.entries = append(s.entries, entry)
}

// Close flushes and closes the report file
func (s *Service) Close() {
	// Generate HTML report
	tmpl := template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Stress Report</title></head><body>
<h1>Stress Test Report</h1><table border="1"><tr><th>Time</th><th>Level</th><th>Source</th><th>Message</th></tr>{{range .}}<tr>
<td>{{.Timestamp}}</td><td>{{.Level}}</td><td>{{.Source}}</td><td>{{.Message}}</td>
</tr>{{end}}</table></body></html>`))
	s.mu.Lock()
	tmpl.Execute(s.file, s.entries)
	s.mu.Unlock()
	s.file.Close()
}
