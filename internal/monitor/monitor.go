// File: internal/monitor/monitor.go
package monitor

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// Monitor collects system metrics
type Monitor struct {
	Interval time.Duration
	StopCh   chan struct{}
}

// New creates a new Monitor
func New(interval time.Duration) *Monitor {
	return &Monitor{Interval: interval, StopCh: make(chan struct{})}
}

// Start beginning monitoring loop
func Start(m *Monitor) {
	ticker := time.NewTicker(m.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cpuPercent, _ := cpu.Percent(0, false)
			vm, _ := mem.VirtualMemory()
			// TODO: send metrics via channel or callback
			_ = cpuPercent
			_ = vm
		case <-m.StopCh:
			return
		}
	}
}
