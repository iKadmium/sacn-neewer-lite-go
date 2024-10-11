package status

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

type Status struct {
	text         string
	textColor    tcell.Color
	eventCounter *EventRateCounter
	colorTracker *ColorTracker
}

func NewStatus(withEventCounter bool, withColorTracker bool) Status {
	var eventCounter *EventRateCounter
	if withEventCounter {
		eventCounter = NewEventRateCounter(time.Second)
	}

	var colorTracker *ColorTracker
	if withColorTracker {
		colorTracker = NewColorTracker()
	}

	return Status{
		text:         "No data",
		textColor:    tcell.ColorReset,
		eventCounter: eventCounter,
		colorTracker: colorTracker,
	}
}

func (s *Status) GetText() string {
	return s.text
}

func (s *Status) GetTextColor() tcell.Color {
	return s.textColor
}

func (s *Status) Update(text string, color tcell.Color) {
	s.text = text
	s.textColor = color
}

func (s *Status) SetErr(err error) {
	s.Update(err.Error(), tcell.ColorRed)
}

func (s *Status) Increment() {
	s.eventCounter.Increment()
}

func (s *Status) LastCount() int {
	return s.eventCounter.LastCount()
}

func (s *Status) HasUpdateRate() bool {
	return s.eventCounter != nil
}

func (s *Status) GetColorTracker() *ColorTracker {
	return s.colorTracker
}
