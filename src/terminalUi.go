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

func (ui *TerminalUI) UpdateSendText(ctx context.Context) {
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

func (ui *TerminalUI) UpdateSend(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for _, statusUi := range ui.statusList {
					statusUi.Update(ui.app)
				}

				time.Sleep(time.Second / 60)
			}
		}
	}()
}

func (ui *TerminalUI) SetupSend(config *Config, app_status_tracker *status.Status, sacn_status_tracker *status.Status, light_status_trackers map[string]*status.Status, cancel func()) {
	root := tview.NewFlex()
	root.SetTitle("Sacn-Neewer-Lite").SetTitleAlign(tview.AlignLeft).SetBorder(true)
	root.SetDirection(tview.FlexRow)

	ui.statusList = append(ui.statusList, NewStatusUiTuple(app_status_tracker, "Status", true, root, []string{}))
	ui.statusList = append(ui.statusList, NewStatusUiTuple(sacn_status_tracker, "sACN", true, root, []string{}))

	lights_container := tview.NewFlex()
	lights_container.SetTitle("Lights").SetTitleAlign(tview.AlignLeft).SetBorder(true)
	lights_container.SetDirection(tview.FlexColumn)
	root.AddItem(lights_container, 0, 1, false)

	for _, lightConfig := range config.Lights {
		universeText := "Universe: " + strconv.Itoa(int(lightConfig.Universe))
		addressText := "Address: " + strconv.Itoa(int(lightConfig.Address))
		ui.statusList = append(ui.statusList, NewStatusUiTuple(light_status_trackers[lightConfig.ID], lightConfig.ID, false, lights_container, []string{
			universeText,
			addressText,
		}))
	}

	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			cancel()
		}
		return event
	})

	ui.app.SetRoot(root, true)
}
