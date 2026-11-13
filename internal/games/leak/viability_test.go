package leak

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// Viability (S8): over a fixed set of seeds, a simple autopilot must clear
// the first ViableLevels levels with its StartLives lives in at least
// ViableSeeds of AutopilotSeeds games.
const (
	ViableLevels   = 3
	AutopilotSeeds = 30
	ViableSeeds    = 27
)

func TestAutopilotClearsTheFirstLevels(t *testing.T) {
	// Games run in parallel; each one writes only its own slot.
	won := make([]bool, AutopilotSeeds)
	t.Run("games", func(t *testing.T) {
		for i := range won {
			t.Run(fmt.Sprintf("seed %d", i+1), func(t *testing.T) {
				t.Parallel()
				g := NewGame(Config{Maps: Maps, Rand: rand.New(rand.NewSource(int64(i + 1)))})
				for tick := 0; tick < 60*60*10 && !g.Over() && g.Level() <= ViableLevels; tick++ {
					g.Step(autopilot(g))
				}
				won[i] = g.Level() > ViableLevels
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
		t.Fatalf("autopilot cleared %d levels in %d of %d games, want at least %d", ViableLevels, cleared, AutopilotSeeds, ViableSeeds)
	}
	t.Logf("autopilot cleared %d levels in %d of %d games", ViableLevels, cleared, AutopilotSeeds)
}

// autopilot picks the leak's direction for this tick. It searches from both
// cells the leak is between at once (the one behind costs a reverse) and
// heads for the nearest byte or malloc it reaches with a margin before any
// collector; with none, it runs to the cell farthest ahead of the
// collectors.
func autopilot(g *Game) Input {
	const margin = 2
	leak := g.Leak()
	threat := threatField(g)

	type node struct {
		dist  int
		input grid.Point // the input to give now to follow this path
		// turn marks the cell where the leak takes its next decision: a
		// path leaving it is followed by pressing that path's direction.
		turn bool
	}
	// seen is indexed by cell; a zero dist+1 marks a cell not reached yet.
	seen := make([]node, g.Width()*g.Height())
	reached := func(p grid.Point) bool { return seen[p.Y*g.Width()+p.X].dist > 0 }
	at := func(p grid.Point) *node { return &seen[p.Y*g.Width()+p.X] }
	queue := make([]grid.Point, 0, len(seen))
	visit := func(p grid.Point, n node) {
		n.dist++ // stored one higher so that the zero value means unreached
		*at(p) = n
		queue = append(queue, p)
	}
	seed := func(p grid.Point, n node) {
		if threat(p) > n.dist {
			visit(p, n)
		}
	}
	if leak.Progress == 0 {
		seed(leak.Cell, node{dist: 0, turn: true})
	} else {
		ahead, behind := 1, 0
		if leak.Progress >= CellUnits/2 {
			ahead, behind = 0, 1
		}
		seed(leak.Cell.Add(leak.Dir), node{dist: ahead, input: leak.Dir, turn: true})
		seed(leak.Cell, node{dist: behind, input: neg(leak.Dir)})
	}

	food, flee := grid.Point{}, grid.Point{}
	foodDist, fleeGap, fleeDist := -1, -1, -1
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		n := *at(p)
		n.dist--
		// Flee along a move (never stand still), as far ahead of the
		// collectors as possible.
		if gap := threat(p) - n.dist; n.input != (grid.Point{}) && (gap > fleeGap || gap == fleeGap && n.dist > fleeDist) {
			flee, fleeGap, fleeDist = n.input, gap, n.dist
		}
		if t := g.Tile(p); foodDist < 0 && (t == TileByte || t == TileMalloc) && threat(p) > n.dist+margin {
			food, foodDist = n.input, n.dist
		}
		for _, d := range grid.Neighbours {
			q := p.Add(d)
			if !g.leakCanEnter(q) || reached(q) || threat(q) <= n.dist+1 {
				continue
			}
			input := n.input
			if n.turn {
				input = d
			}
			visit(q, node{dist: n.dist + 1, input: input})
		}
	}
	if foodDist >= 0 {
		return Input{Dir: food}
	}
	return Input{Dir: flee}
}

// threatField returns, for a cell, how soon any chasing collector can reach
// it, in the leak's steps: a collector's steps scaled by its slower speed.
func threatField(g *Game) func(grid.Point) int {
	var fields []grid.DistanceField
	for _, c := range g.Collectors() {
		if c.Fleeing() || c.state == stateWaiting {
			continue
		}
		fields = append(fields, grid.Distances(g.Width(), g.Height(), c.Occupied(), g.collectorCanEnter(stateLeaving)))
	}
	return func(p grid.Point) int {
		least := 1 << 30
		for _, f := range fields {
			if d := f.At(p); d != grid.Unreachable && d < least {
				least = d
			}
		}
		return least * LeakSpeed / g.collectorSpeed()
	}
}
