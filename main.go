package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
)

const (
	HISTORY_DURATION = 7
)

type Config struct {
	ExamName       string `yaml:"exam_name"`
	ExamSection    string `yaml:"exam_section"`
	TotalQuestions uint   `yaml:"total_questions"`
	TotalTime      uint   `yaml:"total_time"`
}

func main() {
	var config Config
	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	// initialize termui
	err = ui.Init()
	if err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	session := NewSession(config.TotalQuestions)

	logo := widgets.NewParagraph()
	logo.Title = "Session Summary"
	logo.Text = fmt.Sprintf("Exam: %s - Section %s\nTime: %d Hour %d Minutes\nQuestions: %d", config.ExamName, config.ExamSection, int(config.TotalTime/60), config.TotalTime%60, config.TotalQuestions)
	//logo.Border = false
	logo.TextStyle.Fg = ui.ColorBlue

	clock := widgets.NewParagraph()
	clock.Title = "Time"
	clock.Text = fmt.Sprintf("%s\n%s", time.Now().Format("2006/01/02"), time.Now().Format("15:04:05"))
	//clock.Border = false
	clock.TextStyle.Fg = ui.ColorMagenta

	progress := widgets.NewGauge()
	//progress.Border = false
	progress.Percent = 0
	progress.Title = "Progress"
	progress.Label = fmt.Sprintf("Q%s/%s (%s%%)", "--", "--", "--")
	progress.LabelStyle.Fg = ui.ColorMagenta

	current := widgets.NewParagraph()
	current.Title = "Current"
	//current.Border = false
	current.PaddingLeft = 15
	current.PaddingTop, current.PaddingBottom = 3, 3
	current.TextStyle.Fg = ui.ColorGreen
	current.Text = "Q-- - --:--\nPress \"s\" to start"

	history := widgets.NewBarChart()
	history.Title = "History"
	//history.Border = false
	history.Labels = []string{}
	history.Data = []float64{}
	history.BarGap = 2

	quote := widgets.NewParagraph()
	//quote.Border = false
	quote.Title = "Quote"
	quote.PaddingTop = 1
	quote.Text = "It does not matter how slowly you go as long as you do not stop. - Confucius"
	quote.TextStyle.Fg = ui.ColorBlue

	help := widgets.NewParagraph()
	//help.Border = false
	help.Title = "Keybinds"
	help.PaddingTop = 1
	help.Text = "[q] Quit   (space) Next Question   [e] End Session   [p] Pause/resume"
	help.TextStyle.Fg = ui.ColorYellow

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	refresh := func() {
		clock.Text = fmt.Sprintf("%s\n%s", time.Now().Format("2006/01/02"), time.Now().Format("15:04:05"))
		ui.Render(clock)
		if session.Started {
			currentTimer := session.CurrentTimer()
			currentTimerElapsed := time.Now().Sub(currentTimer.LastStart)
			elapsed := session.GetEdited()
			percent := int((float64(elapsed) / float64(config.TotalQuestions)) * 100)

			progress.Percent = int(percent)
			progress.Label = fmt.Sprintf("Q%02d/%02d (%02d%%)", elapsed, config.TotalQuestions, percent)

			var labels []string
			var dataset []float64

			var loopStart, loopUntil uint
			currentIndex := session.Current
			if currentIndex <= HISTORY_DURATION-1 {
				loopStart = 0
				loopUntil = HISTORY_DURATION - 1
			} else {
				loopStart = currentIndex - (HISTORY_DURATION - 1)
				loopUntil = currentIndex
			}

			for i := loopStart; i <= loopUntil; i++ {
				timer := session.Timers[i]
				seconds := timer.Duration // +0
				if timer.Running() {
					seconds += timer.LastStart.Sub(time.Now())
				}
				seconds_value := -int(seconds.Seconds())
				//fmt.Print(seconds)
				if seconds_value == 0 {
					continue
				}
				labels = append(labels, fmt.Sprintf("Q%d", i+1))
				dataset = append(dataset, float64(seconds_value))
			}
			history.Labels = labels
			history.Data = dataset
			//history.Text = fmt.Sprintf("%v, %v", labels, dataset)
			current.Text = fmt.Sprintf("Q%02d - %02d:%02d", currentTimer.Index, int(currentTimerElapsed.Minutes()), int(currentTimerElapsed.Seconds()))
			ui.Render(progress, history, current)
		}
	}

	grid.Set(
		ui.NewRow(1.0/6,
			ui.NewCol(1.0/2, logo),
			ui.NewCol(1.0/2, clock),
		),
		ui.NewRow(1.0/6, progress),
		ui.NewRow(2.0/6,
			ui.NewCol(1.0/2, current),
			ui.NewCol(1.0/2, history),
		),
		ui.NewRow(1.0/6, quote),
		ui.NewRow(1.0/6, help),
	)

	ui.Render(grid)

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(1 * time.Second).C

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(grid)
			case "s":
				err := session.Start()
				refresh()
				if err != nil {
					panic(err)
				}
			case "<Space>":
				err := session.Next()
				refresh()
				if err != nil {
					panic(err)
				}
			case "e":
				err := session.End()
				refresh()
				if err != nil {
					panic(err)
				}
			case "p":
				err := session.SwitchPause()
				refresh()
				if err != nil {
					panic(err)
				}
			}
		case <-ticker:
			// update timer
			refresh()
		}
	}
}
