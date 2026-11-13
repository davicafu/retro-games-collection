package grid

import (
	"reflect"
	"testing"
)

// mazeOf builds a passable function from rows where '#' blocks.
func mazeOf(rows ...string) (int, int, func(Point) bool) {
	return len(rows[0]), len(rows), func(p Point) bool { return rows[p.Y][p.X] != '#' }
}

func TestDistancesCountStepsAroundWalls(t *testing.T) {
	w, h, open := mazeOf(
		"...",
		"##.",
		"...",
	)
	got := Distances(w, h, Point{X: 0, Y: 2}, open)
	want := [][]int{
		{6, 5, 4},
		{Unreachable, Unreachable, 3},
		{0, 1, 2},
	}
	cells := make([][]int, h)
	for y := range cells {
		for x := 0; x < w; x++ {
			cells[y] = append(cells[y], got.At(Point{X: x, Y: y}))
		}
	}
	if !reflect.DeepEqual(cells, want) {
		t.Fatalf("distances = %v, want %v", cells, want)
	}
}

func TestDistancesMarkSealedAndOutsideCellsUnreachable(t *testing.T) {
	w, h, open := mazeOf(
		".#.",
	)
	got := Distances(w, h, Point{X: 0, Y: 0}, open)
	if got.At(Point{X: 2, Y: 0}) != Unreachable {
		t.Fatalf("sealed cell reachable")
	}
	for _, p := range []Point{{X: -1, Y: 0}, {X: 3, Y: 0}, {X: 0, Y: 1}} {
		if got.At(p) != Unreachable {
			t.Fatalf("outside cell %v reachable", p)
		}
	}
}

func TestDistancesFromABlockedTargetReachNothing(t *testing.T) {
	w, h, open := mazeOf("#.")
	if got := Distances(w, h, Point{X: 0, Y: 0}, open); got.At(Point{X: 1, Y: 0}) != Unreachable {
		t.Fatalf("a blocked target must reach nothing")
	}
}
