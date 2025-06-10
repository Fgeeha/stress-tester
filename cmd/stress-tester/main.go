package main

import (
	"log"
	"time"

	"github.com/fgeeha/stress-tester/internal/logger"
	"github.com/fgeeha/stress-tester/internal/monitor"
	"github.com/fgeeha/stress-tester/internal/stress/cpu"
	"github.com/fgeeha/stress-tester/internal/stress/mem"
	"github.com/fgeeha/stress-tester/internal/ui"
)

func main() {
	// Logger
	logSrv := logger.NewService()
	defer logSrv.Close()

	// Системный монитор
	mon := monitor.New(1 * time.Second)
	go monitor.Start(mon)

	// Стресс-тестеры
	cpuTester := cpu.NewTester()
	memTester := mem.NewTester()

	// UI
	app := ui.NewApp(cpuTester, memTester, mon, logSrv)
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run UI: %v", err)
	}
}
