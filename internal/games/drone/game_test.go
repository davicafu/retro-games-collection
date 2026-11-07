package drone

import (
	"reflect"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/scroll"
)

const (
	testWidth  = 30
	testHeight = 12
	testDroneX = 5
)

func newTestGame(antennas ...Antenna) *Game {
	return NewGame(Config{Width: testWidth, Height: testHeight, DroneX: testDroneX, Antennas: antennas})
}

// positions steps the game once per thrust flag and records the drone position after each tick.
func positions(g *Game, thrusts ...bool) []int {
	out := make([]int, 0, len(thrusts))
	for _, t := range thrusts {
		g.Step(t)
		out = append(out, g.Position())
	}
	return out
}

func idle(n int) []bool { return make([]bool, n) }

func TestDroneHoversAndWorldWaitsUntilFirstThrust(t *testing.T) {
	g := newTestGame(Antenna{X: 20, GapTop: 3})
	start := g.Position()

	positions(g, idle(50)...)

	if g.Position() != start || g.Started() {
		t.Fatalf("drone moved before the first thrust: %d -> %d", start, g.Position())
	}
	if g.Antennas()[0].X != 20 {
		t.Fatalf("world scrolled before the first thrust")
	}
	if start != testHeight/2*SubCells {
		t.Fatalf("start position = %d, want middle row %d", start, testHeight/2*SubCells)
	}
}

func TestThrustImpulseThenGravity(t *testing.T) {
	g := newTestGame()
	p := g.Position()

	got := positions(g, true, false, false)

	// Thrust sets velocity to -Thrust; then gravity adds Gravity per tick.
	want := []int{p - Thrust, p - Thrust - (Thrust - Gravity), p - Thrust - (Thrust - Gravity) - (Thrust - 2*Gravity)}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("positions = %v, want %v", got, want)
	}
}

func TestThrustReplacesVelocityInsteadOfAddingUp(t *testing.T) {
	g := newTestGame()
	positions(g, true, true)
	if g.Velocity() != -Thrust {
		t.Fatalf("velocity after two thrusts = %d, want %d", g.Velocity(), -Thrust)
	}
}

func TestFallSpeedIsCapped(t *testing.T) {
	g := NewGame(Config{Width: testWidth, Height: 200, DroneX: testDroneX})
	positions(g, append([]bool{true}, idle(60)...)...)
	if g.Velocity() != MaxFall {
		t.Fatalf("velocity = %d, want cap %d", g.Velocity(), MaxFall)
	}
}

func TestCeilingStopsTheDrone(t *testing.T) {
	g := newTestGame()
	for i := 0; i < 30; i++ {
		g.Step(true)
	}
	if g.Dead() {
		t.Fatalf("the ceiling must not end the round")
	}
	if g.Position() != 0 || g.Row() != 0 {
		t.Fatalf("position = %d, want clamped to 0", g.Position())
	}
	g.Step(false)
	if g.Velocity() != Gravity {
		t.Fatalf("velocity after the ceiling = %d, want reset then one tick of gravity %d", g.Velocity(), Gravity)
	}
}

func TestTouchingTheGroundEndsTheRound(t *testing.T) {
	g := newTestGame()
	g.Step(true)
	for i := 0; i < 200 && !g.Dead(); i++ {
		g.Step(false)
	}
	if !g.Dead() {
		t.Fatalf("falling must end the round")
	}
	if g.Row() < testHeight {
		t.Fatalf("died at row %d, above the ground row %d", g.Row(), testHeight)
	}
}

// hoverCycle is a thrust period whose net drift is zero: with Thrust 45 and
// Gravity 5 the drone rises and falls back to where it thrusted after 19 ticks.
const hoverCycle = 19

// scrollUntil flies the drone, thrusting every thrustEvery ticks, until done or death.
func scrollUntil(g *Game, thrustEvery int, done func() bool) {
	for tick := 0; tick < 2000 && !g.Dead() && !done(); tick++ {
		g.Step(tick%thrustEvery == 0)
	}
}

func TestHittingAMastEndsTheRound(t *testing.T) {
	// Gap rows 8..12; the drone hovers in rows 3..6, so it hits the upper mast.
	g := newTestGame(Antenna{X: testDroneX + 3, GapTop: 8})
	scrollUntil(g, hoverCycle, func() bool { return false })
	if !g.Dead() {
		t.Fatalf("expected a collision with the mast")
	}
	if a := g.Antennas()[0]; a.X > testDroneX || a.X+AntennaWidth <= testDroneX {
		t.Fatalf("died away from the antenna: antenna X=%d, drone X=%d", a.X, testDroneX)
	}
}

func TestFlyingThroughTheGapScores(t *testing.T) {
	// Gap rows 2..6 around the drone's hover band (rows 3..6).
	g := newTestGame(Antenna{X: testDroneX + 3, GapTop: 2})
	scrollUntil(g, hoverCycle, func() bool { return g.Score() == 1 })
	if g.Dead() || g.Score() != 1 {
		t.Fatalf("dead=%v score=%d row=%d, want alive with score 1", g.Dead(), g.Score(), g.Row())
	}
	if a := g.Antennas()[0]; a.X+AntennaWidth-1 != testDroneX-1 {
		t.Fatalf("scored with antenna right edge at %d, want just past the drone (%d)", a.X+AntennaWidth-1, testDroneX-1)
	}
}

func TestAntennasLeavingTheScreenAreDropped(t *testing.T) {
	g := newTestGame(Antenna{X: 0, GapTop: 0})
	g.Step(true)
	for i := 0; i < 20 && len(g.Antennas()) > 0; i++ {
		g.Step(i%9 == 0)
	}
	if len(g.Antennas()) != 0 {
		t.Fatalf("off-screen antenna kept: %+v", g.Antennas())
	}
}

func TestSpeedRampsWithDistanceAndCaps(t *testing.T) {
	g := newTestGame()
	if g.Speed() != BaseSpeed {
		t.Fatalf("initial speed = %d, want %d", g.Speed(), BaseSpeed)
	}
	g.distance = SpeedDistance
	if g.Speed() != BaseSpeed+SpeedStep {
		t.Fatalf("speed at %d cells = %d, want %d", SpeedDistance, g.Speed(), BaseSpeed+SpeedStep)
	}
	g.distance = 100000
	if g.Speed() != MaxSpeed {
		t.Fatalf("speed = %d, want cap %d", g.Speed(), MaxSpeed)
	}
}

func TestMaxSpeedNeverScrollsMoreThanOneCellPerTick(t *testing.T) {
	if MaxSpeed >= scroll.Unit {
		t.Fatalf("MaxSpeed %d must stay below %d so masts cannot tunnel through the drone", MaxSpeed, scroll.Unit)
	}
}

func TestDeadGameDoesNotAdvance(t *testing.T) {
	g := newTestGame(Antenna{X: 20, GapTop: 3})
	g.Step(true)
	g.dead = true
	pos, x := g.Position(), g.Antennas()[0].X
	positions(g, idle(40)...)
	if g.Position() != pos || g.Antennas()[0].X != x {
		t.Fatalf("dead game advanced")
	}
}

func TestGroundRowIsTheFirstRowBelowTheSky(t *testing.T) {
	g := newTestGame()
	g.Step(true)
	for i := 0; i < 200 && !g.Dead(); i++ {
		g.Step(false)
		if !g.Dead() && g.Row() >= testHeight {
			t.Fatalf("alive at row %d, the ground is row %d", g.Row(), testHeight)
		}
	}
	if g.Row() != testHeight {
		t.Fatalf("died at row %d, want exactly the ground row %d", g.Row(), testHeight)
	}
}

func TestGapEdgesAreExact(t *testing.T) {
	a := Antenna{X: testDroneX, GapTop: 3}
	cases := map[int]bool{2: true, 3: false, 3 + GapSize - 1: false, 3 + GapSize: true}
	for row, blocked := range cases {
		g := newTestGame(a)
		g.started = true
		g.position = row * SubCells
		if got := g.collides(); got != blocked {
			t.Fatalf("row %d: collides = %v, want %v (gap rows 3..%d)", row, got, blocked, 3+GapSize-1)
		}
	}
}
