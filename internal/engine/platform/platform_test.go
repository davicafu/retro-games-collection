package platform

import (
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// level is a map from rows: '=' solid, 'H' ladder, anything else empty.
type level []string

func (l level) Tile(p grid.Point) Tile {
	if p.Y < 0 || p.Y >= len(l) || p.X < 0 || p.X >= len(l[p.Y]) {
		return Solid
	}
	switch l[p.Y][p.X] {
	case '=':
		return Solid
	case 'H':
		return Ladder
	}
	return Empty
}

var sample = level{
	"     ",
	"  H  ",
	"==H= ",
	"  H  ",
	"=====",
}

func pt(x, y int) grid.Point { return grid.Point{X: x, Y: y} }

func TestSupportIsAFloorOrALadderBelowOrALadderHere(t *testing.T) {
	cases := map[grid.Point]bool{
		pt(0, 1): true,  // floor below
		pt(2, 0): true,  // on top of a ladder
		pt(2, 1): true,  // on a ladder
		pt(4, 1): false, // over the gap
		pt(4, 3): true,  // bottom floor
		pt(1, 3): true,
		pt(3, 0): false,
	}
	for p, want := range cases {
		if got := Supported(sample, p); got != want {
			t.Fatalf("Supported(%v) = %v, want %v", p, got, want)
		}
	}
}

func TestClimbingFollowsTheLadder(t *testing.T) {
	if !CanClimbUp(sample, pt(2, 1)) || CanClimbUp(sample, pt(1, 1)) {
		t.Fatalf("up only from a ladder cell into an open cell")
	}
	if !CanClimbDown(sample, pt(2, 0)) || !CanClimbDown(sample, pt(2, 2)) || CanClimbDown(sample, pt(2, 3)) || CanClimbDown(sample, pt(0, 1)) {
		t.Fatalf("down onto a ladder below, or along a ladder into an open cell, never into a floor")
	}
	if !Open(sample, pt(2, 2)) || Open(sample, pt(0, 2)) || Open(sample, pt(-1, 0)) {
		t.Fatalf("ladders and empty cells are open, floors and the outside are not")
	}
}
