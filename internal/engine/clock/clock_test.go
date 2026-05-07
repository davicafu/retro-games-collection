package clock

import (
	"testing"
	"time"
)

func TestManualAdvanceIsDeterministic(t *testing.T) {
	m := NewManual()

	m.Advance(5 * time.Millisecond)
	m.Advance(7 * time.Millisecond)

	first := <-m.Ticks()
	second := <-m.Ticks()

	if !first.Equal(time.Unix(0, 0).Add(5 * time.Millisecond)) {
		t.Fatalf("unexpected first tick: %v", first)
	}

	if !second.Equal(time.Unix(0, 0).Add(12 * time.Millisecond)) {
		t.Fatalf("unexpected second tick: %v", second)
	}
}
