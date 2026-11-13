package leak

import (
	"fmt"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// scriptedRand returns its values in turn (0 when empty) and counts calls.
type scriptedRand struct {
	values []int
	calls  int
}

func (r *scriptedRand) Intn(n int) int {
	r.calls++
	if len(r.values) == 0 {
		return 0
	}
	v := r.values[0]
	r.values = r.values[1:]
	return v % n
}

func newTestGame(rows ...string) *Game {
	return NewGame(Config{Maps: [][]string{rows}, Rand: &scriptedRand{}})
}

func pt(x, y int) grid.Point { return grid.Point{X: x, Y: y} }

// addCollector puts an active collector in the corridors.
func addCollector(g *Game, kind Kind, cell, dir grid.Point) *Collector {
	g.collectors = append(g.collectors, Collector{Kind: kind, Actor: Actor{Cell: cell, Dir: dir}, home: cell, state: stateActive})
	return &g.collectors[len(g.collectors)-1]
}

func steps(g *Game, n int) {
	for i := 0; i < n; i++ {
		g.Step(Input{})
	}
}

// along is an actor's position along x in units, the way it is drawn.
func along(a Actor) int { return a.Cell.X*CellUnits + a.Dir.X*a.Progress }

var corridor = []string{
	"#######",
	"#L....#",
	"###.###",
	"###.###",
	"#######",
}

func TestMapsAreSymmetricWithoutDeadEndsAndEveryByteIsReachable(t *testing.T) {
	if len(Maps) < 3 {
		t.Fatalf("%d maps, want at least 3", len(Maps))
	}
	for i, m := range Maps {
		t.Run(fmt.Sprintf("map %d", i+1), func(t *testing.T) { checkMap(t, m) })
	}
}

func checkMap(t *testing.T, rows []string) {
	g := newTestGame(rows...)
	if g.Width() != len(RAM[0]) || g.Height() != len(RAM) {
		t.Fatalf("map is %dx%d, want the size of RAM so the frame stays put", g.Width(), g.Height())
	}
	for y, row := range rows {
		for x := range row {
			if mirror := row[len(row)-1-x]; (row[x] == '#') != (mirror == '#') {
				t.Fatalf("row %d is not symmetric at column %d", y, x)
			}
		}
	}
	reach := grid.Distances(g.Width(), g.Height(), g.Leak().Cell, g.leakCanEnter)
	mallocs := 0
	for y := 0; y < g.Height(); y++ {
		for x := 0; x < g.Width(); x++ {
			p := pt(x, y)
			if !g.leakCanEnter(p) {
				continue
			}
			open := 0
			for _, d := range grid.Neighbours {
				if g.leakCanEnter(p.Add(d)) {
					open++
				}
			}
			if open < 2 {
				t.Fatalf("dead end at %v", p)
			}
			if reach.At(p) == grid.Unreachable {
				t.Fatalf("cell %v unreachable from the leak's start", p)
			}
			if g.Tile(p) == TileMalloc {
				mallocs++
			}
		}
	}
	if mallocs != 4 || len(g.Collectors()) != 3 {
		t.Fatalf("mallocs = %d, collectors = %d, want 4 and 3", mallocs, len(g.Collectors()))
	}
	if g.Tile(g.Leak().Cell) != TileFree {
		t.Fatalf("the leak must start on a cell with no byte")
	}
}

func TestNothingMovesBeforeTheFirstMove(t *testing.T) {
	g := newTestGame(RAM...)
	leak, collectors := g.Leak(), g.Collectors()
	steps(g, 300)
	if g.Started() || g.Leak() != leak {
		t.Fatalf("the leak moved before the first move")
	}
	for i, c := range g.Collectors() {
		if c.Cell != collectors[i].Cell || !c.InPool() {
			t.Fatalf("collector %d left the pool before the first move", i)
		}
	}
}

func TestLeakMovesAtItsSpeedInTheDirectionPressed(t *testing.T) {
	g := newTestGame(corridor...)
	g.Step(Input{Dir: Right})
	for i := 2; i <= 10; i++ {
		g.Step(Input{})
		if got, want := along(g.Leak()), CellUnits+i*LeakSpeed; got != want {
			t.Fatalf("tick %d: position %d, want %d", i, got, want)
		}
	}
}

func TestEarlyTurnIsTakenAtTheFirstOpenCell(t *testing.T) {
	g := newTestGame(corridor...)
	g.Step(Input{Dir: Right})
	g.Step(Input{Dir: Down}) // blocked at (1,1) and (2,1), open at (3,1)
	for i := 0; i < 30 && g.Leak().Cell.Y == 1; i++ {
		g.Step(Input{})
	}
	if l := g.Leak(); l.Cell != pt(3, 2) || l.Dir != Down {
		t.Fatalf("leak at %v heading %v, want turned down at column 3", l.Cell, l.Dir)
	}
}

func TestReverseIsImmediateAndContinuous(t *testing.T) {
	g := newTestGame(corridor...)
	g.Step(Input{Dir: Right})
	steps(g, 2)
	before := along(g.Leak())
	g.Step(Input{Dir: Left})
	if l := g.Leak(); l.Dir != Left || along(l) != before-LeakSpeed {
		t.Fatalf("after reversing: dir %v at %d, want left at %d", l.Dir, along(l), before-LeakSpeed)
	}
}

func TestWallAheadStopsTheLeak(t *testing.T) {
	g := newTestGame(corridor...)
	g.Step(Input{Dir: Right})
	steps(g, 60)
	if l := g.Leak(); l.Cell != pt(5, 1) || l.Progress != 0 || l.Dir != (grid.Point{}) {
		t.Fatalf("leak = %+v, want stopped at the wall on (5,1)", l)
	}
}

func TestEatingAByteScoresAndClearsIt(t *testing.T) {
	g := newTestGame(corridor...)
	left := g.BytesLeft()
	g.Step(Input{Dir: Right})
	steps(g, 3) // past half a cell: the leak is in (2,1)
	if g.Score() != PointsByte || g.Tile(pt(2, 1)) != TileFree || g.BytesLeft() != left-1 {
		t.Fatalf("score %d, tile %v, bytes %d, want the byte at (2,1) eaten", g.Score(), g.Tile(pt(2, 1)), g.BytesLeft())
	}
}

// pooled has a pool the leak cannot reach, so collectors never catch it.
var pooled = []string{
	"#######",
	"#.....#",
	"###-###",
	"##GGG##",
	"#######",
	"#L.####",
	"#######",
}

func TestCollectorsLeaveThePoolOneAfterAnother(t *testing.T) {
	g := newTestGame(pooled...)
	homes := g.Collectors()
	g.Step(Input{Dir: Right})
	steps(g, g.releaseInterval()-2)
	cs := g.Collectors()
	if cs[0].InPool() {
		t.Fatalf("the first collector must be out of the pool")
	}
	if cs[1].Cell != homes[1].Cell || cs[2].Cell != homes[2].Cell {
		t.Fatalf("collectors 2 and 3 left before their release")
	}
	steps(g, 61) // one tick past the second release, well before the third
	cs = g.Collectors()
	if cs[1].InPool() || cs[2].Cell != homes[2].Cell {
		t.Fatalf("after one more interval only the second collector must have left")
	}
	steps(g, g.releaseInterval())
	if g.Collectors()[2].InPool() {
		t.Fatalf("the third collector never left the pool")
	}
}

func TestCollectorsOutOfThePoolNeverGoBackIn(t *testing.T) {
	g := newTestGame(pooled...)
	g.Step(Input{Dir: Right})
	for i := 0; i < 1200; i++ {
		g.Step(Input{})
		for j, c := range g.Collectors() {
			if tile := g.Tile(c.Occupied()); !c.InPool() && (tile == TileGate || tile == TilePool) {
				t.Fatalf("tick %d: collector %d went back into the pool", i, j)
			}
		}
	}
}

var loop = []string{
	"#############",
	"#...........#",
	"#.#.#######.#",
	"#..L........#",
	"#############",
}

// firstMove places the leak, adds one collector at (6,1) and returns the
// direction the collector picks on the next tick, during a sweep.
func firstMove(t *testing.T, kind Kind, leak grid.Point, rng *scriptedRand) grid.Point {
	t.Helper()
	return moveAt(t, PauseTicks, kind, leak, rng)
}

// moveAt is firstMove at a given number of ticks into the life.
func moveAt(t *testing.T, lifeTicks int, kind Kind, leak grid.Point, rng *scriptedRand) grid.Point {
	t.Helper()
	g := newTestGame(loop...)
	g.rng = rng
	g.leak = Actor{Cell: leak, Dir: Right}
	g.started = true
	g.lifeTicks = lifeTicks
	addCollector(g, kind, pt(6, 1), grid.Point{})
	g.Step(Input{})
	return g.Collectors()[0].Dir
}

func TestTracerTakesTheShortestWayToTheLeak(t *testing.T) {
	if d := firstMove(t, KindTracer, pt(6, 3), &scriptedRand{}); d != Left {
		t.Fatalf("tracer went %v, want left (7 cells against 11)", d)
	}
}

func TestInterceptorAimsAheadOfTheLeak(t *testing.T) {
	// Ahead of (6,3) moving right is (10,3): 7 cells by the right, 11 by the left.
	if d := firstMove(t, KindInterceptor, pt(6, 3), &scriptedRand{}); d != Right {
		t.Fatalf("interceptor went %v, want right", d)
	}
}

func TestWandererChasesOnlyWithinSight(t *testing.T) {
	near := &scriptedRand{values: []int{1}}
	if d := firstMove(t, KindWanderer, pt(3, 1), near); d != Left || near.calls != 0 {
		t.Fatalf("leak 3 cells away: wanderer went %v with %d random picks, want a chase to the left", d, near.calls)
	}
	far := &scriptedRand{values: []int{1}}
	if d := firstMove(t, KindWanderer, pt(6, 3), far); d != Right || far.calls != 1 {
		t.Fatalf("leak 8 cells away: wanderer went %v with %d random picks, want the random pick (right)", d, far.calls)
	}
}

func TestCollectorsPauseInTheirCornersBetweenSweeps(t *testing.T) {
	// The leak is 7 cells to the left; the Tracer's corner is top right.
	for _, c := range []struct {
		lifeTicks int
		want      grid.Point
	}{
		{0, Right},                          // every life starts with a pause
		{PauseTicks - 1, Right},             // last tick of the pause
		{PauseTicks, Left},                  // the sweep chases the leak
		{PauseTicks + SweepTicks - 1, Left}, // last tick of the sweep
		{PauseTicks + SweepTicks, Right},    // the next cycle pauses again
	} {
		if d := moveAt(t, c.lifeTicks-1, KindTracer, pt(6, 3), &scriptedRand{}); d != c.want {
			t.Fatalf("tick %d of the life: tracer went %v, want %v", c.lifeTicks, d, c.want)
		}
	}
	// The Interceptor pauses top left, the Wanderer bottom right.
	if d := moveAt(t, 0, KindInterceptor, pt(10, 3), &scriptedRand{}); d != Left {
		t.Fatalf("pausing interceptor went %v, want left to its corner", d)
	}
	if d := moveAt(t, 0, KindWanderer, pt(3, 1), &scriptedRand{}); d != Right {
		t.Fatalf("pausing wanderer went %v, want right to its corner", d)
	}
}

func TestCollectorsTurnBackOnlyAtADeadEnd(t *testing.T) {
	g := newTestGame(loop...)
	g.started = true
	g.leak = Actor{Cell: pt(9, 1)}
	addCollector(g, KindTracer, pt(6, 1), Left) // the leak is behind it
	g.Step(Input{})
	if d := g.Collectors()[0].Dir; d != Left {
		t.Fatalf("tracer went %v, want it to keep going left", d)
	}

	g = newTestGame("#####", "#L..#", "#####")
	g.started = true
	addCollector(g, KindTracer, pt(3, 1), Right)
	g.Step(Input{})
	if d := g.Collectors()[0].Dir; d != Left {
		t.Fatalf("at a dead end the tracer went %v, want back left", d)
	}
}

func TestMeetingACollectorCostsALifeKeepingEatenBytes(t *testing.T) {
	g := newTestGame(corridor...)
	addCollector(g, KindTracer, pt(4, 1), Left)
	g.Step(Input{Dir: Right})
	for i := 0; i < 30 && g.Lives() == StartLives; i++ {
		g.Step(Input{})
	}
	if g.Lives() != StartLives-1 {
		t.Fatalf("lives = %d, want one lost", g.Lives())
	}
	if g.Leak().Cell != pt(1, 1) || g.Started() || g.Tile(pt(2, 1)) != TileFree {
		t.Fatalf("after the hit the leak must wait at its start with eaten bytes kept")
	}
}

func TestCrossingACollectorCostsALife(t *testing.T) {
	g := newTestGame(corridor...)
	g.started = true
	g.leak = Actor{Cell: pt(1, 1), Dir: Right, Progress: CellUnits/2 - 4}
	addCollector(g, KindTracer, pt(2, 1), Left).Progress = CellUnits/2 - 4
	g.Step(Input{})
	if g.Lives() != StartLives-1 {
		t.Fatalf("leak and collector swapped cells without a hit")
	}
}

func TestLosingTheLastLifeEndsTheRound(t *testing.T) {
	g := newTestGame(corridor...)
	g.lives = 1
	addCollector(g, KindTracer, pt(2, 1), Left)
	g.Step(Input{Dir: Right})
	steps(g, 10)
	if !g.Over() {
		t.Fatalf("losing the last life must end the round")
	}
	score := g.Score()
	g.Step(Input{Dir: Right})
	if g.Score() != score {
		t.Fatalf("a finished round advanced")
	}
}

var mallocLine = []string{
	"#########",
	"#Lo.....#",
	"#########",
}

func TestMallocTurnsActiveCollectorsBackIntoPrey(t *testing.T) {
	g := newTestGame(mallocLine...)
	addCollector(g, KindTracer, pt(7, 1), Left)
	g.Step(Input{Dir: Right})
	for i := 0; i < 10 && g.FrightLeft() == 0; i++ {
		g.Step(Input{})
	}
	c := g.Collectors()[0]
	if !c.Fleeing() || c.Dir != Right || g.Score() != PointsMalloc {
		t.Fatalf("fleeing=%v dir=%v score=%d, want the collector fleeing right after a malloc", c.Fleeing(), c.Dir, g.Score())
	}
	if g.FrightLeft() <= g.frightTicks()-10 {
		t.Fatalf("malloc lasts %d ticks, want %d", g.FrightLeft(), g.frightTicks())
	}
	before := along(c.Actor)
	g.Step(Input{})
	if got := along(g.Collectors()[0].Actor); got != before+FleeSpeed {
		t.Fatalf("fleeing collector moved %d units, want %d", got-before, FleeSpeed)
	}
}

func TestMallocWearsOff(t *testing.T) {
	g := newTestGame(mallocLine...)
	g.started = true
	g.fright = 2
	addCollector(g, KindTracer, pt(6, 1), Left).fleeing = true
	steps(g, 2)
	if g.FrightLeft() != 0 || g.Collectors()[0].Fleeing() {
		t.Fatalf("the malloc must wear off")
	}
}

func TestEatingFleeingCollectorsDoublesThePointsAndSendsThemHome(t *testing.T) {
	// The sealed byte on the last row keeps the level from being cleared.
	g := newTestGame("#############", "#L..........#", "#############", "#.###########", "#############")
	g.fright = 1000
	for _, x := range []int{3, 5} {
		c := addCollector(g, KindTracer, pt(x, 1), Right)
		c.fleeing = true
		c.home = pt(7, 1)
	}
	g.Step(Input{Dir: Right})
	steps(g, 90)
	bytes := (11 - g.BytesLeft()) * PointsByte
	if got := g.Score() - bytes; got != PointsCollector+2*PointsCollector {
		t.Fatalf("collector points = %d, want %d", got, 3*PointsCollector)
	}
	for i, c := range g.Collectors() {
		if !c.InPool() || c.Fleeing() {
			t.Fatalf("eaten collector %d not sent home", i)
		}
	}
}

func TestEatingEverythingStartsTheNextLevelRefilled(t *testing.T) {
	g := newTestGame("#####", "#L.o#", "#####")
	g.Step(Input{Dir: Right})
	for i := 0; i < 60 && g.Level() == 1; i++ {
		g.Step(Input{})
	}
	if g.Level() != 2 || g.BytesLeft() != 2 || g.Score() != PointsByte+PointsMalloc {
		t.Fatalf("level %d, bytes %d, score %d, want level 2 refilled keeping the score", g.Level(), g.BytesLeft(), g.Score())
	}
	if g.Started() || g.Leak().Cell != pt(1, 1) || g.Lives() != StartLives {
		t.Fatalf("a new level must wait at the start with lives kept")
	}
}

func TestEachLevelPlaysTheNextMapInTurn(t *testing.T) {
	one := []string{"#####", "#L.o#", "#####"}
	two := []string{"#####", "#o.L#", "#####"}
	g := NewGame(Config{Maps: [][]string{one, two}, Rand: &scriptedRand{}})
	clear := func(dir grid.Point) {
		g.Step(Input{Dir: dir})
		for level, i := g.Level(), 0; i < 60 && g.Level() == level; i++ {
			g.Step(Input{})
		}
	}
	for _, want := range []struct {
		dir   grid.Point
		level int
		start grid.Point
	}{{Right, 2, pt(3, 1)}, {Left, 3, pt(1, 1)}} {
		clear(want.dir)
		if g.Level() != want.level || g.Leak().Cell != want.start || g.BytesLeft() != 2 {
			t.Fatalf("level %d, leak at %v, bytes %d, want level %d starting at %v refilled", g.Level(), g.Leak().Cell, g.BytesLeft(), want.level, want.start)
		}
	}
}

func TestLevelsGetHarderAndLevelOffBelowTheLeakSpeed(t *testing.T) {
	g := newTestGame(RAM...)
	type pace struct{ speed, fright, release int }
	at := func(level int) pace {
		g.level = level
		return pace{g.collectorSpeed(), g.frightTicks(), g.releaseInterval()}
	}
	one, three, far := at(1), at(3), at(100)
	if three.speed <= one.speed || three.fright >= one.fright || three.release >= one.release {
		t.Fatalf("level 3 %+v not harder than level 1 %+v", three, one)
	}
	if far.speed >= LeakSpeed || far.fright <= 0 || far.release <= 0 {
		t.Fatalf("level 100 pace %+v must stay playable", far)
	}
}

// TestRealTimeSpeedsDoNotDependOnTheTickRate pins speeds per second, so a
// change of TickPeriod must rescale the per-tick constants with it.
func TestRealTimeSpeedsDoNotDependOnTheTickRate(t *testing.T) {
	if TickPeriod != time.Second/60 {
		t.Fatalf("tick period = %v, want 60 ticks per second", TickPeriod)
	}
	perMinute := func(speed int) int { return speed * int(time.Minute/TickPeriod) / CellUnits }
	g := newTestGame(RAM...)
	cases := []struct {
		name      string
		got, want int
	}{
		{"leak cells/min", perMinute(LeakSpeed), 480},
		{"level 1 collector cells/min", perMinute(g.collectorSpeed()), 390},
		{"fleeing collector cells/min", perMinute(FleeSpeed), 240},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Fatalf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
	if ms := g.frightTicks() * 1000 / int(time.Second/TickPeriod); ms != 7000 {
		t.Fatalf("level 1 malloc = %d ms, want 7000", ms)
	}
	if ms := g.releaseInterval() * 1000 / int(time.Second/TickPeriod); ms != 4000 {
		t.Fatalf("level 1 release interval = %d ms, want 4000", ms)
	}
}
