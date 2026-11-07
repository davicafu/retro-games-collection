package scroll

import "testing"

func TestScrollerAdvancesOneCellPerAccumulatedUnit(t *testing.T) {
	var s Scroller
	got := 0
	for tick := 1; tick <= 10; tick++ {
		got += s.Advance(35)
		// 35 per tick: cells at ticks 3 (105), 6 (210), 9 (315).
		want := tick * 35 / Unit
		if got != want {
			t.Fatalf("tick %d: scrolled %d cells, want %d", tick, got, want)
		}
	}
}

func TestScrollerNeverMovesMoreThanOneCellBelowUnitSpeed(t *testing.T) {
	var s Scroller
	for tick := 0; tick < 1000; tick++ {
		if n := s.Advance(Unit - 1); n > 1 {
			t.Fatalf("tick %d scrolled %d cells", tick, n)
		}
	}
}

func TestScrollerZeroSpeedDoesNotMove(t *testing.T) {
	var s Scroller
	for i := 0; i < 5; i++ {
		if n := s.Advance(0); n != 0 {
			t.Fatalf("zero speed scrolled %d cells", n)
		}
	}
}

func TestRampRisesEveryStepDistanceAndCaps(t *testing.T) {
	r := Ramp{Base: 35, Step: 5, Every: 20, Max: 90}
	cases := map[int]int{0: 35, 19: 35, 20: 40, 39: 40, 40: 45, 219: 85, 220: 90, 10000: 90}
	for distance, want := range cases {
		if got := r.At(distance); got != want {
			t.Fatalf("At(%d) = %d, want %d", distance, got, want)
		}
	}
}
