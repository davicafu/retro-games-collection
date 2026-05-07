package clock

import "time"

type Clock interface {
	Ticks() <-chan time.Time
}

type Manual struct {
	ch  chan time.Time
	now time.Time
}

func NewManual() *Manual {
	return &Manual{ch: make(chan time.Time, 32), now: time.Unix(0, 0)}
}

func (m *Manual) Ticks() <-chan time.Time { return m.ch }
func (m *Manual) Advance(d time.Duration) {
	m.now = m.now.Add(d)
	m.ch <- m.now
}

type Ticker struct {
	ticker *time.Ticker
}

func NewTicker(interval time.Duration) *Ticker {
	return &Ticker{ticker: time.NewTicker(interval)}
}

func (t *Ticker) Ticks() <-chan time.Time { return t.ticker.C }

func (t *Ticker) Stop() {
	t.ticker.Stop()
}
