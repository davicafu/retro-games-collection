package hack

import (
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// easy is level 1's difficulty, the one most tests use.
var easy = difficulties[0]

func pt(x, y int) grid.Point { return grid.Point{X: x, Y: y} }

var (
	left  = grid.Point{X: -1}
	right = grid.Point{X: 1}
	up    = grid.Point{Y: -1}
	down  = grid.Point{Y: 1}
)

func holdMove(b Breach, d grid.Point, n int) {
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			b.Step(Input{Move: d})
		} else {
			b.Step(Input{})
		}
	}
}

func TestICEReachingThePortWins(t *testing.T) {
	b := newICE(rand.New(rand.NewSource(1)), easy)
	b.walls = map[grid.Point]bool{}
	b.sentries = nil
	holdMove(b, right, 200)
	if b.Player() != b.Port() || b.Outcome() != Won {
		t.Fatalf("player %v port %v outcome %v", b.Player(), b.Port(), b.Outcome())
	}
}

func TestICEWallsBlockAndSentriesChaseAndCatch(t *testing.T) {
	b := newICE(rand.New(rand.NewSource(1)), easy)
	b.walls = map[grid.Point]bool{pt(1, 4): true}
	b.sentries = nil
	b.Step(Input{Move: right})
	if b.Player() != pt(0, 4) {
		t.Fatalf("a wall must block, at %v", b.Player())
	}
	b.walls = map[grid.Point]bool{}
	b.sentries = []grid.Point{pt(5, 4)}
	for i := 0; i < easy.SentryTicks; i++ {
		b.Step(Input{})
	}
	if b.Sentries()[0] != pt(4, 4) {
		t.Fatalf("a sentry must step toward the intruder, at %v", b.Sentries()[0])
	}
	for i := 0; i < 5*easy.SentryTicks && b.Outcome() == Running; i++ {
		b.Step(Input{})
	}
	if b.Outcome() != Lost {
		t.Fatalf("a sentry reaching the intruder must end the breach lost")
	}
}

func TestICEStartsSolvableWithSentriesAwayFromTheIntruder(t *testing.T) {
	for seed := int64(1); seed <= 50; seed++ {
		b := newICE(rand.New(rand.NewSource(seed)), easy)
		f := grid.Distances(ICEWidth, ICEHeight, b.Port(), func(p grid.Point) bool { return !b.walls[p] })
		if f.At(b.Player()) == grid.Unreachable {
			t.Fatalf("seed %d: the port cannot be reached", seed)
		}
		for _, s := range b.Sentries() {
			if d := abs(s.X-b.Player().X) + abs(s.Y-b.Player().Y); d < SentryDistance || b.walls[s] {
				t.Fatalf("seed %d: sentry at %v starts %d from the intruder", seed, s, d)
			}
		}
		if len(b.Sentries()) != easy.Sentries || len(b.walls) == 0 {
			t.Fatalf("seed %d: %d sentries, %d walls", seed, len(b.Sentries()), len(b.walls))
		}
	}
}

func TestSyncHitsInsideTheWindowAndMissesOutside(t *testing.T) {
	b := newSync(rand.New(rand.NewSource(1)), easy)
	b.cursor, b.window = 10, 9
	b.Step(Input{Act: true})
	if b.Hits() != 1 || b.Misses() != 0 {
		t.Fatalf("hits %d misses %d, want a hit inside the window", b.Hits(), b.Misses())
	}
	if b.Width() != easy.SyncWindow-1 {
		t.Fatalf("window width %d, want it smaller after a hit", b.Width())
	}
	b.cursor, b.window = 0, 20
	b.Step(Input{Act: true})
	if b.Misses() != 1 {
		t.Fatalf("misses %d, want a miss outside the window", b.Misses())
	}
}

func TestSyncWinsOnThreeHitsAndLosesOnThreeMisses(t *testing.T) {
	b := newSync(rand.New(rand.NewSource(1)), easy)
	for i := 0; i < easy.SyncHits; i++ {
		b.cursor = b.window
		b.Step(Input{Act: true})
	}
	if b.Outcome() != Won {
		t.Fatalf("outcome %v after %d hits", b.Outcome(), easy.SyncHits)
	}
	b = newSync(rand.New(rand.NewSource(1)), easy)
	for i := 0; i < SyncMisses; i++ {
		b.cursor = (b.window + SyncBar/2) % SyncBar
		b.Step(Input{Act: true})
	}
	if b.Outcome() != Lost {
		t.Fatalf("outcome %v after %d misses", b.Outcome(), SyncMisses)
	}
}

func TestSyncCursorSweepsAndBounces(t *testing.T) {
	b := newSync(rand.New(rand.NewSource(1)), easy)
	b.cursor, b.dir = SyncBar-2, 1
	for i := 0; i < 2*easy.SyncTicks; i++ {
		b.Step(Input{})
	}
	if b.Cursor() != SyncBar-2 || b.dir != -1 {
		t.Fatalf("cursor %d dir %d, want bounced back from the end", b.Cursor(), b.dir)
	}
}

func TestGatesStartUnsolvedHaveASolutionAndWinWhenOutputsMatch(t *testing.T) {
	for seed := int64(1); seed <= 50; seed++ {
		b := newGates(rand.New(rand.NewSource(seed)), easy)
		if b.Solved() {
			t.Fatalf("seed %d starts solved", seed)
		}
		sol, ok := solveGates(b)
		if !ok {
			t.Fatalf("seed %d has no solution", seed)
		}
		b.switches = sol
		b.Step(Input{})
		if b.Outcome() != Won {
			t.Fatalf("seed %d: matching outputs must win", seed)
		}
	}
}

func TestGatesArrowsSelectAndSetSwitches(t *testing.T) {
	b := newGates(rand.New(rand.NewSource(1)), easy)
	b.gates = []Gate{GateAnd, GateAnd, GateAnd, GateAnd}
	b.target = []bool{true, true, true, true} // only all on solves it
	b.Step(Input{Move: right})
	b.Step(Input{Move: up})
	if b.Cursor() != 1 || !b.Switches()[1] {
		t.Fatalf("cursor %d switches %v, want switch 1 on", b.Cursor(), b.Switches())
	}
	b.Step(Input{Move: down})
	if b.Switches()[1] {
		t.Fatalf("down must turn the switch off")
	}
	b.Step(Input{Act: true})
	if !b.Switches()[1] {
		t.Fatalf("space must flip the switch")
	}
	b.Step(Input{Move: left})
	b.Step(Input{Move: left})
	if b.Cursor() != 0 {
		t.Fatalf("the cursor must stop at the first switch, at %d", b.Cursor())
	}
}

func TestGateTruthTables(t *testing.T) {
	cases := []struct {
		g    Gate
		want [4]bool // 00 01 10 11
	}{
		{GateAnd, [4]bool{false, false, false, true}},
		{GateOr, [4]bool{false, true, true, true}},
		{GateXor, [4]bool{false, true, true, false}},
		{GateNand, [4]bool{true, true, true, false}},
	}
	for _, c := range cases {
		for i, want := range c.want {
			if got := c.g.Eval(i&2 != 0, i&1 != 0); got != want {
				t.Fatalf("%v(%v,%v) = %v, want %v", c.g, i&2 != 0, i&1 != 0, got, want)
			}
		}
	}
}

func TestEveryBreachRunsOutOfTime(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for _, b := range []Breach{newICE(rng, easy), newSync(rng, easy), newGates(rng, easy)} {
		if ice, ok := b.(*ice); ok {
			ice.sentries = nil
		}
		left := b.TicksLeft()
		for i := 0; i < left; i++ {
			b.Step(Input{})
		}
		if b.Outcome() != Lost {
			t.Fatalf("%v: outcome %v after its %d ticks", b.Kind(), b.Outcome(), left)
		}
	}
}

func TestBreachTimeLimitsInSeconds(t *testing.T) {
	want := []struct{ ice, sync, gates time.Duration }{{20 * time.Second, 20 * time.Second, 30 * time.Second}, {20 * time.Second, 20 * time.Second, 25 * time.Second}, {18 * time.Second, 20 * time.Second, 25 * time.Second}}
	secs := func(ticks int) time.Duration { return (time.Duration(ticks) * TickPeriod).Round(time.Second) }
	for lvl, d := range difficulties[:3] {
		if secs(d.ICETicks) != want[lvl].ice || secs(d.SyncTimeTicks) != want[lvl].sync || secs(d.GatesTicks) != want[lvl].gates {
			t.Fatalf("level %d limits %v %v %v, want %+v", lvl+1, secs(d.ICETicks), secs(d.SyncTimeTicks), secs(d.GatesTicks), want[lvl])
		}
		// Sentries move 4, 5 and about 5.5 cells/s on the first three levels.
		if got := float64(time.Second/TickPeriod) / float64(d.SentryTicks); got < []float64{4, 5, 5.4}[lvl] || got > []float64{4, 5, 5.5}[lvl] {
			t.Fatalf("level %d sentries %.2f cells/s", lvl+1, got)
		}
	}
}

func TestEveryLevelMakesTheBreachesHarder(t *testing.T) {
	for i := 1; i < len(difficulties); i++ {
		a, b := difficulties[i-1], difficulties[i]
		// Never easier in anything...
		if b.ICEWalls < a.ICEWalls || b.Sentries < a.Sentries || b.SentryTicks > a.SentryTicks || b.ICETicks > a.ICETicks ||
			b.SyncWindow > a.SyncWindow || b.SyncHits < a.SyncHits || b.SyncTicks > a.SyncTicks || b.SyncTimeTicks > a.SyncTimeTicks ||
			b.Switches < a.Switches || b.GatesTicks > a.GatesTicks || b.BreachTrace < a.BreachTrace || b.WatchTicks > a.WatchTicks {
			t.Fatalf("level %d is easier than level %d somewhere: %+v vs %+v", i+1, i, b, a)
		}
		// ...and harder in every kind of breach.
		ice := b.ICEWalls > a.ICEWalls || b.Sentries > a.Sentries || b.SentryTicks < a.SentryTicks || b.ICETicks < a.ICETicks
		sync := b.SyncWindow < a.SyncWindow || b.SyncHits > a.SyncHits || b.SyncTicks < a.SyncTicks || b.SyncTimeTicks < a.SyncTimeTicks
		gates := b.Switches > a.Switches || b.GatesTicks < a.GatesTicks
		if !ice || !sync || !gates {
			t.Fatalf("level %d must be harder than level %d in every breach: ice %v sync %v gates %v", i+1, i, ice, sync, gates)
		}
	}
}

func TestBreachesFollowTheirLevelsDifficulty(t *testing.T) {
	for lvl, d := range difficulties {
		rng := rand.New(rand.NewSource(int64(lvl + 1)))
		ice := newICE(rng, d)
		if len(ice.Sentries()) != d.Sentries || len(ice.walls) != d.ICEWalls || ice.TicksLeft() != d.ICETicks {
			t.Fatalf("level %d ICE: %d sentries %d walls %d ticks", lvl+1, len(ice.Sentries()), len(ice.walls), ice.TicksLeft())
		}
		ice.walls, ice.sentries = map[grid.Point]bool{}, []grid.Point{pt(10, 4)}
		for i := 0; i < d.SentryTicks; i++ {
			ice.Step(Input{})
		}
		if ice.Sentries()[0] != pt(9, 4) {
			t.Fatalf("level %d: a sentry must step every %d ticks", lvl+1, d.SentryTicks)
		}
		s := newSync(rng, d)
		if s.Width() != d.SyncWindow || s.TicksLeft() != d.SyncTimeTicks {
			t.Fatalf("level %d sync: window %d ticks %d", lvl+1, s.Width(), s.TicksLeft())
		}
		for i := 0; i < d.SyncHits-1; i++ {
			s.cursor = s.window
			s.Step(Input{Act: true})
		}
		if s.Outcome() != Running || s.Width() < SyncMinWindow {
			t.Fatalf("level %d sync: outcome %v width %d after %d hits", lvl+1, s.Outcome(), s.Width(), d.SyncHits-1)
		}
		s.cursor = s.window
		s.Step(Input{Act: true})
		if s.Outcome() != Won {
			t.Fatalf("level %d sync: %d hits must win", lvl+1, d.SyncHits)
		}
		g := newGates(rng, d)
		if len(g.Switches()) != d.Switches || len(g.Gates()) != d.Switches-1 || g.TicksLeft() != d.GatesTicks {
			t.Fatalf("level %d gates: %d switches %d gates", lvl+1, len(g.Switches()), len(g.Gates()))
		}
		if _, ok := solveGates(g); !ok || g.Solved() {
			t.Fatalf("level %d gates must start unsolved with a solution", lvl+1)
		}
	}
}

// solveGates finds switch settings whose outputs match the target.
func solveGates(b *gates) ([]bool, bool) {
	n := len(b.switches)
	for m := 0; m < 1<<n; m++ {
		sw := make([]bool, n)
		for i := range sw {
			sw[i] = m&(1<<i) != 0
		}
		if slices.Equal(b.outputsFor(sw), b.target) {
			return sw, true
		}
	}
	return nil, false
}

// scriptedRand returns its values in turn, then counts up.
type scriptedRand struct {
	values []int
	n      int
}

func (r *scriptedRand) Intn(n int) int {
	if len(r.values) > 0 {
		v := r.values[0]
		r.values = r.values[1:]
		return v % n
	}
	r.n++
	return r.n % n
}

func TestICEDealsAgainWhenTheWallsCutThePortOff(t *testing.T) {
	// The first deal walls column 10 from top to bottom.
	var first []int
	for y := 0; y < ICEHeight; y++ {
		first = append(first, 8, y)
	}
	for i := ICEHeight; i < easy.ICEWalls; i++ {
		first = append(first, 0, i%ICEHeight)
	}
	b := newICE(&scriptedRand{values: first}, easy)
	f := grid.Distances(ICEWidth, ICEHeight, b.Port(), func(p grid.Point) bool { return !b.walls[p] })
	if f.At(b.Player()) == grid.Unreachable {
		t.Fatalf("the breach kept walls that cut the port off")
	}
}

func TestSentriesNeverStepIntoWallsOrOffTheGrid(t *testing.T) {
	b := newICE(rand.New(rand.NewSource(1)), easy)
	b.walls = map[grid.Point]bool{pt(19, 1): true}
	b.sentries = []grid.Point{pt(20, 0), pt(19, 0)}
	for i := 0; i < easy.SentryTicks; i++ {
		b.Step(Input{})
	}
	for _, s := range b.Sentries() {
		if !b.open(s) {
			t.Fatalf("a sentry stepped to %v, off the grid or into a wall", s)
		}
	}
	if b.Sentries()[1] != pt(18, 0) {
		t.Fatalf("the sentry must step toward the intruder, at %v", b.Sentries()[1])
	}
}
