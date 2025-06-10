package ui

import (
	"errors"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/fgeeha/stress-tester/internal/stress/mem"
)

// CPUInterface описывает методы CPU-тестера
type CPUInterface interface {
	Run(duration time.Duration, threads int, progress func(float64), reportError func(error)) error
	Stop()
}

// MemInterface описывает методы Memory-тестера
type MemInterface interface {
	Run(sizeMB int, pattern mem.Pattern, duration time.Duration, progress func(float64), reportError func(error)) error
	Stop()
}

// MonitorInterface — заглушка для мониторинга
type MonitorInterface interface {
	// TODO: добавить методы подписки на метрики
}

type LoggerInterface interface {
	Log(level, source, message string)
}

// App хранит все компоненты UI и ядра
type App struct {
	cpuTester CPUInterface
	memTester MemInterface
	monitor   MonitorInterface
	logger    LoggerInterface
	fyApp     fyne.App
	window    fyne.Window
}

// NewApp создаёт и настраивает главное окно
func NewApp(cpuTester CPUInterface, memTester MemInterface, mon MonitorInterface, logger LoggerInterface) *App {
	fyApp := app.New()
	w := fyApp.NewWindow("Stress Tester")

	a := &App{
		cpuTester: cpuTester,
		memTester: memTester,
		monitor:   mon,
		logger:    logger,
		fyApp:     fyApp,
		window:    w,
	}
	a.buildUI()
	return a
}

// Run запускает главный цикл
func (a *App) Run() error {
	if a.window == nil {
		return errors.New("window not initialized")
	}
	a.window.Resize(fyne.NewSize(800, 600))
	a.window.ShowAndRun()
	return nil
}

// buildUI собирает вкладки
func (a *App) buildUI() {
	tabs := container.NewAppTabs(
		container.NewTabItem("CPU Test", a.makeCPUTab()),
		container.NewTabItem("Memory Test", a.makeMemTab()),
		container.NewTabItem("Logs", a.makeLogTab()),
	)
	a.window.SetContent(tabs)
}

// makeLogTab показывает лог
func (a *App) makeLogTab() fyne.CanvasObject {
	list := widget.NewMultiLineEntry()
	list.Wrapping = fyne.TextWrapWord
	// TODO: подписать list на a.logger.Log
	return container.NewVBox(widget.NewLabel("Logs:"), list)
}
