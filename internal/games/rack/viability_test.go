package rack

import (
	"fmt"
	"maps"
	"math/rand"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// Viability (S8): an autopilot that presses a key every ActTicks, like a
// held key's repeats, clears the first ViableLevels levels with its
// StartLives lives in at least ViableSeeds of AutopilotSeeds games.
const (
	ViableLevels   = 3
	AutopilotSeeds = 30
	ViableSeeds    = 27
	ActTicks       = 2
)

func TestAutopilotClearsTheFirstLevels(t *testing.T) {
	won := make([]bool, AutopilotSeeds)
	t.Run("games", func(t *testing.T) {
		for i := range won {
			t.Run(fmt.Sprintf("seed %d", i+1), func(t *testing.T) {
				t.Parallel()
				g := NewGame(Config{Levels: Levels, Rand: rand.New(rand.NewSource(int64(i + 1)))})
				var a autopilot
				for tick := 0; tick < 60*60*10 && !g.Over() && g.Level() <= ViableLevels; tick++ {
					var in Input
					if tick%ActTicks == 0 {
						in = a.act(g)
					}
					g.Step(in)
				}
				won[i] = g.Level() > ViableLevels
				if !won[i] {
					t.Logf("level %d, lives %d, disks left %d", g.Level(), g.Lives(), g.DisksLeft())
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
		t.Fatalf("autopilot cleared %d levels in %d of %d games, want at least %d", ViableLevels, cleared, AutopilotSeeds, ViableSeeds)
	}
	t.Logf("autopilot cleared %d levels in %d of %d games", ViableLevels, cleared, AutopilotSeeds)
}

// Autopilot tuning: how many ticks ahead it tries each move.
const lookAhead = 24

type autopilot struct{}

// act tries, in order, the next step of the shortest path to the nearest
// disk (or the door once they are collected), that step with a jump, a
// jump in place, waiting, and every other step with and without a jump. It
// takes the first one it survives for lookAhead ticks, whether cables go
// down every ladder or none.
func (a *autopilot) act(g *Game) Input {
	var tries []Input
	if step, ok := nextStep(g); ok {
		d := grid.Point{X: step.X - g.Player().Pos.X, Y: step.Y - g.Player().Pos.Y}
		tries = append(tries, Input{Move: d}, Input{Move: d, Jump: true})
	}
	tries = append(tries, Input{Jump: true}, Input{})
	for _, d := range []grid.Point{{X: -1}, {X: 1}, {Y: -1}, {Y: 1}} {
		tries = append(tries, Input{Move: d}, Input{Move: d, Jump: true})
	}
	best, longest := Input{}, -1
	for _, in := range tries {
		lasts := min(survives(g, in, alwaysDescend{}), survives(g, in, neverDescend{}))
		if lasts >= lookAhead {
			return in
		}
		if lasts > longest {
			best, longest = in, lasts
		}
	}
	return best
}

// survives plays in on a copy of g, pressing its move every ActTicks and
// jumping on the first tick only, and returns how many ticks pass before
// the technician is caught, up to lookAhead. Reaching the door counts as
// surviving.
func survives(g *Game, in Input, rng Rand) int {
	c := g.clone(rng)
	lives, level := c.Lives(), c.Level()
	for t := 0; t < lookAhead; t++ {
		var press Input
		if t%ActTicks == 0 {
			press = Input{Move: in.Move, Jump: in.Jump && t == 0}
		}
		c.Step(press)
		if c.Level() != level {
			return lookAhead
		}
		if c.Lives() != lives || c.Over() {
			return t
		}
	}
	return lookAhead
}

// clone copies g with its own cables, bots, disks and randomness.
func (g *Game) clone(rng Rand) *Game {
	c := *g
	c.rng = rng
	c.cables = append([]Cable(nil), g.cables...)
	c.bots = append([]Bot(nil), g.bots...)
	c.disks = maps.Clone(g.disks)
	c.tools = maps.Clone(g.tools)
	c.thrown = append([]Wrench(nil), g.thrown...)
	return &c
}

type alwaysDescend struct{}

func (alwaysDescend) Intn(int) int { return 0 }

type neverDescend struct{}

func (neverDescend) Intn(n int) int { return n - 1 }

// nextStep is the first cell of the shortest path to the nearest disk, or
// to the door once the disks are collected.
func nextStep(g *Game) (grid.Point, bool) {
	start := g.Player().Pos
	from := map[grid.Point]grid.Point{start: start}
	queue := []grid.Point{start}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		if c := g.Cell(p); p != start && (c == CellDisk || g.DoorOpen() && p == g.Exit()) {
			for from[p] != start {
				p = from[p]
			}
			return p, true
		}
		for _, n := range g.moves(p) {
			if _, seen := from[n]; !seen {
				from[n] = p
				queue = append(queue, n)
			}
		}
	}
	return grid.Point{}, false
}
