package hack

import (
	"math/rand"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

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
	b := newICE(rand.New(rand.NewSource(1)))
	b.walls = map[grid.Point]bool{}
	b.sentries = nil
	holdMove(b, right, 200)
	if b.Player() != b.Port() || b.Outcome() != Won {
		t.Fatalf("player %v port %v outcome %v", b.Player(), b.Port(), b.Outcome())
	}
}

func TestICEWallsBlockAndSentriesChaseAndCatch(t *testing.T) {
	b := newICE(rand.New(rand.NewSource(1)))
	b.walls = map[grid.Point]bool{pt(1, 4): true}
	b.sentries = nil
	b.Step(Input{Move: right})
	if b.Player() != pt(0, 4) {
		t.Fatalf("a wall must block, at %v", b.Player())
	}
	b.walls = map[grid.Point]bool{}
	b.sentries = []grid.Point{pt(5, 4)}
	for i := 0; i < SentryTicks; i++ {
		b.Step(Input{})
	}
	if b.Sentries()[0] != pt(4, 4) {
		t.Fatalf("a sentry must step toward the intruder, at %v", b.Sentries()[0])
	}
	for i := 0; i < 5*SentryTicks && b.Outcome() == Running; i++ {
		b.Step(Input{})
	}
	if b.Outcome() != Lost {
		t.Fatalf("a sentry reaching the intruder must end the breach lost")
	}
}

func TestICEStartsSolvableWithSentriesAwayFromTheIntruder(t *testing.T) {
	for seed := int64(1); seed <= 50; seed++ {
		b := newICE(rand.New(rand.NewSource(seed)))
		f := grid.Distances(ICEWidth, ICEHeight, b.Port(), func(p grid.Point) bool { return !b.walls[p] })
		if f.At(b.Player()) == grid.Unreachable {
			t.Fatalf("seed %d: the port cannot be reached", seed)
		}
		for _, s := range b.Sentries() {
			if d := abs(s.X-b.Player().X) + abs(s.Y-b.Player().Y); d < SentryDistance || b.walls[s] {
				t.Fatalf("seed %d: sentry at %v starts %d from the intruder", seed, s, d)
			}
		}
		if len(b.Sentries()) != Sentries || len(b.walls) == 0 {
			t.Fatalf("seed %d: %d sentries, %d walls", seed, len(b.Sentries()), len(b.walls))
		}
	}
}

func TestSyncHitsInsideTheWindowAndMissesOutside(t *testing.T) {
	b := newSync(rand.New(rand.NewSource(1)))
	b.cursor, b.window = 10, 9
	b.Step(Input{Act: true})
	if b.Hits() != 1 || b.Misses() != 0 {
		t.Fatalf("hits %d misses %d, want a hit inside the window", b.Hits(), b.Misses())
	}
	if b.Width() != SyncWindow-1 {
		t.Fatalf("window width %d, want it smaller after a hit", b.Width())
	}
	b.cursor, b.window = 0, 20
	b.Step(Input{Act: true})
	if b.Misses() != 1 {
		t.Fatalf("misses %d, want a miss outside the window", b.Misses())
	}
}

func TestSyncWinsOnThreeHitsAndLosesOnThreeMisses(t *testing.T) {
	b := newSync(rand.New(rand.NewSource(1)))
	for i := 0; i < SyncHits; i++ {
		b.cursor = b.window
		b.Step(Input{Act: true})
	}
	if b.Outcome() != Won {
		t.Fatalf("outcome %v after %d hits", b.Outcome(), SyncHits)
	}
	b = newSync(rand.New(rand.NewSource(1)))
	for i := 0; i < SyncMisses; i++ {
		b.cursor = (b.window + SyncBar/2) % SyncBar
		b.Step(Input{Act: true})
	}
	if b.Outcome() != Lost {
		t.Fatalf("outcome %v after %d misses", b.Outcome(), SyncMisses)
	}
}

func TestSyncCursorSweepsAndBounces(t *testing.T) {
	b := newSync(rand.New(rand.NewSource(1)))
	b.cursor, b.dir = SyncBar-2, 1
	for i := 0; i < 2*SyncTicks; i++ {
		b.Step(Input{})
	}
	if b.Cursor() != SyncBar-2 || b.dir != -1 {
		t.Fatalf("cursor %d dir %d, want bounced back from the end", b.Cursor(), b.dir)
	}
}

func TestGatesStartUnsolvedHaveASolutionAndWinWhenOutputsMatch(t *testing.T) {
	for seed := int64(1); seed <= 50; seed++ {
		b := newGates(rand.New(rand.NewSource(seed)))
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
	b := newGates(rand.New(rand.NewSource(1)))
	b.gates = [Switches - 1]Gate{GateAnd, GateAnd, GateAnd, GateAnd}
	b.target = [Switches - 1]bool{true, true, true, true} // only all on solves it
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
	for _, b := range []Breach{newICE(rng), newSync(rng), newGates(rng)} {
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
	for kind, want := range map[BreachKind]time.Duration{BreachICE: 20 * time.Second, BreachSync: 20 * time.Second, BreachGates: 30 * time.Second} {
		if got := (time.Duration(breachTicks[kind]) * TickPeriod).Round(time.Second); got != want {
			t.Fatalf("%v lasts %v, want %v", kind, got, want)
		}
	}
	if got := int(time.Second/TickPeriod) / SentryTicks; got != 4 {
		t.Fatalf("sentries %d cells/s, want 4", got)
	}
}

// solveGates finds switch settings whose outputs match the target.
func solveGates(b *gates) ([Switches]bool, bool) {
	for n := 0; n < 1<<Switches; n++ {
		var sw [Switches]bool
		for i := range sw {
			sw[i] = n&(1<<i) != 0
		}
		if b.outputsFor(sw) == b.target {
			return sw, true
		}
	}
	return [Switches]bool{}, false
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
	for i := ICEHeight; i < ICEWalls; i++ {
		first = append(first, 0, i%ICEHeight)
	}
	b := newICE(&scriptedRand{values: first})
	f := grid.Distances(ICEWidth, ICEHeight, b.Port(), func(p grid.Point) bool { return !b.walls[p] })
	if f.At(b.Player()) == grid.Unreachable {
		t.Fatalf("the breach kept walls that cut the port off")
	}
}

func TestSentriesNeverStepIntoWallsOrOffTheGrid(t *testing.T) {
	b := newICE(rand.New(rand.NewSource(1)))
	b.walls = map[grid.Point]bool{pt(19, 1): true}
	b.sentries = []grid.Point{pt(20, 0), pt(19, 0)}
	for i := 0; i < SentryTicks; i++ {
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
