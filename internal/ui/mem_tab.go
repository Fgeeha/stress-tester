package ui

import (
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/fgeeha/stress-tester/internal/stress/mem"
)

func (a *App) makeMemTab() fyne.CanvasObject {
	sz := widget.NewEntry()
	sz.SetText("512")
	pat := widget.NewSelect([]string{"AA", "55", "Random"}, nil)
	pat.SetSelected("AA")
	dur := widget.NewEntry()
	dur.SetText("30s")

	pb := widget.NewProgressBar()
	errLbl := widget.NewLabel("")

	start := widget.NewButton("Start Mem", func() {
		n, e1 := strconv.Atoi(sz.Text)
		d, e2 := time.ParseDuration(dur.Text)
		if e1 != nil || e2 != nil {
			errLbl.SetText("Invalid")
			return
		}
		var mpat mem.Pattern
		switch pat.Selected {
		case "AA":
			mpat = mem.PatternAA
		case "55":
			mpat = mem.Pattern55
		default:
			mpat = mem.PatternRandom
		}
		pb.SetValue(0)
		errLbl.SetText("")
		go func() {
			err := a.memTester.Run(n, mpat, d,
				func(c float64) { pb.SetValue(c / float64(d.Seconds())) },
				func(e error) { errLbl.SetText(e.Error()); a.logger.Log("ERROR", "MEM", e.Error()) })
			if err != nil {
				errLbl.SetText(err.Error())
			}
		}()
	})

	stop := widget.NewButton("Stop", func() { a.memTester.Stop() })

	return container.NewVBox(
		widget.NewLabel("Memory Stress Test"),
		widget.NewForm(
			widget.NewFormItem("Size MB", sz),
			widget.NewFormItem("Pattern", pat),
			widget.NewFormItem("Duration", dur),
		),
		container.NewHBox(start, stop),
		pb,
		errLbl,
	)
}
