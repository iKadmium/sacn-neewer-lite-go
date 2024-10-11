package status

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

type Status struct {
	text          string
	color         tcell.Color
	event_counter *EventRateCounter
}

func NewStatus() Status {
	return Status{
		text:          "No data",
		color:         tcell.ColorReset,
		event_counter: NewEventRateCounter(time.Second),
	}
}

func (s *Status) GetText() string {
	return s.text
}

func (s *Status) GetColor() tcell.Color {
	return s.color
}

func (s *Status) Update(text string, color tcell.Color) {
	s.text = text
	s.color = color
}

func (s *Status) SetErr(err error) {
	s.Update(err.Error(), tcell.ColorRed)
}

func (s *Status) Increment() {
	s.event_counter.Increment()
}

func (s *Status) LastCount() int {
	return s.event_counter.LastCount()
}
