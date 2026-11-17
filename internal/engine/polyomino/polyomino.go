// Package polyomino is spatial manipulation on a grid: shapes made of cells,
// their rotation, and a well that checks where a shape fits, keeps the
// shapes placed in it and clears its full rows.
package polyomino

import (
	"slices"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// Shape is a set of cells inside a Size by Size box, the box rotation turns
// them in.
type Shape struct {
	Size  int
	Cells []grid.Point
}

// Rotate returns the shape turned a quarter clockwise inside its box.
func (s Shape) Rotate() Shape {
	cells := make([]grid.Point, len(s.Cells))
	for i, c := range s.Cells {
		cells[i] = grid.Point{X: s.Size - 1 - c.Y, Y: c.X}
	}
	return Shape{Size: s.Size, Cells: cells}
}

// Well is a W by H grid of values, 0 for an empty cell. Rows above the top
// are open and hold nothing, so shapes can enter from above.
type Well struct {
	W, H  int
	cells [][]int
}

func NewWell(w, h int) *Well {
	cells := make([][]int, h)
	for y := range cells {
		cells[y] = make([]int, w)
	}
	return &Well{W: w, H: h, cells: cells}
}

// At is the value at p, 0 outside the well.
func (w *Well) At(p grid.Point) int {
	if p.X < 0 || p.X >= w.W || p.Y < 0 || p.Y >= w.H {
		return 0
	}
	return w.cells[p.Y][p.X]
}

// Fits reports whether s with its box at at lies within the well's sides
// and floor, on empty cells.
func (w *Well) Fits(s Shape, at grid.Point) bool {
	for _, c := range s.Cells {
		p := at.Add(c)
		if p.X < 0 || p.X >= w.W || p.Y >= w.H || w.At(p) != 0 {
			return false
		}
	}
	return true
}

// Place fills the cells s covers at at with v; cells above the top are
// dropped.
func (w *Well) Place(s Shape, at grid.Point, v int) {
	for _, c := range s.Cells {
		if p := at.Add(c); p.Y >= 0 && p.Y < w.H && p.X >= 0 && p.X < w.W {
			w.cells[p.Y][p.X] = v
		}
	}
}

// DropDistance is how many rows s at at can fall before it rests.
func (w *Well) DropDistance(s Shape, at grid.Point) int {
	d := 0
	for w.Fits(s, at.Add(grid.Point{Y: d + 1})) {
		d++
	}
	return d
}

// FullRows are the rows with no empty cell, top to bottom.
func (w *Well) FullRows() []int {
	var rows []int
	for y, row := range w.cells {
		if !slices.Contains(row, 0) {
			rows = append(rows, y)
		}
	}
	return rows
}

// ClearRows removes the given rows; the rows above fall into their place
// and empty rows enter at the top.
func (w *Well) ClearRows(rows []int) {
	kept := make([][]int, 0, w.H)
	for y, row := range w.cells {
		if !slices.Contains(rows, y) {
			kept = append(kept, row)
		}
	}
	for len(kept) < w.H {
		kept = slices.Insert(kept, 0, make([]int, w.W))
	}
	w.cells = kept
}

// Clone is an independent copy of the well.
func (w *Well) Clone() *Well {
	c := &Well{W: w.W, H: w.H, cells: make([][]int, w.H)}
	for y, row := range w.cells {
		c.cells[y] = slices.Clone(row)
	}
	return c
}
