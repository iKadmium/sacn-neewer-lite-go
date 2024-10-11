package main

import (
	"sacn_neewer_lite_go/status"
	"strconv"

	"github.com/rivo/tview"
)

type StatusUiTuple struct {
	textView       *tview.TextView
	status         *status.Status
	updateRateView *tview.TextView
	colorBox       *tview.Box
}

func NewStatusUiTuple(status *status.Status, title string, fixedSize bool, parent *tview.Flex, extraText []string) *StatusUiTuple {
	rows := 3

	container := tview.NewFlex()
	container.SetTitle(title).SetTitleAlign(tview.AlignLeft).SetBorder(true)
	container.SetDirection(tview.FlexRow)

	status_text := tview.NewTextView()
	status_text.SetText("Starting")
	container.AddItem(status_text, 0, 1, false)

	for _, text := range extraText {
		extra := tview.NewTextView()
		extra.SetText(text)
		container.AddItem(extra, 0, 1, false)
		rows++
	}

	var updateRate *tview.TextView
	if status.HasUpdateRate() {
		updateRate = tview.NewTextView()
		container.AddItem(updateRate, 0, 1, false)
		rows++
	}

	var colorBox *tview.Box
	if status.GetColorTracker() != nil {
		colorBox = tview.NewBox()
		container.AddItem(colorBox, 0, 1, false)
		rows++
	}

	if fixedSize {
		parent.AddItem(container, rows, 1, false)
	} else {
		parent.AddItem(container, 0, 1, false)
	}

	return &StatusUiTuple{
		status:         status,
		textView:       status_text,
		colorBox:       colorBox,
		updateRateView: updateRate,
	}
}

func (s *StatusUiTuple) Update(app *tview.Application) {
	app.QueueUpdateDraw(func() {
		s.textView.SetText(s.status.GetText())
		s.textView.SetTextColor(s.status.GetTextColor())
		if s.updateRateView != nil {
			s.updateRateView.SetText("Update Rate: " + strconv.Itoa(s.status.LastCount()))
		}

		if s.status.GetColorTracker() != nil {
			s.colorBox.SetBackgroundColor(s.status.GetColorTracker().GetColor())
		}
	})
}
