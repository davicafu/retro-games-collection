// Package defrag is the falling-block puzzle inspired by Tetris: fragments of
// files fall onto a hard disk, and every full row is a defragmented sector
// that frees up.
package defrag

import (
	"slices"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/glide"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/polyomino"
)

// TickPeriod is the real time one Step stands for; every per-tick constant
// below is tuned to it.
const TickPeriod = time.Second / 60

// Width and Height are the disk's size in cells.
const (
	Width  = 10
	Height = 20
)

const (
	// A fragment falls a row every GravityTicks ticks on level 1 (0.8 s),
	// GravityStep fewer each level, down to MinGravityTicks (15 rows/s) from
	// level 9 on.
	GravityTicks    = 48
	GravityStep     = 6
	MinGravityTicks = 4
	// LockTicks is how long a resting fragment waits before it locks: 0.5 s.
	// A move or rotation while resting restarts the wait, at most
	// MaxLockResets times.
	LockTicks     = 30
	MaxLockResets = 15

	// ShiftSpeed and DropSpeed are how fast held keys shift (15 cells/s)
	// and soft drop (20 rows/s), in cells per glide.Unit ticks; KeyGlide is
	// how long a press keeps going, so a tap moves exactly one cell.
	ShiftSpeed = 25
	DropSpeed  = 33
	KeyGlide   = 3

	// ClearTicks is how long full rows flash before they free up: 0.3 s.
	ClearTicks      = 18
	SectorsPerLevel = 8

	PointsSoftDrop = 1
	PointsHardDrop = 2
)

// pointsForRows is the score of clearing 1 to 4 rows at once, times the level.
var pointsForRows = [...]int{0, 100, 300, 500, 800}

// Kind is one of the seven shapes of four cells.
type Kind int

const (
	Bar Kind = iota
	Square
	Tee
	Ess
	Zed
	Jay
	Ell
	kinds
)

// shapes are each kind's cells in its rotation box, as it appears.
var shapes = [kinds]polyomino.Shape{
	Bar:    {Size: 4, Cells: []grid.Point{{X: 0, Y: 1}, {X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1}}},
	Square: {Size: 2, Cells: []grid.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}}},
	Tee:    {Size: 3, Cells: []grid.Point{{X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 2, Y: 1}}},
	Ess:    {Size: 3, Cells: []grid.Point{{X: 1, Y: 0}, {X: 2, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}}},
	Zed:    {Size: 3, Cells: []grid.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 1}}},
	Jay:    {Size: 3, Cells: []grid.Point{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 2, Y: 1}}},
	Ell:    {Size: 3, Cells: []grid.Point{{X: 2, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: 2, Y: 1}}},
}

// kicks are the nudges a rotation tries, in order, when the turned
// fragment would overlap: none, sideways up to two cells, then up a row.
var kicks = []grid.Point{{}, {X: -1}, {X: 1}, {X: -2}, {X: 2}, {Y: -1}}

// Rand is the randomness the deal needs; *math/rand.Rand satisfies it.
type Rand interface {
	Intn(n int) int
}

// Input is the player's intent for one tick: Shift is a press of left (-1)
// or right (1), or 0.
type Input struct {
	Shift    int
	Rotate   bool
	SoftDrop bool
	HardDrop bool
}

// Fragment is the falling fragment: its shape as turned, with its rotation
// box at Pos.
type Fragment struct {
	Kind  Kind
	Shape polyomino.Shape
	Pos   grid.Point
}

// Config builds a Game. Sequence, when set, replaces the shuffled deal with
// those kinds, in turn.
type Config struct {
	Rand     Rand
	Sequence []Kind
}

// Game is the headless Defrag simulation, advanced one tick at a time.
type Game struct {
	rng      Rand
	sequence []Kind
	queue    []Kind

	disk     *polyomino.Well
	fragment Fragment
	shift    glide.Glide
	shiftDir int
	drop     glide.Glide

	gravityClock int
	lockClock    int
	lockResets   int
	clearing     []int
	clearClock   int

	score   int
	sectors int
	level   int
	over    bool
}

func NewGame(cfg Config) *Game {
	g := &Game{
		rng: cfg.Rand, sequence: cfg.Sequence,
		disk:  polyomino.NewWell(Width, Height),
		shift: glide.Glide{Speed: ShiftSpeed, Ticks: KeyGlide},
		drop:  glide.Glide{Speed: DropSpeed, Ticks: KeyGlide},
		level: 1,
	}
	g.spawn()
	return g
}

// deal tops up the queue: the next kinds of the sequence, or a shuffled
// set of all seven.
func (g *Game) deal() {
	for len(g.queue) < 2 {
		if len(g.sequence) > 0 {
			g.queue = append(g.queue, g.sequence[0])
			g.sequence = append(g.sequence[1:], g.sequence[0])
			continue
		}
		set := make([]Kind, kinds)
		for i := range set {
			set[i] = Kind(i)
		}
		for i := len(set) - 1; i > 0; i-- {
			j := g.rng.Intn(i + 1)
			set[i], set[j] = set[j], set[i]
		}
		g.queue = append(g.queue, set...)
	}
}

// spawn brings the next fragment in at the top centre; when it does not
// fit, the disk is full and the round ends.
func (g *Game) spawn() {
	g.deal()
	k := g.queue[0]
	g.queue = g.queue[1:]
	g.deal()
	s := shapes[k]
	top := s.Size
	for _, c := range s.Cells {
		top = min(top, c.Y)
	}
	g.fragment = Fragment{Kind: k, Shape: s, Pos: grid.Point{X: (Width - s.Size) / 2, Y: -top}}
	g.gravityClock, g.lockClock, g.lockResets = 0, 0, 0
	g.shift.Stop()
	g.drop.Stop()
	if !g.disk.Fits(s, g.fragment.Pos) {
		g.over = true
	}
}

// Step advances the simulation by one tick. A finished round does not
// advance.
func (g *Game) Step(in Input) {
	if g.over {
		return
	}
	if len(g.clearing) > 0 {
		if g.clearClock++; g.clearClock >= ClearTicks {
			g.sectors += len(g.clearing)
			g.level = 1 + g.sectors/SectorsPerLevel
			g.disk.ClearRows(g.clearing)
			g.clearing = nil
			g.spawn()
		}
		return
	}
	f := &g.fragment
	moved := false
	if in.Rotate {
		turned := f.Shape.Rotate()
		for _, k := range kicks {
			if at := f.Pos.Add(k); g.disk.Fits(turned, at) {
				f.Shape, f.Pos, moved = turned, at, true
				break
			}
		}
	}
	if in.Shift != 0 {
		if in.Shift != g.shiftDir {
			g.shift.Stop()
			g.shiftDir = in.Shift
		}
		g.shift.Press(in.Shift)
	}
	if d := g.shift.Step(); d != 0 && g.disk.Fits(f.Shape, f.Pos.Add(grid.Point{X: d})) {
		f.Pos.X += d
		moved = true
	}
	if in.HardDrop {
		d := g.disk.DropDistance(f.Shape, f.Pos)
		f.Pos.Y += d
		g.score += PointsHardDrop * d
		g.lockFragment()
		return
	}
	if in.SoftDrop {
		g.drop.Press(1)
	}
	if g.drop.Step() != 0 && g.fits(0, 1) {
		f.Pos.Y++
		g.score += PointsSoftDrop
		g.gravityClock = 0
	}
	if g.gravityClock++; g.gravityClock >= g.gravityTicks() {
		g.gravityClock = 0
		if g.fits(0, 1) {
			f.Pos.Y++
		}
	}
	if g.fits(0, 1) {
		g.lockClock = 0
		return
	}
	if moved && g.lockResets < MaxLockResets {
		g.lockResets++
		g.lockClock = 0
		return
	}
	if g.lockClock++; g.lockClock >= LockTicks {
		g.lockFragment()
	}
}

func (g *Game) fits(dx, dy int) bool {
	return g.disk.Fits(g.fragment.Shape, g.fragment.Pos.Add(grid.Point{X: dx, Y: dy}))
}

// lockFragment places the fragment on the disk and scores any rows it
// fills, which then flash before they clear; otherwise the next one comes.
func (g *Game) lockFragment() {
	g.disk.Place(g.fragment.Shape, g.fragment.Pos, int(g.fragment.Kind)+1)
	if rows := g.disk.FullRows(); len(rows) > 0 {
		g.score += pointsForRows[len(rows)] * g.level
		g.clearing, g.clearClock = rows, 0
		return
	}
	g.spawn()
}

// gravityTicks is how many ticks a row takes to fall on the current level.
func (g *Game) gravityTicks() int {
	return max(GravityTicks-GravityStep*(g.level-1), MinGravityTicks)
}

// Ghost is where the fragment would land with a hard drop.
func (g *Game) Ghost() grid.Point {
	return g.fragment.Pos.Add(grid.Point{Y: g.disk.DropDistance(g.fragment.Shape, g.fragment.Pos)})
}

// At is what fills disk cell p: 0 for empty, or the kind that locked there
// plus one.
func (g *Game) At(p grid.Point) int { return g.disk.At(p) }

func (g *Game) Fragment() Fragment {
	f := g.fragment
	f.Shape.Cells = slices.Clone(f.Shape.Cells)
	return f
}
func (g *Game) Next() Kind                   { return g.queue[0] }
func (g *Game) Clearing() []int              { return slices.Clone(g.clearing) }
func (g *Game) ClearClock() int              { return g.clearClock }
func (g *Game) Score() int                   { return g.score }
func (g *Game) Sectors() int                 { return g.sectors }
func (g *Game) Level() int                   { return g.level }
func (g *Game) Over() bool                   { return g.over }
func (g *Game) Shape(k Kind) polyomino.Shape { return shapes[k] }
