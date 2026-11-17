package polyomino

import (
	"reflect"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

func pt(x, y int) grid.Point { return grid.Point{X: x, Y: y} }

func TestRotateTurnsClockwiseInsideItsBox(t *testing.T) {
	// An L in a 3x3 box: the column on the left with a foot to the right.
	l := Shape{Size: 3, Cells: []grid.Point{pt(0, 0), pt(0, 1), pt(0, 2), pt(1, 2)}}
	got := l.Rotate()
	want := Shape{Size: 3, Cells: []grid.Point{pt(2, 0), pt(1, 0), pt(0, 0), pt(0, 1)}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rotate = %v, want %v", got, want)
	}
	back := l.Rotate().Rotate().Rotate().Rotate()
	if !reflect.DeepEqual(back, l) {
		t.Fatalf("four turns = %v, want the shape back %v", back, l)
	}
}

func TestFitsChecksBoundsAndFilledCells(t *testing.T) {
	w := NewWell(4, 3)
	bar := Shape{Size: 2, Cells: []grid.Point{pt(0, 0), pt(1, 0)}}
	if !w.Fits(bar, pt(2, 2)) {
		t.Fatalf("a bar must fit in the bottom right corner")
	}
	for _, at := range []grid.Point{pt(3, 0), pt(-1, 0), pt(0, 3)} {
		if w.Fits(bar, at) {
			t.Fatalf("a bar at %v must not fit outside the well", at)
		}
	}
	w.Place(bar, pt(0, 2), 5)
	if w.At(pt(1, 2)) != 5 || w.Fits(bar, pt(1, 2)) {
		t.Fatalf("placed cells must hold their value and block others")
	}
	if !w.Fits(bar, pt(0, -1)) {
		t.Fatalf("cells above the well are open, so shapes can enter from the top")
	}
}

func TestDropDistanceIsHowFarAShapeFalls(t *testing.T) {
	w := NewWell(3, 5)
	w.Place(Shape{Size: 1, Cells: []grid.Point{pt(0, 0)}}, pt(1, 3), 1)
	dot := Shape{Size: 1, Cells: []grid.Point{pt(0, 0)}}
	if got := w.DropDistance(dot, pt(1, 0)); got != 2 {
		t.Fatalf("drop over a filled cell = %d, want 2", got)
	}
	if got := w.DropDistance(dot, pt(0, 0)); got != 4 {
		t.Fatalf("drop to the floor = %d, want 4", got)
	}
}

func TestClearFullRowsDropsEverythingAbove(t *testing.T) {
	w := NewWell(2, 4)
	for _, c := range []struct {
		at grid.Point
		v  int
	}{{pt(0, 1), 1}, {pt(0, 2), 2}, {pt(1, 2), 2}, {pt(1, 3), 3}, {pt(0, 3), 3}} {
		w.Place(Shape{Size: 1, Cells: []grid.Point{pt(0, 0)}}, c.at, c.v)
	}
	if got := w.FullRows(); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Fatalf("full rows = %v, want [2 3]", got)
	}
	w.ClearRows(w.FullRows())
	if w.At(pt(0, 3)) != 1 || w.At(pt(0, 1)) != 0 || len(w.FullRows()) != 0 {
		t.Fatalf("after clearing, the lone cell must fall to the bottom: %v", w)
	}
}

func TestCloneIsIndependent(t *testing.T) {
	w := NewWell(2, 2)
	c := w.Clone()
	c.Place(Shape{Size: 1, Cells: []grid.Point{pt(0, 0)}}, pt(0, 0), 1)
	if w.At(pt(0, 0)) != 0 {
		t.Fatalf("placing on a clone changed the original")
	}
}
