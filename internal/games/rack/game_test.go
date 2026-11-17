package rack

import (
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/platform"
)

// scriptedRand returns its values in turn, then 0.
type scriptedRand struct{ values []int }

func (r *scriptedRand) Intn(n int) int {
	if len(r.values) == 0 {
		return 0
	}
	v := r.values[0]
	r.values = r.values[1:]
	return v % n
}

func newTestGame(rows ...string) *Game {
	return NewGame(Config{Levels: [][]string{rows}, Rand: &scriptedRand{}})
}

func pt(x, y int) grid.Point { return grid.Point{X: x, Y: y} }

var (
	keyLeft  = Input{Move: grid.Point{X: -1}}
	keyRight = Input{Move: grid.Point{X: 1}}
	keyUp    = Input{Move: grid.Point{Y: -1}}
	keyDown  = Input{Move: grid.Point{Y: 1}}
	keyJump  = Input{Jump: true}
)

// hold presses in every other tick, like a held key's repeats, for n ticks.
func hold(g *Game, in Input, n int) {
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			g.Step(in)
		} else {
			g.Step(Input{})
		}
	}
}

func steps(g *Game, n int) {
	for i := 0; i < n; i++ {
		g.Step(Input{})
	}
}

// room: a floor, a ladder up to a top floor with a gap on the right.
var room = []string{
	"          ",
	"   E      ",
	"=====H==  ",
	"     H    ",
	" P   H  $ ",
	"==========",
}

func TestLevelsAreReachableAndShaped(t *testing.T) {
	if len(Levels) < 3 {
		t.Fatalf("%d levels, want at least 3", len(Levels))
	}
	for i, rows := range Levels {
		g := NewGame(Config{Levels: [][]string{rows}, Rand: &scriptedRand{}})
		if g.Width() != Width || g.Height() != Height {
			t.Fatalf("level %d is %dx%d, want %dx%d", i+1, g.Width(), g.Height(), Width, Height)
		}
		reach := reachable(g)
		for y := 0; y < g.Height(); y++ {
			for x := 0; x < g.Width(); x++ {
				if c := g.Cell(pt(x, y)); (c == CellDisk || c == CellExit || c == CellToolbox) && !reach[pt(x, y)] {
					t.Fatalf("level %d: %v at %v cannot be reached from the start", i+1, c, pt(x, y))
				}
			}
		}
		tools := 0
		for y := 0; y < g.Height(); y++ {
			for x := 0; x < g.Width(); x++ {
				if g.Cell(pt(x, y)) == CellToolbox {
					tools++
				}
			}
		}
		if tools != 1 {
			t.Fatalf("level %d has %d toolboxes, want 1", i+1, tools)
		}
		if g.DisksLeft() < 5 || len(g.Bots()) == 0 {
			t.Fatalf("level %d has %d disks and %d bots", i+1, g.DisksLeft(), len(g.Bots()))
		}
	}
}

// reachable is every cell the technician can get to by walking, climbing
// and falling from the start.
func reachable(g *Game) map[grid.Point]bool {
	seen := map[grid.Point]bool{g.Player().Pos: true}
	queue := []grid.Point{g.Player().Pos}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, n := range g.moves(p) {
			if !seen[n] {
				seen[n] = true
				queue = append(queue, n)
			}
		}
	}
	return seen
}

func TestWalkingMovesAlongTheFloorAndStopsAtWalls(t *testing.T) {
	g := newTestGame(room...)
	g.Step(keyRight)
	if g.Player().Pos != pt(2, 4) {
		t.Fatalf("a tap must step at once, at %v", g.Player().Pos)
	}
	hold(g, keyLeft, 60)
	if g.Player().Pos != pt(0, 4) {
		t.Fatalf("walking left must stop at the wall, at %v", g.Player().Pos)
	}
}

func TestWalkingSpeedIsSteadyWhileHeld(t *testing.T) {
	g := newTestGame("            ", "P           ", "============")
	hold(g, keyRight, 60)
	if x := g.Player().Pos.X; x < 7 || x > 9 {
		t.Fatalf("walked %d cells in a second, want about 8", x)
	}
}

func TestClimbingUpAndDownALadder(t *testing.T) {
	g := newTestGame(room...)
	g.player.Pos = pt(5, 4)
	hold(g, keyUp, 60)
	if g.Player().Pos != pt(5, 1) {
		t.Fatalf("climbing up must reach the top floor, at %v", g.Player().Pos)
	}
	hold(g, keyDown, 60)
	if g.Player().Pos != pt(5, 4) {
		t.Fatalf("climbing down must reach the bottom, at %v", g.Player().Pos)
	}
	g.player.Pos = pt(3, 4)
	hold(g, keyUp, 30)
	if g.Player().Pos != pt(3, 4) {
		t.Fatalf("up off a ladder must not move, at %v", g.Player().Pos)
	}
}

func TestWalkingOffAnEdgeFallsWithoutSteering(t *testing.T) {
	g := newTestGame(room...)
	g.player.Pos = pt(7, 1)
	g.Step(keyRight) // over the gap
	if g.Player().Pos != pt(8, 1) {
		t.Fatalf("at %v, want over the gap", g.Player().Pos)
	}
	hold(g, keyLeft, FallTicks*2)
	if p := g.Player().Pos; p.X != 8 || p.Y != 3 {
		t.Fatalf("falling for %d ticks at %v, want straight down two cells", FallTicks*2, p)
	}
}

func TestJumpRisesACellForAMomentKeepingTheWalk(t *testing.T) {
	g := newTestGame("            ", "            ", "P           ", "============")
	g.Step(keyJump)
	if g.Player().Pos != pt(0, 1) || !g.Player().Jumping() {
		t.Fatalf("jump at %v, want one cell up", g.Player().Pos)
	}
	hold(g, keyRight, JumpTicks-1)
	if p := g.Player().Pos; p.Y != 1 || p.X < 2 {
		t.Fatalf("during the jump at %v, want still up and walking", p)
	}
	steps(g, FallTicks+1)
	if p := g.Player().Pos; p.Y != 2 {
		t.Fatalf("after the jump at %v, want back on the floor", p)
	}
	g.player.Pos = pt(5, 2)
	g.Step(keyJump)
	g.Step(keyJump)
	if g.Player().Pos.Y != 1 {
		t.Fatalf("a jump in the air must not go higher, at %v", g.Player().Pos)
	}
}

func TestCollectingEveryDiskOpensTheDoorAndTheDoorEndsTheLevel(t *testing.T) {
	g := newTestGame(room...)
	g.player.Pos = pt(7, 4)
	g.Step(keyRight)
	if g.DisksLeft() != 0 || g.Score() != PointsDisk || !g.DoorOpen() {
		t.Fatalf("disks %d score %d open %v, want the disk taken and the door open", g.DisksLeft(), g.Score(), g.DoorOpen())
	}
	g.player.Pos = pt(4, 1)
	g.Step(keyLeft)
	if g.Level() != 2 || g.Score() != PointsDisk+PointsExit || g.DisksLeft() != 1 || g.Player().Pos != pt(1, 4) {
		t.Fatalf("level %d score %d disks %d at %v, want level 2 refilled from the start", g.Level(), g.Score(), g.DisksLeft(), g.Player().Pos)
	}
}

func TestAClosedDoorIsJustAPassage(t *testing.T) {
	g := newTestGame(room...)
	g.player.Pos = pt(4, 1)
	g.Step(keyLeft)
	if g.Level() != 1 || g.Player().Pos != pt(3, 1) {
		t.Fatalf("level %d at %v, want to walk past the closed door", g.Level(), g.Player().Pos)
	}
}

// slope: a spool at the top, floors with alternating gaps, a ladder.
var slope = []string{
	"            ",
	" C          ",
	"=========H  ",
	"         H  ",
	"         H  ",
	"  ==========",
	" P          ",
	"============",
}

func TestCablesRollFallLandAndRollBack(t *testing.T) {
	g := newTestGame(slope...)
	g.player.Pos = pt(11, 6)
	g.spawnCable()
	c := g.Cables()[0]
	if c.Pos != pt(1, 1) || c.Dir != 1 {
		t.Fatalf("cable %+v, want at the spool rolling right", c)
	}
	// scriptedRand returns 0 and LadderChance is not 1, so the cable keeps
	// rolling past the ladder top: Intn(LadderChance) == 0 takes the ladder,
	// so script a 1.
	g.rng = &scriptedRand{values: []int{1, 1, 1}}
	for i := 0; i < 400 && g.Cables()[0].Pos.Y < 4; i++ {
		g.moveCables()
	}
	c = g.Cables()[0]
	if c.Pos.X < 10 || c.Pos.Y != 4 || c.Dir != -1 {
		t.Fatalf("cable %+v, want landed right of the ladder and rolling back left", c)
	}
}

func TestCablesSometimesTakeALadder(t *testing.T) {
	g := newTestGame(slope...)
	g.player.Pos = pt(11, 6)
	g.spawnCable()
	g.rng = &scriptedRand{values: []int{0}}
	for i := 0; i < 400 && g.Cables()[0].Pos.X < 9; i++ {
		g.moveCables()
	}
	for i := 0; i < 3; i++ {
		g.moveCables()
	}
	if c := g.Cables()[0]; c.Pos != pt(9, 4) {
		t.Fatalf("cable %+v, want down the ladder at x 9", c)
	}
	g.moveCables()
	g.moveCables()
	if c := g.Cables()[0]; c.Pos != pt(8, 4) || c.Dir != -1 {
		t.Fatalf("cable %+v, want rolling back from the ladder's foot", c)
	}
}

func TestCablesLeaveAtTheBottomFloorsEnds(t *testing.T) {
	g := newTestGame(slope...)
	g.cables = []Cable{{Pos: pt(1, 6), Dir: -1}}
	g.player.Pos = pt(11, 6)
	g.moveCables()
	g.moveCables()
	if len(g.Cables()) != 0 {
		t.Fatalf("cables %+v, want gone off the bottom floor's end", g.Cables())
	}
}

func TestBotsPatrolTheirFloorTurningAtWallsAndGaps(t *testing.T) {
	g := newTestGame(
		"            ",
		"   B        ",
		"=======     ",
		"            ",
		"============",
	)
	seen := map[int]bool{}
	for i := 0; i < 40; i++ {
		g.moveBots()
		b := g.Bots()[0]
		if b.Pos.Y != 1 || b.Pos.X > 6 || b.Pos.X < 0 {
			t.Fatalf("bot at %v left its floor", b.Pos)
		}
		seen[b.Pos.X] = true
	}
	if !seen[0] || !seen[6] {
		t.Fatalf("the bot must walk the whole floor, saw %v", seen)
	}
}

func TestTouchingAHazardCostsALifeAndRestartsTheLevelKeepingDisks(t *testing.T) {
	g := newTestGame(room...)
	g.player.Pos = pt(7, 4)
	g.Step(keyRight) // takes the disk
	steps(g, WalkGlide)
	g.player.Pos = pt(3, 4)
	g.cables = []Cable{{Pos: pt(4, 4), Dir: -1}}
	g.Step(keyRight)
	if g.Lives() != StartLives-1 || g.Player().Pos != pt(1, 4) || len(g.Cables()) != 0 || g.DisksLeft() != 0 {
		t.Fatalf("lives %d at %v cables %d disks %d, want a life lost, back at the start, cables cleared, disks kept", g.Lives(), g.Player().Pos, len(g.Cables()), g.DisksLeft())
	}
	g.lives = 1
	g.bots = []Bot{{Pos: pt(2, 4), Dir: -1}}
	g.Step(keyRight)
	if !g.Over() {
		t.Fatalf("losing the last life must end the round")
	}
}

func TestJumpingOverACableScores(t *testing.T) {
	g := newTestGame("            ", "            ", "     P      ", "============")
	g.cables = []Cable{{Pos: pt(6, 2), Dir: -1}}
	g.Step(keyJump)
	for i := 0; i < JumpTicks && g.Score() == 0 && !g.Over(); i++ {
		g.Step(Input{})
	}
	if g.Score() != PointsJump || g.Lives() != StartLives {
		t.Fatalf("score %d lives %d, want %d for the jump and no life lost", g.Score(), g.Lives(), PointsJump)
	}
	steps(g, JumpTicks)
	if g.Score() != PointsJump {
		t.Fatalf("score %d after the jump, want one cable scored once", g.Score())
	}
}

func TestLaterLevelsSendCablesMoreOftenAndFaster(t *testing.T) {
	g := newTestGame(room...)
	one := [2]int{g.cableInterval(), g.cableTicks()}
	g.level = 3
	three := [2]int{g.cableInterval(), g.cableTicks()}
	g.level = 100
	far := [2]int{g.cableInterval(), g.cableTicks()}
	if three[0] >= one[0] || three[1] >= one[1] || far[0] < MinCableInterval || far[1] < MinCableTicks {
		t.Fatalf("pace by level: 1 %v, 3 %v, 100 %v", one, three, far)
	}
}

// TestRealTimeSpeedsDoNotDependOnTheTickRate pins speeds per second.
func TestRealTimeSpeedsDoNotDependOnTheTickRate(t *testing.T) {
	perSecond := int(time.Second / TickPeriod)
	if got := perSecond * WalkSpeed / 100; got < 7 || got > 8 {
		t.Fatalf("walk %d cells/s, want about 8", got)
	}
	if got := perSecond / FallTicks; got != 15 {
		t.Fatalf("fall %d cells/s, want 15", got)
	}
	if got := time.Duration(JumpTicks) * TickPeriod; got.Round(time.Millisecond) != 350*time.Millisecond {
		t.Fatalf("jump %v, want 350ms", got)
	}
	if got := perSecond / BotTicks; got != 4 {
		t.Fatalf("bots %d cells/s, want 4", got)
	}
}

// moves are the cells the technician can reach from p in one move: down
// when nothing holds it, otherwise a walk or a climb.
func (g *Game) moves(p grid.Point) []grid.Point {
	if !platform.Supported(g, p) {
		return []grid.Point{p.Add(down)}
	}
	var out []grid.Point
	for _, d := range []grid.Point{{X: -1}, {X: 1}} {
		if platform.Open(g, p.Add(d)) {
			out = append(out, p.Add(d))
		}
	}
	if platform.CanClimbUp(g, p) {
		out = append(out, p.Add(up))
	}
	if platform.CanClimbDown(g, p) {
		out = append(out, p.Add(down))
	}
	return out
}

// shed: a floor with a toolbox, a bot and a wall on the right.
var shed = []string{
	"            ",
	" PT      B =",
	"============",
}

func TestTheToolboxGivesWrenchesToThrow(t *testing.T) {
	g := newTestGame(shed...)
	if g.Cell(pt(2, 1)) != CellToolbox || g.WrenchesLeft() != 0 {
		t.Fatalf("cell %v wrenches %d, want a toolbox and none in hand", g.Cell(pt(2, 1)), g.WrenchesLeft())
	}
	g.Step(keyRight)
	if g.WrenchesLeft() != WrenchesPerToolbox || g.Cell(pt(2, 1)) != CellEmpty {
		t.Fatalf("wrenches %d cell %v, want %d and the toolbox gone", g.WrenchesLeft(), g.Cell(pt(2, 1)), WrenchesPerToolbox)
	}
}

func TestAThrowWithoutWrenchesDoesNothing(t *testing.T) {
	g := newTestGame(shed...)
	g.Step(Input{Throw: true})
	if len(g.Thrown()) != 0 {
		t.Fatalf("threw %v with no wrenches", g.Thrown())
	}
}

func TestAWrenchFliesTheWayTheTechnicianFacesAndDisablesABot(t *testing.T) {
	g := newTestGame(shed...)
	g.Step(keyRight) // picks up the toolbox, facing right
	g.Step(Input{Throw: true})
	if g.WrenchesLeft() != WrenchesPerToolbox-1 || len(g.Thrown()) != 1 {
		t.Fatalf("wrenches %d in flight %v, want one thrown", g.WrenchesLeft(), g.Thrown())
	}
	g.Step(Input{Throw: true})
	if g.WrenchesLeft() != WrenchesPerToolbox-1 {
		t.Fatalf("a second throw while one flies must wait, wrenches %d", g.WrenchesLeft())
	}
	for i := 0; i < 60 && len(g.Thrown()) > 0; i++ {
		g.Step(Input{})
	}
	if len(g.Bots()) != 0 || g.Score() != PointsBot {
		t.Fatalf("bots %v score %d, want the bot disabled for %d", g.Bots(), g.Score(), PointsBot)
	}
}

func TestAWrenchStopsAtAWallAndFliesOverGaps(t *testing.T) {
	g := newTestGame(
		"              ",
		" P            ",
		"=====   ======",
	)
	g.wrenches = 1
	g.Step(Input{Throw: true}) // faces right from the start
	far := 0
	for i := 0; i < 120 && len(g.Thrown()) > 0; i++ {
		far = max(far, g.Thrown()[0].Pos.X)
		if y := g.Thrown()[0].Pos.Y; y != 1 {
			t.Fatalf("wrench at row %d, want flying straight along row 1", y)
		}
		g.Step(Input{})
	}
	if far != 13 || len(g.Thrown()) != 0 {
		t.Fatalf("wrench reached x %d, want to fly over the gap to the last cell and stop at the wall", far)
	}
}

func TestABotWalkingIntoAWrenchIsDisabled(t *testing.T) {
	g := newTestGame(shed...)
	g.thrown = []Wrench{{Pos: pt(5, 1), Dir: 1}}
	g.bots = []Bot{{Pos: pt(6, 1), Dir: -1}}
	g.botClock = BotTicks - 1 // the bot steps onto the wrench this tick, before it flies on
	g.Step(Input{})
	if len(g.Bots()) != 0 || g.Score() != PointsBot {
		t.Fatalf("bots %v score %d, want the bot that walked into the wrench disabled", g.Bots(), g.Score())
	}
}

func TestAWrenchKnocksACableAway(t *testing.T) {
	g := newTestGame(shed...)
	g.bots = nil
	g.wrenches = 1
	g.player.Pos = pt(4, 1)
	g.Step(keyLeft)
	g.cables = []Cable{{Pos: pt(1, 1), Dir: 1}}
	g.Step(Input{Throw: true})
	for i := 0; i < 30 && len(g.Thrown()) > 0; i++ {
		g.Step(Input{})
	}
	if len(g.Cables()) != 0 || g.Score() != PointsCable || g.Lives() != StartLives {
		t.Fatalf("cables %v score %d lives %d, want the cable knocked away for %d", g.Cables(), g.Score(), g.Lives(), PointsCable)
	}
}

func TestDisabledBotsComeBackAfterALostLife(t *testing.T) {
	g := newTestGame(shed...)
	g.bots = nil
	g.thrown = []Wrench{{Pos: pt(6, 1), Dir: 1}}
	g.cables = []Cable{{Pos: g.Player().Pos.Add(grid.Point{X: 1})}}
	g.Step(keyRight)
	if len(g.Thrown()) != 0 {
		t.Fatalf("wrenches %v still in flight after a lost life", g.Thrown())
	}
	if g.Lives() != StartLives-1 || len(g.Bots()) != 1 || g.Bots()[0].Pos != pt(9, 1) {
		t.Fatalf("lives %d bots %v, want a life lost and the bot back at its post", g.Lives(), g.Bots())
	}
	if g.WrenchesLeft() != WrenchesPerToolbox || g.Cell(pt(2, 1)) == CellToolbox {
		t.Fatalf("wrenches %d toolbox %v, want the toolbox's wrenches kept and it not back", g.WrenchesLeft(), g.Cell(pt(2, 1)))
	}
}

func TestWrenchSpeedIsPinned(t *testing.T) {
	if got := int(time.Second/TickPeriod) / WrenchTicks; got != 20 {
		t.Fatalf("wrench %d cells/s, want 20", got)
	}
}
