package hack

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// Viability (S6, S16, S18): an autopilot that types each level's solution
// at TypeTicks a character and plays each breach at a human pace clears
// levels 1 to 3 in a row in at least ViableSeeds of AutopilotSeeds games,
// and, starting fresh on them, clears level 4 in about 50% and level 5 in
// about 40% of RateSeeds games.
const (
	AutopilotSeeds = 30
	ViableSeeds    = 27
	RateSeeds      = 200
	TypeTicks      = 6 // 10 characters a second
	MoveTicks      = 4 // a held key's repeats while breaching
	maxTicks       = 60 * 60 * 30
)

// targetRates are the share of games the autopilot clears levels 4 and 5
// in, give or take rateTolerance.
var targetRates = map[int]float64{3: 0.50, 4: 0.40}

const rateTolerance = 0.05

func TestAutopilotClearsTheFirstThreeLevels(t *testing.T) {
	cleared := make([]bool, AutopilotSeeds)
	t.Run("games", func(t *testing.T) {
		for i := range cleared {
			t.Run(fmt.Sprintf("seed %d", i+1), func(t *testing.T) {
				t.Parallel()
				g := NewGame(Config{Rand: rand.New(rand.NewSource(int64(i + 1)))})
				a := &autopilot{}
				for tick := 0; tick < maxTicks && !g.Over() && g.Level() < 3; tick++ {
					g.Step(a.act(g, tick))
				}
				cleared[i] = g.Level() == 3
				if !cleared[i] {
					t.Logf("level %d trace %d, node %s, step %d", g.Level()+1, g.Trace(), g.Node(), a.step)
				}
			})
		}
	})
	n := 0
	for _, c := range cleared {
		if c {
			n++
		}
	}
	if n < ViableSeeds {
		t.Fatalf("autopilot cleared levels 1 to 3 in %d of %d games, want at least %d", n, AutopilotSeeds, ViableSeeds)
	}
	t.Logf("autopilot cleared levels 1 to 3 in %d of %d games", n, AutopilotSeeds)
}

// clearRate is the share of RateSeeds games, each starting fresh on level
// lvl, that the autopilot clears it in.
func clearRate(lvl int) float64 {
	results := make(chan bool, RateSeeds)
	for i := 0; i < RateSeeds; i++ {
		go func(seed int64) {
			g := NewGame(Config{Rand: rand.New(rand.NewSource(seed))})
			g.loadLevel(lvl)
			a := &autopilot{level: lvl}
			for tick := 0; tick < maxTicks && !g.Over() && !g.cleared; tick++ {
				g.Step(a.act(g, tick))
			}
			results <- g.cleared || g.Won()
		}(int64(i + 1))
	}
	n := 0
	for i := 0; i < RateSeeds; i++ {
		if <-results {
			n++
		}
	}
	return float64(n) / RateSeeds
}

func TestTheLastTwoLevelsHitTheirTargetRates(t *testing.T) {
	for lvl := 3; lvl < len(levels); lvl++ {
		rate, want := clearRate(lvl), targetRates[lvl]
		t.Logf("level %d cleared in %.0f%% of %d games", lvl+1, 100*rate, RateSeeds)
		if rate < want-rateTolerance || rate > want+rateTolerance {
			t.Fatalf("level %d cleared in %.0f%% of games, want %.0f%% ± %.0f", lvl+1, 100*rate, 100*want, 100*rateTolerance)
		}
	}
}

// step is a line of a solution, done once its effect shows.
type step struct {
	line string
	done func(g *Game) bool
}

func asksPassword(g *Game) bool { return strings.HasPrefix(g.Prompt(), "contraseña") }

func standsOn(id string) func(*Game) bool { return func(g *Game) bool { return g.Node() == id } }

// solutions are every level's walk-through, reading the files a player
// needs.
var solutions = [][]step{
	{
		{line: "cat readme.txt"},
		{line: "scan"},
		{line: "connect mail", done: owns("mail")},
		{line: "cat correo_0412.txt"},
		{line: "cat boletin.txt"},
		{line: "connect dev", done: owns("dev")},
		{line: "cat todo.txt"},
		{line: "connect vault", done: asksPassword},
		{line: "pixel", done: owns("vault")},
		{line: "decrypt core.key.enc ember"},
		{line: "connect core", done: owns("core")},
		{line: "cat lantern.txt"},
	},
	{
		{line: "cat readme.txt"},
		{line: "connect rrhh", done: owns("rrhh")},
		{line: "cat politica.txt"},
		{line: "cat ficha_marta.txt"},
		{line: "connect fw", done: owns("fw")},
		{line: "connect lab", done: asksPassword},
		{line: "nimbus1987", done: owns("lab")},
		{line: "cat correo_marta.txt"},
		{line: "connect datos", done: owns("datos")},
		{line: "decrypt hx12_real.enc 21xh"},
	},
	{
		{line: "cat readme.txt"},
		{line: "connect cam", done: owns("cam")},
		{line: "cat mantenimiento.txt"},
		{line: "connect nas", done: owns("nas")},
		{line: "cat fragmento_1.txt"},
		{line: "connect backup", done: asksPassword},
		{line: "lertsek", done: owns("backup")},
		{line: "cat fragmento_2.txt"},
		{line: "connect nas", done: standsOn("nas")},
		{line: "connect fw", done: owns("fw")},
		{line: "decrypt nucleo.key.enc orion42"},
		{line: "connect core", done: owns("core")},
		{line: "cat kestrel.txt"},
	},
	{
		{line: "cat readme.txt"},
		{line: "connect puerto", done: owns("puerto")},
		{line: "cat grua12.log"},
		{line: "connect ids", done: owns("ids")},
		{line: "connect correo", done: owns("correo")},
		{line: "cat correo_jefe.txt"},
		{line: "connect aduanas", done: asksPassword},
		{line: "albatros7", done: owns("aduanas")},
		{line: "cat manifiesto.txt"},
		{line: "connect flota", done: owns("flota")},
		{line: "decrypt manifiesto_real.enc 4471"},
	},
	{
		{line: "cat readme.txt"},
		{line: "cat notas.txt"},
		{line: "connect legal", done: owns("legal")},
		{line: "cat circular.txt"},
		{line: "connect auditoria", done: owns("auditoria")},
		{line: "connect archivo", done: owns("archivo")},
		{line: "cat candidatas.txt"},
		{line: "connect prensa", done: owns("prensa")},
		{line: "cat entrevista.txt"},
		{line: "connect archivo", done: standsOn("archivo")},
		{line: "connect presidencia", done: asksPassword},
		{line: "aurora1959", done: owns("presidencia")},
		{line: "cat nota_presidenta.txt"},
		{line: "connect core", done: owns("core")},
		{line: "decrypt halcyon.enc lhk"},
	},
}

type autopilot struct {
	level     int
	step      int
	typed     int
	submitted bool
}

func (a *autopilot) act(g *Game, tick int) Input {
	if g.Mode() == ModeBreach {
		return playBreach(g.Breach(), tick)
	}
	if tick%TypeTicks != 0 {
		return Input{}
	}
	if g.cleared {
		// Enter opens the next level.
		a.level, a.step, a.typed, a.submitted = g.Level()+1, 0, 0, false
		return Input{Enter: true}
	}
	sol := solutions[a.level]
	if a.step >= len(sol) {
		return Input{}
	}
	s := sol[a.step]
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
