package ui

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"stress_tester/internal/i18n"
	"stress_tester/internal/stress"
	"stress_tester/logger"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	eventsChan    chan logger.Event
	currentCancel context.CancelFunc
	translator    = i18n.New(i18n.EN)
)

func Run() {

	myApp := fyneApp.New()
	w := myApp.NewWindow(translator.T("windowTitle"))
	w.Resize(fyne.NewSize(600, 400))

	vmStat, _ := mem.VirtualMemory()
	totalMemMB := vmStat.Total / (1024 * 1024)
	freeMemMB := vmStat.Available / (1024 * 1024)

	// *** System Status UI elements ***
	sysStatusLabel := widget.NewLabel(translator.T("systemStatus"))
	sysStatusLabel.TextStyle = fyne.TextStyle{Bold: true}
	sysStatusLabel.Refresh()
	totalMemLabel := widget.NewLabel(fmt.Sprintf(translator.T("totalMemory"), totalMemMB))
	freeMemLabel := widget.NewLabel(fmt.Sprintf(translator.T("freeMemory"), freeMemMB))

	// *** Test Settings UI elements ***
	memoryEntry := widget.NewEntry()
	memoryEntry.SetText("8192") // default 8 GB
	threadsEntry := widget.NewEntry()
	threadsEntry.SetText(strconv.Itoa(runtime.NumCPU()))

	alertCheck := widget.NewCheck("", nil)
	stopCheck := widget.NewCheck("", nil)

	testSettingsLabel := widget.NewLabel(translator.T("testSettings"))
	testSettingsLabel.TextStyle = fyne.TextStyle{Bold: true}
	testSettingsLabel.Refresh()
	testSettingsForm := widget.NewForm(
		widget.NewFormItem(translator.T("memoryMB"), memoryEntry),
		widget.NewFormItem(translator.T("cpuThreads"), threadsEntry),
		widget.NewFormItem(translator.T("alertOnError"), alertCheck),
		widget.NewFormItem(translator.T("stopOnError"), stopCheck),
	)

	// *** Test Status UI elements ***
	testStatusLabel := widget.NewLabel(translator.T("testStatus"))
	testStatusLabel.TextStyle = fyne.TextStyle{Bold: true}
	testStatusLabel.Refresh()
	timeLabel := widget.NewLabel(fmt.Sprintf(translator.T("time"), "0:00:00:00")) // Elapsed test time
	coverageLabel := widget.NewLabel(fmt.Sprintf(translator.T("coverage"), "0%")) // Memory coverage percentage
	errorLabel := widget.NewLabel(fmt.Sprintf(translator.T("errors"), 0))         // Error count

	// Buttons for starting and stopping the test
	startButton := widget.NewButton(translator.T("start"), nil)
	stopButton := widget.NewButton(translator.T("stop"), nil)
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
	cacheCheck := widget.NewCheck(translator.T("enableCache"), nil)
	cacheCheck.SetChecked(true) // default enabled (checked)

	// Dropdown (Select) for RNG mode: Default or Custom
	rngOptions := []string{"Default", "Custom"}
	rngSelect := widget.NewSelect(rngOptions, nil)
	rngSelect.SetSelected("Default") // default selection

	// Experimental section: "Stress FPU" checkbox
	expLabel := widget.NewLabel(translator.T("experimental"))
	expLabel.TextStyle = fyne.TextStyle{Bold: true}
	expLabel.Refresh()
	fpuCheck := widget.NewCheck(translator.T("stressFPU"), nil)
	fpuCheck.SetChecked(false) // default off

	// Arrange Advanced tab elements
	advancedTabContent := container.NewVBox(
		cacheCheck,
		container.NewHBox(widget.NewLabel(translator.T("rng")), rngSelect),
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
	testingTab := container.NewTabItem(translator.T("testingTab"), testingTabContent)
	advancedTab := container.NewTabItem(translator.T("advancedTab"), advancedTabContent)
	logTab := container.NewTabItem(translator.T("logTab"), logTabContent)

	// Settings tab with language selection
	langOptions := []string{"English", "Русский"}
	langSelect := widget.NewSelect(langOptions, nil)
	langSelect.SetSelected("English")
	settingsForm := widget.NewForm(widget.NewFormItem(translator.T("language"), langSelect))
	settingsTabContent := container.NewVBox(settingsForm)
	settingsTab := container.NewTabItem(translator.T("settingsTab"), settingsTabContent)

	tabs := container.NewAppTabs(testingTab, advancedTab, logTab, settingsTab)
	tabs.SetTabLocation(container.TabLocationTop)
	w.SetContent(tabs)

	applyLanguage := func() {
		w.SetTitle(translator.T("windowTitle"))
		testingTab.Text = translator.T("testingTab")
		advancedTab.Text = translator.T("advancedTab")
		logTab.Text = translator.T("logTab")
		settingsTab.Text = translator.T("settingsTab")

		sysStatusLabel.SetText(translator.T("systemStatus"))
		totalMemLabel.SetText(fmt.Sprintf(translator.T("totalMemory"), totalMemMB))
		freeMemLabel.SetText(fmt.Sprintf(translator.T("freeMemory"), freeMemMB))
		testSettingsLabel.SetText(translator.T("testSettings"))
		testSettingsForm.Items[0].Text = translator.T("memoryMB")
		testSettingsForm.Items[1].Text = translator.T("cpuThreads")
		testSettingsForm.Items[2].Text = translator.T("alertOnError")
		testSettingsForm.Items[3].Text = translator.T("stopOnError")
		testSettingsForm.Refresh()
		testStatusLabel.SetText(translator.T("testStatus"))
		startButton.SetText(translator.T("start"))
		stopButton.SetText(translator.T("stop"))
		cacheCheck.SetText(translator.T("enableCache"))
		expLabel.SetText(translator.T("experimental"))
		fpuCheck.SetText(translator.T("stressFPU"))
		tabs.Refresh()
	}

	// Language selector handler
	langSelect.OnChanged = func(value string) {
		if value == "Русский" {
			translator.SetLanguage(i18n.RU)
		} else {
			translator.SetLanguage(i18n.EN)
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
	errorCount := 0

	// Goroutine to listen for events from the stress test logic and update the UI accordingly
	go func() {
		for event := range eventsChan {
			// All UI updates must be executed on the main thread
			fyne.Do(func() {
				switch event.Type {
				case logger.EventLog:
					// Append log message to the log text area
					appendLogLine(logEntry, event.Message)
				case logger.EventError:
					// Append error message to log and update error count label
					appendLogLine(logEntry, event.Message)
					errorCount++
					errorLabel.SetText(fmt.Sprintf(translator.T("errors"), errorCount))
					// If "Alert on error" is enabled, show a pop-up alert for this error
					if alertCheck.Checked {
						dialog.ShowError(fmt.Errorf("%s", event.Message), w)
					}
				case logger.EventProgress:
					// Update coverage percentage display
					coverageLabel.SetText(fmt.Sprintf(translator.T("coverage"), fmt.Sprintf("%.2f%%", event.Coverage)))
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
			dialog.ShowError(fmt.Errorf(translator.T("invalidNumbers")), w)
			return
		}
		// Check if requested memory is not more than available memory
		vmStat, _ = mem.VirtualMemory()
		availMB := vmStat.Available / (1024 * 1024)
		if uint64(memMB) > availMB {
			dialog.ShowError(fmt.Errorf(translator.T("memExceed", memMB, availMB)), w)
			return
		}

		// Prepare the test configuration from UI settings
		config := stress.Config{
			MemoryMB:    memMB,
			NumThreads:  threads,
			EnableCache: cacheCheck.Checked,
			UseFPU:      fpuCheck.Checked,
			StopOnError: stopCheck.Checked,
		}
		// (RNG selection is not used in this implementation, but we store it for future use)
		config.RNGCustom = (rngSelect.Selected == "Custom")

		// Reset UI status fields for a new test run
		timeLabel.SetText(fmt.Sprintf(translator.T("time"), "0:00:00:00"))
		coverageLabel.SetText(fmt.Sprintf(translator.T("coverage"), "0%"))
		errorLabel.SetText(fmt.Sprintf(translator.T("errors"), 0))
		errorCount = 0

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
		runner := stress.Runner{
			Events: eventsChan,
			OnAbort: func() {
				if currentCancel != nil {
					currentCancel()
				}
			},
		}
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
						timeLabel.SetText(fmt.Sprintf(translator.T("time"), formatDuration(elapsed)))
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
				freeMemLabel.SetText(fmt.Sprintf(translator.T("freeMemory"), freeMemMB))
				// Final time label
				timeLabel.SetText(fmt.Sprintf(translator.T("time"), finalDuration))
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
			runner.RunMemory(ctx, config)
		}()

		// Start CPU stress test
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.RunCPU(ctx, config)
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
							freeMemLabel.SetText(fmt.Sprintf(translator.T("freeMemory"), freeMemMB))
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
				freeMemLabel.SetText(fmt.Sprintf(translator.T("freeMemory"), freeMemMB))

				// Log final memory state
				logger.LogEvent(fmt.Sprintf("Test stopped. Free memory: %d MB", freeMemMB), logger.EventLog, eventsChan)
			})

			// Clear the cancel function
			currentCancel = nil
		}
	}

	w.ShowAndRun()
}

func appendLogLine(logEntry *widget.Entry, message string) {
	if logEntry.Text == "" {
		logEntry.SetText(message)
	} else {
		logEntry.SetText(logEntry.Text + "\n" + message)
	}
	logEntry.Refresh()
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
