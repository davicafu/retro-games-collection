package defrag

import (
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/polyomino"
)

func pt(x, y int) grid.Point { return grid.Point{X: x, Y: y} }

func newTestGame(kinds ...Kind) *Game {
	return NewGame(Config{Rand: rand.New(rand.NewSource(1)), Sequence: kinds})
}

func steps(g *Game, n int, in Input) {
	for i := 0; i < n; i++ {
		g.Step(in)
	}
}

// cells are the disk cells the falling fragment covers.
func cells(g *Game) []grid.Point {
	f := g.Fragment()
	var out []grid.Point
	for _, c := range f.Shape.Cells {
		out = append(out, f.Pos.Add(c))
	}
	slices.SortFunc(out, func(a, b grid.Point) int { return a.Y*100 + a.X - b.Y*100 - b.X })
	return out
}

// fill fills row y except the columns in gaps.
func fill(g *Game, y int, gaps ...int) {
	for x := 0; x < Width; x++ {
		if !slices.Contains(gaps, x) {
			g.disk.Place(polyomino.Shape{Size: 1, Cells: []grid.Point{{}}}, pt(x, y), int(Square)+1)
		}
	}
}

func TestFragmentsAppearAtTheTopCentreAndTheNextIsKnown(t *testing.T) {
	g := newTestGame(Bar, Tee)
	if got := cells(g); !slices.Equal(got, []grid.Point{pt(3, 0), pt(4, 0), pt(5, 0), pt(6, 0)}) {
		t.Fatalf("bar at %v, want across the top centre", got)
	}
	if g.Next() != Tee {
		t.Fatalf("next = %v, want Tee", g.Next())
	}
}

func TestEverySetOfSevenDealsEachShapeOnce(t *testing.T) {
	g := NewGame(Config{Rand: rand.New(rand.NewSource(7))})
	for set := 0; set < 5; set++ {
		seen := map[Kind]bool{}
		for i := 0; i < 7; i++ {
			seen[g.Fragment().Kind] = true
			g.Step(Input{HardDrop: true})
			steps(g, ClearTicks, Input{})
			g.disk = polyomino.NewWell(Width, Height) // keep room
		}
		if len(seen) != 7 {
			t.Fatalf("set %d dealt %v, want all seven shapes", set+1, seen)
		}
	}
}

func TestShiftMovesACellPerTapAndStopsAtTheSides(t *testing.T) {
	g := newTestGame(Square)
	g.Step(Input{Shift: -1})
	if x := cells(g)[0].X; x != 3 {
		t.Fatalf("a tap must shift one cell at once, at x %d", x)
	}
	steps(g, 10, Input{})
	if x := cells(g)[0].X; x != 3 {
		t.Fatalf("a tap must shift exactly one cell, at x %d", x)
	}
	for i := 0; i < 40; i++ {
		g.Step(Input{Shift: 1})
		g.Step(Input{})
	}
	if x := cells(g)[len(cells(g))-1].X; x != Width-1 {
		t.Fatalf("held right must reach the right side, at x %d", x)
	}
}

func TestRotateTurnsClockwiseAndNudgesOffTheSide(t *testing.T) {
	g := newTestGame(Tee)
	g.Step(Input{Rotate: true})
	if got := cells(g); !slices.Equal(got, []grid.Point{pt(4, 0), pt(4, 1), pt(5, 1), pt(4, 2)}) {
		t.Fatalf("tee after a turn = %v", got)
	}
	g = newTestGame(Bar)
	g.fragment.Shape = g.fragment.Shape.Rotate() // standing up
	g.fragment.Pos = pt(-2, 0)                   // cells in column 0
	g.Step(Input{Rotate: true})
	for _, c := range cells(g) {
		if c.X < 0 || c.X >= Width {
			t.Fatalf("rotation by the side left cells out: %v", cells(g))
		}
	}
	if cells(g)[0].Y != cells(g)[3].Y {
		t.Fatalf("the bar must lie flat after the nudge, got %v", cells(g))
	}
}

func TestGravityDropsARowPerIntervalAndLevelsSpeedItUp(t *testing.T) {
	g := newTestGame(Square)
	steps(g, GravityTicks-1, Input{})
	if y := cells(g)[0].Y; y != 0 {
		t.Fatalf("fell to row %d before the interval", y)
	}
	g.Step(Input{})
	if y := cells(g)[0].Y; y != 1 {
		t.Fatalf("at row %d after the interval, want 1", y)
	}
	if g.gravityTicks() != GravityTicks {
		t.Fatalf("level 1 gravity %d", g.gravityTicks())
	}
	g.level = 3
	three := g.gravityTicks()
	g.level = 50
	if three >= GravityTicks || g.gravityTicks() != MinGravityTicks {
		t.Fatalf("gravity level 3 %d, level 50 %d", three, g.gravityTicks())
	}
}

func TestSoftDropFallsFasterAndScoresARow(t *testing.T) {
	g := newTestGame(Square)
	g.Step(Input{SoftDrop: true})
	if y := cells(g)[0].Y; y != 1 || g.Score() != PointsSoftDrop {
		t.Fatalf("soft drop at row %d score %d, want one row for %d", y, g.Score(), PointsSoftDrop)
	}
}

func TestHardDropLandsLocksAndDealsTheNext(t *testing.T) {
	g := newTestGame(Square, Bar)
	g.Step(Input{HardDrop: true})
	if g.disk.At(pt(4, Height-1)) == 0 || g.disk.At(pt(5, Height-2)) == 0 {
		t.Fatalf("the square must lock at the bottom")
	}
	if g.Score() != PointsHardDrop*(Height-2) || g.Fragment().Kind != Bar {
		t.Fatalf("score %d fragment %v, want %d and the bar next", g.Score(), g.Fragment().Kind, PointsHardDrop*(Height-2))
	}
}

func TestARestingFragmentLocksAfterTheDelayAndMovesRestartIt(t *testing.T) {
	g := newTestGame(Square, Square)
	g.fragment.Pos = pt(4, Height-2) // resting on the floor
	steps(g, LockTicks-1, Input{})
	g.Step(Input{Shift: 1})
	steps(g, LockTicks-1, Input{})
	if g.disk.At(pt(5, Height-1)) != 0 {
		t.Fatalf("a move must restart the lock delay")
	}
	g.Step(Input{})
	if g.disk.At(pt(5, Height-1)) == 0 {
		t.Fatalf("the fragment must lock once the delay runs out")
	}
}

func TestLockRestartsAreLimited(t *testing.T) {
	g := newTestGame(Square, Square)
	g.fragment.Pos = pt(4, Height-2)
	for i := 0; i < 200 && g.disk.At(pt(4, Height-1)) == 0 && g.disk.At(pt(5, Height-1)) == 0 && g.disk.At(pt(6, Height-1)) == 0; i++ {
		g.Step(Input{Rotate: true})
		steps(g, LockTicks/2, Input{})
	}
	locked := 0
	for x := 0; x < Width; x++ {
		if g.disk.At(pt(x, Height-1)) != 0 {
			locked++
		}
	}
	if locked != 2 {
		t.Fatalf("endless rotating must not stop the fragment from locking")
	}
}

func TestFullRowsFlashThenClearAndScoreByLevel(t *testing.T) {
	g := newTestGame(Bar, Square)
	fill(g, Height-1, 3, 4, 5, 6)
	fill(g, Height-2, 0, 3, 4, 5, 6)
	g.Step(Input{HardDrop: true})
	if !slices.Equal(g.Clearing(), []int{Height - 1}) {
		t.Fatalf("clearing %v, want the bottom row flashing", g.Clearing())
	}
	score := g.Score()
	if score != PointsHardDrop*(Height-1)+100 {
		t.Fatalf("score %d, want the drop plus 100", score)
	}
	steps(g, ClearTicks-1, Input{Shift: 1})
	if len(g.Clearing()) != 1 {
		t.Fatalf("the flash must last %d ticks", ClearTicks)
	}
	g.Step(Input{})
	if len(g.Clearing()) != 0 || g.Sectors() != 1 || g.disk.At(pt(1, Height-1)) == 0 || g.disk.At(pt(0, Height-1)) != 0 {
		t.Fatalf("clearing %v sectors %d, want the row gone and the row above fallen into its place", g.Clearing(), g.Sectors())
	}
	if g.Fragment().Kind != Square {
		t.Fatalf("the next fragment must appear after the clear")
	}
}

func TestClearingFourRowsAtOnceScoresMostAndLevelsUp(t *testing.T) {
	g := newTestGame(Bar)
	g.sectors = SectorsPerLevel - 1
	for y := Height - 4; y < Height; y++ {
		fill(g, y, 0)
	}
	g.fragment.Shape = g.fragment.Shape.Rotate()
	g.fragment.Pos = pt(-2, 0)
	g.Step(Input{HardDrop: true})
	if want := PointsHardDrop*(Height-4) + 800; g.Score() != want {
		t.Fatalf("score %d, want %d", g.Score(), want)
	}
	steps(g, ClearTicks, Input{})
	if g.Sectors() != SectorsPerLevel+3 || g.Level() != 2 {
		t.Fatalf("sectors %d level %d, want %d and level 2", g.Sectors(), g.Level(), SectorsPerLevel+3)
	}
}

func TestTheRoundEndsWhenAFragmentCannotAppear(t *testing.T) {
	g := newTestGame(Square, Square)
	fill(g, 1, 0)
	g.fragment.Pos = pt(0, -5) // out of the way
	g.lockFragment()
	if !g.Over() {
		t.Fatalf("a blocked top must end the round")
	}
	score := g.Score()
	g.Step(Input{HardDrop: true})
	if g.Score() != score {
		t.Fatalf("a finished round must not advance")
	}
}

func TestRealTimeSpeedsDoNotDependOnTheTickRate(t *testing.T) {
	perSecond := int(time.Second / TickPeriod)
	if got := (time.Duration(GravityTicks) * TickPeriod).Round(time.Millisecond); got != 800*time.Millisecond {
		t.Fatalf("level 1 gravity %v a row, want 800ms", got)
	}
	if got := perSecond / MinGravityTicks; got != 15 {
		t.Fatalf("top gravity %d rows/s, want 15", got)
	}
	if SectorsPerLevel != 8 {
		t.Fatalf("a level every %d sectors, want 8", SectorsPerLevel)
	}
	g := newTestGame(Square)
	g.level = 8
	eight := g.gravityTicks()
	g.level = 9
	if eight <= MinGravityTicks || g.gravityTicks() != MinGravityTicks {
		t.Fatalf("gravity level 8 %d, level 9 %d, want the top speed first on level 9", eight, g.gravityTicks())
	}
	if got := (time.Duration(LockTicks) * TickPeriod).Round(time.Millisecond); got != 500*time.Millisecond {
		t.Fatalf("lock delay %v, want 500ms", got)
	}
	if got := perSecond * ShiftSpeed / 100; got != 15 {
		t.Fatalf("shift %d cells/s, want 15", got)
	}
}

func TestRowsScoreTimesTheLevel(t *testing.T) {
	g := newTestGame(Bar)
	g.level = 3
	fill(g, Height-1, 3, 4, 5, 6)
	g.Step(Input{HardDrop: true})
	if want := PointsHardDrop*(Height-1) + 3*100; g.Score() != want {
		t.Fatalf("score %d, want %d", g.Score(), want)
	}
}

// TestFallingOffALedgeGivesAFreshLockDelay: once the restarts are used up,
// a fragment that slides off a ledge still waits the whole delay where it
// lands.
func TestFallingOffALedgeGivesAFreshLockDelay(t *testing.T) {
	g := newTestGame(Square, Square)
	g.disk.Place(polyomino.Shape{Size: 1, Cells: []grid.Point{{}}}, pt(4, Height-2), 1)
	g.fragment.Pos = pt(4, Height-4) // resting on the ledge
	g.lockResets, g.lockClock = MaxLockResets, LockTicks-2
	g.Step(Input{Shift: 1}) // off the ledge
	for i := 0; i < 3*GravityTicks && g.Fragment().Pos.Y < Height-2; i++ {
		g.Step(Input{})
	}
	if g.Fragment().Pos.Y != Height-2 {
		t.Fatalf("the fragment did not land on the floor, at %v", g.Fragment().Pos)
	}
	steps(g, LockTicks-2, Input{})
	if g.At(pt(5, Height-1)) != 0 {
		t.Fatalf("the fragment locked before a full delay where it landed")
	}
}

func TestTheDealIsShuffled(t *testing.T) {
	inOrder := 0
	for seed := int64(1); seed <= 5; seed++ {
		g := NewGame(Config{Rand: rand.New(rand.NewSource(seed))})
		if g.Fragment().Kind == Bar && g.Next() == Square {
			inOrder++
		}
	}
	if inOrder == 5 {
		t.Fatalf("every game deals the shapes in the same order")
	}
}
