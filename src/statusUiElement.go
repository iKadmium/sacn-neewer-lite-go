package main

import (
	"sacn_neewer_lite_go/status"

	"github.com/rivo/tview"
)

type StatusUiTuple struct {
	textView       *tview.TextView
	status         *status.Status
	updateRateView *tview.TextView
}

func NewStatusUiTuple(status *status.Status, text_view *tview.TextView, update_rate_view *tview.TextView) *StatusUiTuple {
	return &StatusUiTuple{
		status:         status,
		textView:       text_view,
		updateRateView: update_rate_view,
	}
}
