// Package scroll moves a side-scrolling world at an integer speed, with no
// floating point, so simulations stay deterministic.
package scroll

// Unit is the number of speed points that scroll the world by one cell.
// Speeds are measured in cells per Unit ticks; a speed below Unit scrolls at
// most one cell per tick, so nothing can tunnel through the player.
const Unit = 100

// Scroller accumulates speed tick by tick and reports whole cells to scroll.
// Its zero value is ready to use.
type Scroller struct {
	accumulator int
}

// Advance adds one tick at the given speed and returns how many cells the
// world scrolls this tick.
func (s *Scroller) Advance(speed int) int {
	s.accumulator += speed
	cells := s.accumulator / Unit
	s.accumulator %= Unit
	return cells
}

// Ramp raises the speed by Step every Every cells travelled, from Base up to Max.
type Ramp struct {
	Base  int
	Step  int
	Every int
	Max   int
}

// At returns the speed after travelling distance cells.
func (r Ramp) At(distance int) int {
	return min(r.Base+r.Step*(distance/r.Every), r.Max)
}
