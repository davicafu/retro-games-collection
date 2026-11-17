// Package platform is the structure of a platform level on a tile grid:
// which cells are solid, which are ladders, and the rules for standing,
// climbing and falling that every walker, climber and roller shares.
package platform

import "github.com/davidbarreiro/cli-games-collection/internal/engine/grid"

// Tile is what a cell of a platform level is made of.
type Tile int

const (
	Empty Tile = iota
	Solid
	Ladder
)

// Map is a platform level; cells outside it must read as Solid.
type Map interface {
	Tile(p grid.Point) Tile
}

var (
	up   = grid.Point{X: 0, Y: -1}
	down = grid.Point{X: 0, Y: 1}
)

// Open reports a cell something can be in: anything but Solid.
func Open(m Map, p grid.Point) bool { return m.Tile(p) != Solid }

// Supported reports whether something at p stands: on a ladder, or on a
// floor or a ladder's top just below. What is not supported falls.
func Supported(m Map, p grid.Point) bool {
	below := m.Tile(p.Add(down))
	return m.Tile(p) == Ladder || below == Solid || below == Ladder
}

// CanClimbUp reports whether a climber at p can go up a cell: from a
// ladder into an open cell.
func CanClimbUp(m Map, p grid.Point) bool {
	return m.Tile(p) == Ladder && Open(m, p.Add(up))
}

// CanClimbDown reports whether a climber at p can go down a cell: onto a
// ladder below, or along a ladder into an open cell.
func CanClimbDown(m Map, p grid.Point) bool {
	below := p.Add(down)
	return m.Tile(below) == Ladder || m.Tile(p) == Ladder && Open(m, below)
}
