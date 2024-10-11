package status

import (
	"sync"
	"time"
)

type EventRateCounter struct {
	duration   time.Duration
	count      int
	mu         sync.Mutex
	ticker     *time.Ticker
	last_count int
}

func NewEventRateCounter(duration time.Duration) *EventRateCounter {
	erc := &EventRateCounter{
		duration:   duration,
		ticker:     time.NewTicker(duration),
		last_count: 0,
	}

	go func() {
		for range erc.ticker.C {
			erc.reset()
		}
	}()

	return erc
}

func (erc *EventRateCounter) Increment() {
	erc.mu.Lock()
	defer erc.mu.Unlock()
	erc.count++
}

func (erc *EventRateCounter) LastCount() int {
	return erc.last_count
}

func (erc *EventRateCounter) reset() {
	erc.mu.Lock()
	defer erc.mu.Unlock()
	erc.last_count = erc.count
	erc.count = 0
}

func (erc *EventRateCounter) Stop() {
	erc.ticker.Stop()
}
