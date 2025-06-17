package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"runtime"
	"runtime/debug"

	"stress_tester/logger"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/shirou/gopsutil/v3/mem"
)

type language int

const (
	langEN language = iota
	langRU
)

var currentLang = langEN

var translations = map[language]map[string]string{
	langEN: {
		"windowTitle":    "Memory/CPU Stress Tester",
		"testingTab":     "Testing",
		"advancedTab":    "Advanced",
		"logTab":         "Log",
		"settingsTab":    "Settings",
		"systemStatus":   "System Status:",
		"totalMemory":    "Total Memory: %d MB",
		"freeMemory":     "Free Memory: %d MB",
		"testSettings":   "Test Settings:",
		"memoryMB":       "Memory (MB)",
		"cpuThreads":     "CPU Threads",
		"alertOnError":   "Alert on error",
		"stopOnError":    "Stop on error",
		"testStatus":     "Test Status:",
		"time":           "Time: %s",
		"coverage":       "Coverage: %s",
		"errors":         "Errors: %d",
		"start":          "Start",
		"stop":           "Stop",
		"enableCache":    "Enable CPU Cache",
		"rng":            "RNG:",
		"experimental":   "Experimental:",
		"stressFPU":      "Stress FPU (Floating Point Unit)",
		"language":       "Language",
		"selectLang":     "Select Language",
		"invalidNumbers": "please enter valid positive numbers for memory and threads",
		"memExceed":      "requested memory (%d MB) exceeds available memory (%d MB)",
	},
	langRU: {
		"windowTitle":    "Тестер нагрузки памяти/процессора",
		"testingTab":     "Тестирование",
		"advancedTab":    "Дополнительно",
		"logTab":         "Лог",
		"settingsTab":    "Настройки",
		"systemStatus":   "Состояние системы:",
		"totalMemory":    "Всего памяти: %d МБ",
		"freeMemory":     "Свободно памяти: %d МБ",
		"testSettings":   "Настройки теста:",
		"memoryMB":       "Память (МБ)",
		"cpuThreads":     "Потоки CPU",
		"alertOnError":   "Оповещать при ошибке",
		"stopOnError":    "Остановить при ошибке",
		"testStatus":     "Статус теста:",
		"time":           "Время: %s",
		"coverage":       "Покрытие: %s",
		"errors":         "Ошибки: %d",
		"start":          "Старт",
		"stop":           "Стоп",
		"enableCache":    "Использовать кэш CPU",
		"rng":            "ГПСЧ:",
		"experimental":   "Экспериментально:",
		"stressFPU":      "Нагрузка на FPU",
		"language":       "Язык",
		"selectLang":     "Выберите язык",
		"invalidNumbers": "введите корректные положительные значения памяти и потоков",
		"memExceed":      "запрошенная память (%d МБ) превышает доступную (%d МБ)",
	},
}

func tr(key string, args ...interface{}) string {
	t := translations[currentLang][key]
	if len(args) > 0 {
		return fmt.Sprintf(t, args...)
	}
	return t
}

var (
	eventsChan    chan logger.Event
	currentCancel context.CancelFunc
)

func main() {

	myApp := fyneApp.New()
	w := myApp.NewWindow(tr("windowTitle"))
	w.Resize(fyne.NewSize(600, 400))

	vmStat, _ := mem.VirtualMemory()
	totalMemMB := vmStat.Total / (1024 * 1024)
	freeMemMB := vmStat.Available / (1024 * 1024)

	// *** System Status UI elements ***
	sysStatusLabel := widget.NewLabel(tr("systemStatus"))
	sysStatusLabel.TextStyle = fyne.TextStyle{Bold: true}
	sysStatusLabel.Refresh()
	totalMemLabel := widget.NewLabel(fmt.Sprintf(tr("totalMemory"), totalMemMB))
	freeMemLabel := widget.NewLabel(fmt.Sprintf(tr("freeMemory"), freeMemMB))

	// *** Test Settings UI elements ***
	memoryEntry := widget.NewEntry()
	memoryEntry.SetText("8192") // default 8 GB
	threadsEntry := widget.NewEntry()
	threadsEntry.SetText(strconv.Itoa(runtime.NumCPU()))

	alertCheck := widget.NewCheck("", nil)
	stopCheck := widget.NewCheck("", nil)

	testSettingsLabel := widget.NewLabel(tr("testSettings"))
	testSettingsLabel.TextStyle = fyne.TextStyle{Bold: true}
	testSettingsLabel.Refresh()
	testSettingsForm := widget.NewForm(
		widget.NewFormItem(tr("memoryMB"), memoryEntry),
		widget.NewFormItem(tr("cpuThreads"), threadsEntry),
		widget.NewFormItem(tr("alertOnError"), alertCheck),
		widget.NewFormItem(tr("stopOnError"), stopCheck),
	)

	// *** Test Status UI elements ***
	testStatusLabel := widget.NewLabel(tr("testStatus"))
	testStatusLabel.TextStyle = fyne.TextStyle{Bold: true}
	testStatusLabel.Refresh()
	timeLabel := widget.NewLabel(fmt.Sprintf(tr("time"), "0:00:00:00")) // Elapsed test time
	coverageLabel := widget.NewLabel(fmt.Sprintf(tr("coverage"), "0%")) // Memory coverage percentage
	errorLabel := widget.NewLabel(fmt.Sprintf(tr("errors"), 0))         // Error count

	// Buttons for starting and stopping the test
	startButton := widget.NewButton(tr("start"), nil)
	stopButton := widget.NewButton(tr("stop"), nil)
	stopButton.Disable() // disable Stop initially, as no test is running

	// Layout the Testing tab components vertically
	testingTabContent := container.NewVBox(
		sysStatusLabel,
		totalMemLabel,
		freeMemLabel,
		widget.NewSeparator(),
		testSettingsLabel,
		testSettingsForm,
		widget.NewSeparator(),
		testStatusLabel,
		timeLabel,
		coverageLabel,
		errorLabel,
		widget.NewSeparator(),
		container.NewHBox(layoutSpacer(), startButton, stopButton, layoutSpacer()),
	)

	// *** Advanced tab UI elements ***
	// Checkbox for "Enable CPU Cache"
	cacheCheck := widget.NewCheck(tr("enableCache"), nil)
	cacheCheck.SetChecked(true) // default enabled (checked)

	// Dropdown (Select) for RNG mode: Default or Custom
	rngOptions := []string{"Default", "Custom"}
	rngSelect := widget.NewSelect(rngOptions, nil)
	rngSelect.SetSelected("Default") // default selection

	// Experimental section: "Stress FPU" checkbox
	expLabel := widget.NewLabel(tr("experimental"))
	expLabel.TextStyle = fyne.TextStyle{Bold: true}
	expLabel.Refresh()
	fpuCheck := widget.NewCheck(tr("stressFPU"), nil)
	fpuCheck.SetChecked(false) // default off

	// Arrange Advanced tab elements
	advancedTabContent := container.NewVBox(
		cacheCheck,
		container.NewHBox(widget.NewLabel(tr("rng")), rngSelect),
		widget.NewSeparator(),
		expLabel,
		fpuCheck,
	)

	// *** Log tab UI elements ***
	// Multiline text area for logs (read-only)
	logEntry := widget.NewMultiLineEntry()
	logEntry.Wrapping = fyne.TextWrapWord // Wrap text at word boundaries
	logEntry.Disable()                    // Make it read-only (user cannot edit)
	logEntry.SetMinRowsVisible(8)         // Show at least 8 lines by default
	logEntry.Text = ""                    // Initially empty
	logEntry.Refresh()
	// Put the logEntry in a scroll container in case logs exceed the view
	logScroll := container.NewVScroll(logEntry)
	logTabContent := container.NewBorder(nil, nil, nil, nil, logScroll)

	// Create tab items for each tab and assemble them into an AppTabs container
	testingTab := container.NewTabItem(tr("testingTab"), testingTabContent)
	advancedTab := container.NewTabItem(tr("advancedTab"), advancedTabContent)
	logTab := container.NewTabItem(tr("logTab"), logTabContent)

	// Settings tab with language selection
	langOptions := []string{"English", "Русский"}
	langSelect := widget.NewSelect(langOptions, nil)
	langSelect.SetSelected("English")
	settingsForm := widget.NewForm(widget.NewFormItem(tr("language"), langSelect))
	settingsTabContent := container.NewVBox(settingsForm)
	settingsTab := container.NewTabItem(tr("settingsTab"), settingsTabContent)

	tabs := container.NewAppTabs(testingTab, advancedTab, logTab, settingsTab)
	tabs.SetTabLocation(container.TabLocationTop)
	w.SetContent(tabs)

	applyLanguage := func() {
		w.SetTitle(tr("windowTitle"))
		testingTab.Text = tr("testingTab")
		advancedTab.Text = tr("advancedTab")
		logTab.Text = tr("logTab")
		settingsTab.Text = tr("settingsTab")

		sysStatusLabel.SetText(tr("systemStatus"))
		totalMemLabel.SetText(fmt.Sprintf(tr("totalMemory"), totalMemMB))
		freeMemLabel.SetText(fmt.Sprintf(tr("freeMemory"), freeMemMB))
		testSettingsLabel.SetText(tr("testSettings"))
		testSettingsForm.Items[0].Text = tr("memoryMB")
		testSettingsForm.Items[1].Text = tr("cpuThreads")
		testSettingsForm.Items[2].Text = tr("alertOnError")
		testSettingsForm.Items[3].Text = tr("stopOnError")
		testSettingsForm.Refresh()
		testStatusLabel.SetText(tr("testStatus"))
		startButton.SetText(tr("start"))
		stopButton.SetText(tr("stop"))
		cacheCheck.SetText(tr("enableCache"))
		expLabel.SetText(tr("experimental"))
		fpuCheck.SetText(tr("stressFPU"))
		tabs.Refresh()
	}

	// Language selector handler
	langSelect.OnChanged = func(value string) {
		if value == "Русский" {
			currentLang = langRU
		} else {
			currentLang = langEN
		}
		applyLanguage()
	}

	// Apply initial language
	applyLanguage()

	// Initialize the global event channel and log file for the test logic
	eventsChan = make(chan logger.Event, 100)
	if err := logger.InitLogFile(); err != nil {
		dialog.ShowError(fmt.Errorf("log file error: %v", err), w)
	}

	// Goroutine to listen for events from the stress test logic and update the UI accordingly
	go func() {
		for event := range eventsChan {
			// All UI updates must be executed on the main thread
			fyne.Do(func() {
				switch event.Type {
				case logger.EventLog:
					// Append log message to the log text area
					logger.AppendLogLine(logEntry, event.Message)
				case logger.EventError:
					// Append error message to log and update error count label
					logger.AppendLogLine(logEntry, event.Message)
					errorLabel.SetText(fmt.Sprintf(tr("errors"), event.ErrorCount))
					// If "Alert on error" is enabled, show a pop-up alert for this error
					if alertCheck.Checked {
						dialog.ShowError(fmt.Errorf("%s", event.Message), w)
					}
				case logger.EventProgress:
					// Update coverage percentage display
					coverageLabel.SetText(fmt.Sprintf(tr("coverage"), fmt.Sprintf("%.2f%%", event.Coverage)))
				}
			})
		}
	}()

	// Handler for the Start button
	startButton.OnTapped = func() {
		// Validate and parse input values
		memMB, err1 := strconv.Atoi(memoryEntry.Text)
		threads, err2 := strconv.Atoi(threadsEntry.Text)
		if err1 != nil || err2 != nil || memMB <= 0 || threads <= 0 {
			dialog.ShowError(fmt.Errorf(tr("invalidNumbers")), w)
			return
		}
		// Check if requested memory is not more than available memory
		vmStat, _ = mem.VirtualMemory()
		availMB := vmStat.Available / (1024 * 1024)
		if uint64(memMB) > availMB {
			dialog.ShowError(fmt.Errorf(tr("memExceed", memMB, availMB)), w)
			return
		}

		// Prepare the test configuration from UI settings
		config := TestConfig{
			MemoryMB:    memMB,
			NumThreads:  threads,
			EnableCache: cacheCheck.Checked,
			UseFPU:      fpuCheck.Checked,
			StopOnError: stopCheck.Checked,
		}
		// (RNG selection is not used in this implementation, but we store it for future use)
		config.RNGCustom = (rngSelect.Selected == "Custom")

		// Reset UI status fields for a new test run
		timeLabel.SetText(fmt.Sprintf(tr("time"), "0:00:00:00"))
		coverageLabel.SetText(fmt.Sprintf(tr("coverage"), "0%"))
		errorLabel.SetText(fmt.Sprintf(tr("errors"), 0))

		// Disable input controls during the test run
		memoryEntry.Disable()
		threadsEntry.Disable()
		cacheCheck.Disable()
		rngSelect.Disable()
		fpuCheck.Disable()
		alertCheck.Disable()
		stopCheck.Disable()
		startButton.Disable()
		stopButton.Enable()

		// Start the stress test in background goroutines
		ctx, cancel := context.WithCancel(context.Background())
		currentCancel = cancel // Store the cancel function
		wg := sync.WaitGroup{}

		// Record the start time for the timer display
		startTime := time.Now()

		// Log test start
		logger.LogEvent("Starting stress test...", logger.EventLog, eventsChan)

		// Goroutine to update the elapsed time label every second while test is running
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return // stop updating time when test is finished
				case <-ticker.C:
					elapsed := time.Since(startTime)
					fyne.Do(func() {
						timeLabel.SetText(fmt.Sprintf(tr("time"), formatDuration(elapsed)))
					})
				}
			}
		}()

		// Goroutine to handle test completion (when all worker goroutines finish)
		go func() {
			wg.Wait() // wait for all test goroutines to finish
			// Perform cleanup after test completes
			runtime.GC()
			debug.FreeOSMemory() // free memory back to OS, in case a large allocation was made
			// Refresh system free memory info after test
			vmStat, _ := mem.VirtualMemory()
			freeMemMB = vmStat.Available / (1024 * 1024)
			// Log test completion and update UI components on the main thread
			elapsed := time.Since(startTime)
			finalDuration := formatDuration(elapsed)
			// Log a summary of the test result
			logger.LogEvent(fmt.Sprintf("Test finished. Duration: %s", finalDuration), logger.EventLog, eventsChan)
			// Update UI elements back on main thread
			fyne.Do(func() {
				// Update free memory label after releasing memory
				freeMemLabel.SetText(fmt.Sprintf(tr("freeMemory"), freeMemMB))
				// Final time label
				timeLabel.SetText(fmt.Sprintf(tr("time"), finalDuration))
				// Re-enable controls for next test
				memoryEntry.Enable()
				threadsEntry.Enable()
				cacheCheck.Enable()
				rngSelect.Enable()
				fpuCheck.Enable()
				alertCheck.Enable()
				stopCheck.Enable()
				startButton.Enable()
				stopButton.Disable()
				// Clear the cancel function
				currentCancel = nil
			})
		}()

		// Start memory stress test
		wg.Add(1)
		go func() {
			defer wg.Done()
			runMemoryTest(ctx, config)
		}()

		// Start CPU stress test
		wg.Add(1)
		go func() {
			defer wg.Done()
			runCPUTest(ctx, config)
		}()

		// Start periodic memory info update
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					vmStat, err := mem.VirtualMemory()
					if err == nil {
						freeMemMB := vmStat.Available / (1024 * 1024)
						fyne.Do(func() {
							freeMemLabel.SetText(fmt.Sprintf(tr("freeMemory"), freeMemMB))
						})
					}
				}
			}
		}()
	}

	// Handler for the Stop button
	stopButton.OnTapped = func() {
		if currentCancel != nil {
			// Log the stop action
			logger.LogEvent("Stopping stress test...", logger.EventLog, eventsChan)

			// Cancel the context to stop all goroutines
			currentCancel()

			// Force garbage collection to free memory
			runtime.GC()
			debug.FreeOSMemory()

			// Update UI immediately
			fyne.Do(func() {
				// Re-enable controls
				memoryEntry.Enable()
				threadsEntry.Enable()
				cacheCheck.Enable()
				rngSelect.Enable()
				fpuCheck.Enable()
				alertCheck.Enable()
				stopCheck.Enable()
				startButton.Enable()
				stopButton.Disable()

				// Update memory info
				vmStat, _ := mem.VirtualMemory()
				freeMemMB := vmStat.Available / (1024 * 1024)
				freeMemLabel.SetText(fmt.Sprintf(tr("freeMemory"), freeMemMB))

				// Log final memory state
				logger.LogEvent(fmt.Sprintf("Test stopped. Free memory: %d MB", freeMemMB), logger.EventLog, eventsChan)
			})

			// Clear the cancel function
			currentCancel = nil
		}
	}

	w.ShowAndRun()
}

// layoutSpacer returns a spacer object for centering elements in a horizontal box
func layoutSpacer() *fyne.Container {
	return container.NewHBox(layout.NewSpacer())
}

// formatDuration formats a duration as "days:hh:mm:ss"
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%d:%02d:%02d:%02d", h/24, h%24, m, s)
}
