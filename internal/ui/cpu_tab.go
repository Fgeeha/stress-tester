package ui

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (a *App) makeCPUTab() fyne.CanvasObject {
	threadsEntry := widget.NewEntry()
	threadsEntry.SetText("4")
	durationEntry := widget.NewEntry()
	durationEntry.SetText("30s")

	progressBar := widget.NewProgressBar()
	errorLabel := widget.NewLabel("")

	startBtn := widget.NewButton("Start CPU Test", func() {
		threads, err := strconv.Atoi(threadsEntry.Text)
		if err != nil {
			errorLabel.SetText("Invalid threads number")
			return
		}
		dur, err := time.ParseDuration(durationEntry.Text)
		if err != nil {
			errorLabel.SetText("Invalid duration")
			return
		}
		progressBar.SetValue(0)
		errorLabel.SetText("")

		// Run in background
		go func() {
			err := a.cpuTester.Run(dur, threads,
				func(count float64) {
					// Simple progress mapping
					progressBar.SetValue(count / float64(threads*100))
				},
				func(e error) {
					errorLabel.SetText(fmt.Sprintf("Error: %v", e))
					a.logger.Log("ERROR", "CPU", e.Error())
				},
			)
			if err != nil {
				errorLabel.SetText(fmt.Sprintf("Start error: %v", err))
				return
			}
		}()
	})

	stopBtn := widget.NewButton("Stop", func() {
		a.cpuTester.Stop()
	})

	form := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Threads", threadsEntry),
			widget.NewFormItem("Duration", durationEntry),
		),
		container.NewHBox(startBtn, stopBtn),
		progressBar,
		errorLabel,
	)
	return form
}
