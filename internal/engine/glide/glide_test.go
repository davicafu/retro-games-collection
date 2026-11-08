package glide

import (
	"reflect"
	"testing"
)

// moves presses dir once (0 for no press) per entry and records each tick's move.
func moves(g *Glide, presses ...int) []int {
	out := make([]int, 0, len(presses))
	for _, dir := range presses {
		if dir != 0 {
			g.Press(dir)
		}
		out = append(out, g.Step())
	}
	return out
}

func TestAPressMovesOnItsOwnTickThenGlidesAndStops(t *testing.T) {
	g := Glide{Speed: 50, Ticks: 4}
	got := moves(&g, 1, 0, 0, 0, 0, 0)
	want := []int{1, 0, 1, 0, 0, 0}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("moves = %v, want %v", got, want)
	}
}

func TestKeyRepeatsKeepTheGlideGoing(t *testing.T) {
	g := Glide{Speed: 50, Ticks: 2}
	got := moves(&g, -1, 0, -1, 0, -1, 0, 0, 0)
	want := []int{-1, 0, -1, 0, -1, 0, 0, 0}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("moves = %v, want %v", got, want)
	}
}

func TestAnOppositePressReversesAtOnce(t *testing.T) {
	g := Glide{Speed: 30, Ticks: 10}
	if got := moves(&g, 1, -1); !reflect.DeepEqual(got, []int{1, -1}) {
		t.Fatalf("moves = %v, want [1 -1]", got)
	}
}

func TestStopEndsTheGlide(t *testing.T) {
	g := Glide{Speed: Unit, Ticks: 10}
	g.Press(1)
	g.Step()
	g.Stop()
	if got := g.Step(); got != 0 {
		t.Fatalf("move after Stop = %d, want 0", got)
	}
}

func TestFullSpeedMovesEveryTick(t *testing.T) {
	g := Glide{Speed: Unit, Ticks: 3}
	if got := moves(&g, 1, 0, 0, 0); !reflect.DeepEqual(got, []int{1, 1, 1, 0}) {
		t.Fatalf("moves = %v, want [1 1 1 0]", got)
	}
}
