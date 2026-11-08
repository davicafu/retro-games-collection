// Package glide turns key presses into smooth movement. Terminals report key
// presses but not releases: holding a key sends a first press, a pause, then
// repeats. A Glide keeps moving for a few ticks after each press so repeats
// join into continuous motion, with no floating point.
package glide

// Unit is the number of speed points that move one cell.
const Unit = 100

// Glide moves at Speed cells per Unit ticks for Ticks ticks after each press.
// Speed must not exceed Unit, so it moves at most one cell per tick. Its zero
// state is at rest.
type Glide struct {
	Speed int
	Ticks int

	dir  int
	left int
	acc  int // progress towards the next cell, in Unit points
}

// Press starts or refreshes the glide in direction dir (-1 or 1). A fresh
// press, or one that reverses the direction, moves on the very next Step.
func (g *Glide) Press(dir int) {
	if g.left == 0 || dir != g.dir {
		g.acc = Unit - g.Speed
	}
	g.dir, g.left = dir, g.Ticks
}

// Step advances one tick and returns the move for it: -1, 0 or 1 cells.
func (g *Glide) Step() int {
	if g.left == 0 {
		return 0
	}
	g.left--
	g.acc += g.Speed
	if g.acc < Unit {
		return 0
	}
	g.acc -= Unit
	return g.dir
}

// Stop ends the glide at once.
func (g *Glide) Stop() { g.left = 0 }
