package defrag

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/polyomino"
)

// Viability (S7): an autopilot that presses one key every ActTicks clears
// ViableSectors rows in at least ViableSeeds of AutopilotSeeds games.
const (
	ViableSectors  = 100
	AutopilotSeeds = 30
	ViableSeeds    = 27
	ActTicks       = 6
)

func TestAutopilotDefragmentsTheDisk(t *testing.T) {
	won := make([]bool, AutopilotSeeds)
	t.Run("games", func(t *testing.T) {
		for i := range won {
			t.Run(fmt.Sprintf("seed %d", i+1), func(t *testing.T) {
				t.Parallel()
				g := NewGame(Config{Rand: rand.New(rand.NewSource(int64(i + 1)))})
				for tick := 0; tick < 60*60*30 && !g.Over() && g.Sectors() < ViableSectors; tick++ {
					var in Input
					if tick%ActTicks == 0 {
						in = autopilot(g)
					}
					g.Step(in)
				}
				won[i] = g.Sectors() >= ViableSectors
				if !won[i] {
					t.Logf("sectors %d, level %d, over %v", g.Sectors(), g.Level(), g.Over())
				}
			})
		}
	})
	cleared := 0
	for _, w := range won {
		if w {
			cleared++
		}
	}
	if cleared < ViableSeeds {
		t.Fatalf("autopilot cleared %d rows in %d of %d games, want at least %d", ViableSectors, cleared, AutopilotSeeds, ViableSeeds)
	}
	t.Logf("autopilot cleared %d rows in %d of %d games", ViableSectors, cleared, AutopilotSeeds)
}

// autopilot presses the next key towards the best landing for the falling
// fragment: rotate until it is turned right, shift until it is above the
// spot, then drop it.
func autopilot(g *Game) Input {
	if len(g.Clearing()) > 0 {
		return Input{}
	}
	f := g.Fragment()
	turns, x, ok := bestLanding(g)
	if !ok {
		return Input{HardDrop: true}
	}
	switch {
	case turns > 0:
		return Input{Rotate: true}
	case x < f.Pos.X:
		return Input{Shift: -1}
	case x > f.Pos.X:
		return Input{Shift: 1}
	}
	return Input{HardDrop: true}
}

// bestLanding is how many turns from now and which column the best landing
// needs, among those the fragment can reach shifting along its row.
func bestLanding(g *Game) (int, int, bool) {
	f := g.Fragment()
	best, bestTurns, bestX, found := 0.0, 0, 0, false
	s := f.Shape
	for turns := 0; turns < 4; turns, s = turns+1, s.Rotate() {
		if turns > 0 && sameCells(s, f.Shape) {
			break // a shape that looks the same turned needs no turns
		}
		for x := -s.Size; x <= Width; x++ {
			if !reachable(g, s, f.Pos, x) {
				continue
			}
			at := grid.Point{X: x, Y: f.Pos.Y}
			at.Y += g.disk.DropDistance(s, at)
			if v := evaluate(g.disk, s, at); !found || v > best {
				best, bestTurns, bestX, found = v, turns, x, true
			}
		}
	}
	return bestTurns, bestX, found
}

func sameCells(a, b polyomino.Shape) bool {
	key := func(p grid.Point) int { return p.Y*100 + p.X }
	ka, kb := make([]int, 0, 4), make([]int, 0, 4)
	for i := range a.Cells {
		ka, kb = append(ka, key(a.Cells[i])), append(kb, key(b.Cells[i]))
	}
	slices.Sort(ka)
	slices.Sort(kb)
	return slices.Equal(ka, kb)
}

// reachable reports whether s fits at every column from the fragment's
// to x along its row.
func reachable(g *Game, s polyomino.Shape, from grid.Point, x int) bool {
	step := 1
	if x < from.X {
		step = -1
	}
	for c := from.X; ; c += step {
		if !g.disk.Fits(s, grid.Point{X: c, Y: from.Y}) {
			return false
		}
		if c == x {
			return true
		}
	}
}

// evaluate scores a landing by the disk it leaves, with weights from a
// well-known hand-tuned heuristic: low landings, cleared rows, few holes,
// few transitions between filled and empty cells and shallow wells.
func evaluate(disk *polyomino.Well, s polyomino.Shape, at grid.Point) float64 {
	w := disk.Clone()
	w.Place(s, at, 1)
	landing := 0
	for _, c := range s.Cells {
		landing += Height - at.Y - c.Y
	}
	rows := len(w.FullRows())
	w.ClearRows(w.FullRows())
	filled := func(x, y int) bool {
		return x < 0 || x >= Width || y >= Height || (y >= 0 && w.At(grid.Point{X: x, Y: y}) != 0)
	}
	rowTrans, colTrans, holes, wells := 0, 0, 0, 0
	for y := 0; y < Height; y++ {
		for x := 0; x <= Width; x++ {
			if filled(x-1, y) != filled(x, y) {
				rowTrans++
			}
		}
	}
	for x := 0; x < Width; x++ {
		covered := false
		for y := 0; y <= Height; y++ {
			if filled(x, y-1) != filled(x, y) && y > 0 {
				colTrans++
			}
			if filled(x, y) && y < Height {
				covered = true
			} else if covered && y < Height {
				holes++
			}
		}
		depth := 0
		for y := 0; y < Height; y++ {
			if !filled(x, y) && filled(x-1, y) && filled(x+1, y) {
				depth++
				wells += depth
			} else if filled(x, y) {
				break
			} else {
				depth = 0
			}
		}
	}
	return -4.5*float64(landing)/4 + 3.4*float64(rows) - 3.2*float64(rowTrans) - 9.3*float64(colTrans) - 7.9*float64(holes) - 3.4*float64(wells)
}
