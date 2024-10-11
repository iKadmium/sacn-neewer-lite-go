package main

import (
	"context"
	"sacn_neewer_lite_go/status"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TerminalUI struct {
	app        *tview.Application
	statusList []*StatusUiTuple
}

func NewTerminalUI() *TerminalUI {
	return &TerminalUI{
		app:        tview.NewApplication(),
		statusList: make([]*StatusUiTuple, 0),
	}
}

func (ui *TerminalUI) Run() error {
	return ui.app.Run()
}

func (ui *TerminalUI) PoorUpdate(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				println(time.Now().String())
				for _, statusUi := range ui.statusList {
					println(statusUi.status.GetText())
					println("Update Rate: " + strconv.Itoa(statusUi.status.LastCount()))
				}

				time.Sleep(time.Second)
			}
		}
	}()
}

func (ui *TerminalUI) Update(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for _, statusUi := range ui.statusList {
					ui.app.QueueUpdateDraw(func() {
						statusUi.textView.SetText(statusUi.status.GetText())
						statusUi.textView.SetTextColor(statusUi.status.GetColor())
						if statusUi.updateRateView != nil {
							statusUi.updateRateView.SetText("Update Rate: " + strconv.Itoa(statusUi.status.LastCount()))
						}
					})
				}

				time.Sleep(time.Second / 10)
			}
		}
	}()
}

func (ui *TerminalUI) Setup(config *Config, app_status_tracker *status.Status, sacn_status_tracker *status.Status, light_status_trackers map[string]*status.Status, cancel func()) {
	root := tview.NewFlex()
	root.SetTitle("Sacn-Neewer-Lite").SetTitleAlign(tview.AlignLeft).SetBorder(true)
	root.SetDirection(tview.FlexRow)

	app_status_container := tview.NewFlex()
	app_status_container.SetTitle("Status").SetTitleAlign(tview.AlignLeft).SetBorder(true)
	app_status_container.SetDirection(tview.FlexRow)
	root.AddItem(app_status_container, 3, 1, false)

	app_status_text := tview.NewTextView()
	app_status_text.SetText(app_status_tracker.GetText())
	app_status_container.AddItem(app_status_text, 0, 1, false)
	ui.statusList = append(ui.statusList, NewStatusUiTuple(app_status_tracker, app_status_text, nil))

	sacn_container := tview.NewFlex()
	sacn_container.SetTitle("Sacn").SetTitleAlign(tview.AlignLeft).SetBorder(true)
	sacn_container.SetDirection(tview.FlexRow)
	root.AddItem(sacn_container, 4, 1, false)

	sacn_status_text := tview.NewTextView()
	sacn_status_text.SetText(sacn_status_tracker.GetText())
	sacn_container.AddItem(sacn_status_text, 0, 1, false)
	sacn_update_rate_text := tview.NewTextView()
	sacn_container.AddItem(sacn_update_rate_text, 0, 1, false)
	ui.statusList = append(ui.statusList, NewStatusUiTuple(sacn_status_tracker, sacn_status_text, sacn_update_rate_text))

	lights_container := tview.NewFlex()
	lights_container.SetTitle("Lights").SetTitleAlign(tview.AlignLeft).SetBorder(true)
	lights_container.SetDirection(tview.FlexColumn)
	root.AddItem(lights_container, 8, 1, false)

	for _, lightConfig := range config.Lights {
		container := tview.NewFlex()
		container.SetDirection(tview.FlexRow)
		container.SetTitle(lightConfig.ID).SetTitleAlign(tview.AlignLeft).SetBorder(true)

		universeText := "Universe: " + strconv.Itoa(int(lightConfig.Universe))
		container.AddItem(tview.NewTextView().SetText(universeText), 1, 1, false)
		addressText := "Address: " + strconv.Itoa(int(lightConfig.Address))
		container.AddItem(tview.NewTextView().SetText(addressText), 1, 1, false)

		statusText := tview.NewTextView()
		container.AddItem(statusText, 1, 1, false)

		rateText := tview.NewTextView()
		container.AddItem(rateText, 1, 1, false)

		ui.statusList = append(ui.statusList, NewStatusUiTuple(light_status_trackers[lightConfig.ID], statusText, rateText))
		lights_container.AddItem(container, 0, 1, false)
	}

	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			cancel()
		}
		return event
	})

	ui.app.SetRoot(root, true)
}
