package hack

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// Viability (S6): an autopilot that types the solution at TypeTicks a
// character and plays each breach at a human pace wins at least
// ViableSeeds of AutopilotSeeds games.
const (
	AutopilotSeeds = 30
	ViableSeeds    = 27
	TypeTicks      = 6 // 10 characters a second
	MoveTicks      = 4 // a held key's repeats while breaching
)

func TestAutopilotHacksTheCore(t *testing.T) {
	won := make([]bool, AutopilotSeeds)
	t.Run("games", func(t *testing.T) {
		for i := range won {
			t.Run(fmt.Sprintf("seed %d", i+1), func(t *testing.T) {
				t.Parallel()
				g := NewGame(Config{Rand: rand.New(rand.NewSource(int64(i + 1)))})
				a := &autopilot{}
				for tick := 0; tick < 60*60*10 && !g.Over(); tick++ {
					g.Step(a.act(g, tick))
				}
				won[i] = g.Won()
				if !won[i] {
					t.Logf("trace %d, node %s, step %d", g.Trace(), g.Node(), a.step)
				}
			})
		}
	})
	n := 0
	for _, w := range won {
		if w {
			n++
		}
	}
	if n < ViableSeeds {
		t.Fatalf("autopilot won %d of %d games, want at least %d", n, AutopilotSeeds, ViableSeeds)
	}
	t.Logf("autopilot won %d of %d games", n, AutopilotSeeds)
}

// step is a line of the solution, done once its effect shows.
type step struct {
	line string
	done func(g *Game) bool
}

func owns(id string) func(*Game) bool { return func(g *Game) bool { return g.Owned(id) } }

var solution = []step{
	{line: "cat readme.txt"},
	{line: "scan"},
	{line: "connect mail", done: owns("mail")},
	{line: "cat inbox_0412.txt"},
	{line: "cat newsletter.txt"},
	{line: "connect dev", done: owns("dev")},
	{line: "cat todo.txt"},
	{line: "connect vault", done: func(g *Game) bool { return strings.HasPrefix(g.Prompt(), "password") }},
	{line: "pixel", done: owns("vault")},
	{line: "decrypt core.key.enc ember"},
	{line: "connect core", done: owns("core")},
	{line: "cat lantern.txt"},
}

type autopilot struct {
	step      int
	typed     int
	submitted bool
}

func (a *autopilot) act(g *Game, tick int) Input {
	if g.Mode() == ModeBreach {
		return playBreach(g.Breach(), tick)
	}
	if tick%TypeTicks != 0 || a.step >= len(solution) {
		return Input{}
	}
	s := solution[a.step]
	if a.submitted {
		a.submitted = false
		if s.done == nil || s.done(g) {
			a.step++
			return Input{}
		}
		// A lost breach or a refused password: type the line again.
	}
	if a.typed < len(s.line) {
		a.typed++
		return Input{Text: []rune{rune(s.line[a.typed-1])}}
	}
	a.typed, a.submitted = 0, true
	return Input{Enter: true}
}

func playBreach(b Breach, tick int) Input {
	switch b := b.(type) {
	case *ice:
		if tick%MoveTicks == 0 {
			return Input{Move: iceMove(b)}
		}
	case *syncBreach:
		// Press once the cursor is well inside the window.
		if tick%2 == 0 && b.Cursor() > b.Window() && b.Cursor() < b.Window()+b.Width()-1 ||
			b.Width() <= 2 && b.Cursor() >= b.Window() && b.Cursor() < b.Window()+b.Width() && tick%2 == 0 {
			return Input{Act: true}
		}
	case *gates:
		if tick%TypeTicks == 0 {
			return gatesMove(b)
		}
	}
	return Input{}
}

// iceMove heads for the port along a shortest path that keeps a cell away
// from every sentry, or waits when there is none.
func iceMove(b *ice) grid.Point {
	safe := func(p grid.Point) bool {
		if !b.open(p) {
			return false
		}
		for _, s := range b.Sentries() {
			if p != b.Port() && manhattan(p, s) <= 1 {
				return false
			}
		}
		return true
	}
	f := grid.Distances(ICEWidth, ICEHeight, b.Port(), safe)
	best, bestD := grid.Point{}, f.At(b.Player())
	for _, d := range grid.Neighbours {
		if n := b.Player().Add(d); f.At(n) != grid.Unreachable && (bestD == grid.Unreachable || f.At(n) < bestD) {
			best, bestD = d, f.At(n)
		}
	}
	return best
}

// gatesMove walks the cursor to the first switch that differs from a
// solution and sets it.
func gatesMove(b *gates) Input {
	sol, ok := solveGates(b)
	if !ok {
		return Input{}
	}
	sw := b.Switches()
	for i := range sw {
		if sw[i] == sol[i] {
			continue
		}
		switch {
		case i < b.Cursor():
			return Input{Move: left}
		case i > b.Cursor():
			return Input{Move: right}
		case sol[i]:
			return Input{Move: up}
		}
		return Input{Move: down}
	}
	return Input{}
}
