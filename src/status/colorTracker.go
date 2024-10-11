package status

import "github.com/gdamore/tcell/v2"

type ColorTracker struct {
	color tcell.Color
}

func NewColorTracker() *ColorTracker {
	return &ColorTracker{color: tcell.ColorBlack}
}

func (ct *ColorTracker) GetColor() tcell.Color {
	return ct.color
}

func (ct *ColorTracker) Update(color tcell.Color) {
	ct.color = color
}
