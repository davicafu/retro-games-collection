package debris

import (
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/glide"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/motion"
)

// zeroRand always returns 0.
type zeroRand struct{}

func (zeroRand) Intn(int) int { return 0 }

const testCols, testRows = 40, 20

// newTestGame starts a game whose first wave is the given pieces.
func newTestGame(pieces ...Piece) *Game {
	return NewGame(Config{Cols: testCols, Rows: testRows, Rand: zeroRand{}, Pieces: pieces})
}

// far is a piece parked far from the drone (which starts in the centre), at rest.
func far(size Size) Piece {
	return Piece{Size: size, Body: motion.Body{Pos: motion.Vec{X: 5000, Y: 30000}}}
}

func steps(g *Game, n int, in Input) {
	for i := 0; i < n; i++ {
		g.Step(in)
	}
}

func TestDroneStartsInTheCentreAtRestFacingUp(t *testing.T) {
	g := newTestGame(far(Large))
	d := g.Drone()
	if d.Pos != (motion.Vec{X: testCols * CellW / 2, Y: testRows * CellH / 2}) || d.Vel != (motion.Vec{}) || d.Heading != 0 {
		t.Fatalf("drone = %+v, want the centre, at rest, heading 0", d)
	}
}

func TestATurnTapTurnsOneHeadingThenStops(t *testing.T) {
	g := newTestGame(far(Large))
	g.Step(Input{Turn: 1})
	if g.Drone().Heading != 1 {
		t.Fatalf("a tap must turn on its first tick, heading %d", g.Drone().Heading)
	}
	steps(g, TurnGlide, Input{})
	tap := g.Drone().Heading
	steps(g, 60, Input{})
	if tap != 1 || g.Drone().Heading != tap {
		t.Fatalf("a tap turned %d headings and then %d, want exactly 1 and then stop", tap, g.Drone().Heading)
	}
	g.Step(Input{Turn: -1})
	if g.Drone().Heading != tap-1 {
		t.Fatalf("a tap the other way must turn back at once, heading %d", g.Drone().Heading)
	}
}

func TestHoldingATurnKeyTurnsSteadily(t *testing.T) {
	g := newTestGame(far(Large))
	turned, last := 0, 0
	// A held key repeats every other tick.
	for tick := 0; tick < 120; tick++ {
		in := Input{}
		if tick%2 == 0 {
			in.Turn = -1
		}
		g.Step(in)
		if h := g.Drone().Heading; h != last {
			turned++
			last = h
		}
	}
	if want := 120 * TurnSpeed / 100; turned < want-2 || turned > want+2 {
		t.Fatalf("held for 2 s the drone turned %d headings, want about %d", turned, want)
	}
}

func TestThrustBurstAcceleratesAlongTheHeadingThenTheDroneDrifts(t *testing.T) {
	g := newTestGame(far(Large))
	g.drone.Heading = motion.Headings / 4 // right
	g.Step(Input{Thrust: true})
	steps(g, ThrustTicks-1, Input{})
	burst := g.Drone().Vel
	if burst.X <= 0 || burst.Y != 0 {
		t.Fatalf("velocity after a burst %v, want to the right", burst)
	}
	g.Step(Input{})
	if v := g.Drone().Vel; v.X >= burst.X {
		t.Fatalf("after the burst, drag must slow the drone: %v then %v", burst, v)
	}
	x := g.Drone().Pos.X
	steps(g, 60, Input{})
	if v := g.Drone().Vel; v.X <= 0 || g.Drone().Pos.X == x {
		t.Fatalf("the drone must keep drifting a second later, velocity %v", v)
	}
}

func TestHoldingThrustReachesAndKeepsTheTopSpeed(t *testing.T) {
	g := newTestGame(far(Large))
	g.drone.Heading = 6
	steps(g, 600, Input{Thrust: true})
	if l := g.Drone().Vel.Len(); l > MaxDroneSpeed || l < MaxDroneSpeed*9/10 {
		t.Fatalf("speed %d, want close to and at most %d", l, MaxDroneSpeed)
	}
}

func TestTheDroneWrapsAcrossTheEdges(t *testing.T) {
	g := newTestGame(far(Large))
	g.drone.Pos = motion.Vec{X: testCols*CellW - 10, Y: 5}
	g.drone.Vel = motion.Vec{X: 50, Y: -50}
	g.Step(Input{})
	if p := g.Drone().Pos; p.X >= 100 || p.Y < testRows*CellH-100 {
		t.Fatalf("pos %v, want wrapped to the left and bottom edges", p)
	}
}

func TestFiringIsLimitedByCooldownAndShotsInFlight(t *testing.T) {
	g := newTestGame(far(Large))
	g.drone.Heading = motion.Headings / 4
	g.Step(Input{Fire: true})
	g.Step(Input{Fire: true})
	if n := len(g.Shots()); n != 1 {
		t.Fatalf("%d shots right after firing twice, want 1 (cooldown)", n)
	}
	most := 0
	for i := 0; i < 200; i++ {
		g.Step(Input{Fire: true})
		most = max(most, len(g.Shots()))
	}
	if most != MaxShots {
		t.Fatalf("at most %d shots at once, want %d", most, MaxShots)
	}
	steps(g, ShotTicks, Input{})
	if n := len(g.Shots()); n != 0 {
		t.Fatalf("%d shots left after their range, want 0", n)
	}
}

func TestShotsBreakDebrisIntoSmallerPiecesAndScore(t *testing.T) {
	// A large piece right of the drone, at rest: shoot it to the end.
	centre := motion.Vec{X: testCols * CellW / 2, Y: testRows * CellH / 2}
	target := Piece{Size: Large, Body: motion.Body{Pos: centre.Add(motion.Vec{X: 8000})}}
	g := newTestGame(target, far(Large))
	g.drone.Heading = motion.Headings / 4
	hits := []struct {
		score  int
		pieces map[Size]int
	}{
		{PointsLarge, map[Size]int{Large: 1, Medium: 2}},
	}
	for i := 0; i < 30 && g.Score() == 0; i++ {
		g.Step(Input{Fire: i == 0})
	}
	for _, h := range hits {
		if g.Score() != h.score {
			t.Fatalf("score %d, want %d", g.Score(), h.score)
		}
		count := map[Size]int{}
		for _, p := range g.Pieces() {
			count[p.Size]++
		}
		for size, n := range h.pieces {
			if count[size] != n {
				t.Fatalf("pieces %v, want %v", count, h.pieces)
			}
		}
	}
	if len(g.Shots()) != 0 {
		t.Fatalf("the shot that hit must be gone")
	}
}

func TestMediumBreaksIntoSmallAndSmallIsDestroyed(t *testing.T) {
	g := newTestGame(far(Large))
	g.pieces = []Piece{{Size: Medium, Body: motion.Body{Pos: motion.Vec{X: 5000, Y: 5000}}}}
	g.hit(0)
	if len(g.pieces) != 2 || g.pieces[0].Size != Small || g.pieces[1].Size != Small || g.Score() != PointsMedium {
		t.Fatalf("pieces %+v score %d, want two small pieces and %d", g.pieces, g.Score(), PointsMedium)
	}
	if g.pieces[0].Vel == g.pieces[1].Vel || g.pieces[0].Vel.Len() == 0 {
		t.Fatalf("the two halves must fly apart: %v and %v", g.pieces[0].Vel, g.pieces[1].Vel)
	}
	g.hit(0)
	if len(g.pieces) != 1 || g.Score() != PointsMedium+PointsSmall {
		t.Fatalf("a small piece must be destroyed: pieces %d score %d", len(g.pieces), g.Score())
	}
}

func TestTouchingDebrisCostsALifeAndTheDroneRespawnsShielded(t *testing.T) {
	centre := motion.Vec{X: testCols * CellW / 2, Y: testRows * CellH / 2}
	g := newTestGame(Piece{Size: Large, Body: motion.Body{Pos: centre}})
	g.drone.Pos = centre.Add(motion.Vec{X: 500})
	g.drone.Vel = motion.Vec{X: 40}
	g.Step(Input{})
	if g.Lives() != StartLives-1 {
		t.Fatalf("lives %d, want %d", g.Lives(), StartLives-1)
	}
	if d := g.Drone(); d.Pos != centre || d.Vel != (motion.Vec{}) || !g.Shielded() {
		t.Fatalf("drone %+v shielded %v, want back at the centre, at rest, shielded", d, g.Shielded())
	}
	steps(g, ShieldTicks-1, Input{})
	if g.Lives() != StartLives-1 {
		t.Fatalf("a shielded drone must not lose lives, lives %d", g.Lives())
	}
	g.Step(Input{})
	if g.Lives() != StartLives-2 {
		t.Fatalf("after the shield the drone is hit again, lives %d", g.Lives())
	}
	for i := 0; i < 10*ShieldTicks && !g.Over(); i++ {
		g.Step(Input{})
	}
	if !g.Over() || g.Lives() != 0 {
		t.Fatalf("losing the last life must end the round")
	}
}

func TestClearingAWaveStartsTheNextAfterAPause(t *testing.T) {
	g := newTestGame(Piece{Size: Small, Body: motion.Body{Pos: motion.Vec{X: 1000, Y: 2000}}})
	g.hit(0)
	steps(g, WaveDelay-1, Input{})
	if g.Wave() != 1 || len(g.Pieces()) != 0 {
		t.Fatalf("wave %d with %d pieces before the pause ends, want wave 1 empty", g.Wave(), len(g.Pieces()))
	}
	g.Step(Input{})
	if g.Wave() != 2 || len(g.Pieces()) != PiecesAtWave(2) {
		t.Fatalf("wave %d with %d pieces, want wave 2 with %d", g.Wave(), len(g.Pieces()), PiecesAtWave(2))
	}
	space := motion.Space{W: testCols * CellW, H: testRows * CellH}
	for _, p := range g.Pieces() {
		if p.Size != Large || space.Near(p.Pos, g.Drone().Pos, SafeDistance-1) {
			t.Fatalf("piece %+v must be large and at least %d from the drone", p, SafeDistance)
		}
	}
}

func TestWavesGrowAndSpeedUpUpToALimit(t *testing.T) {
	if PiecesAtWave(1) != 4 || PiecesAtWave(2) != 5 || PiecesAtWave(50) != MaxPieces {
		t.Fatalf("pieces per wave %d %d %d", PiecesAtWave(1), PiecesAtWave(2), PiecesAtWave(50))
	}
	if !(PieceSpeed(Large, 2) > PieceSpeed(Large, 1) && PieceSpeed(Small, 1) > PieceSpeed(Medium, 1) && PieceSpeed(Medium, 1) > PieceSpeed(Large, 1)) {
		t.Fatalf("later waves and smaller pieces must be faster")
	}
	if PieceSpeed(Small, 100) != PieceSpeed(Small, 30) || PieceSpeed(Small, 100) >= ShotSpeed {
		t.Fatalf("piece speeds must level off below the shots")
	}
}

// TestRealTimeSpeedsDoNotDependOnTheTickRate pins speeds per second, so a
// change of TickPeriod must rescale the per-tick constants with it.
func TestRealTimeSpeedsDoNotDependOnTheTickRate(t *testing.T) {
	perSecond := int(time.Second / TickPeriod)
	cells := func(units int) int { return units * perSecond / CellW }
	if got := cells(MaxDroneSpeed); got != 18 {
		t.Fatalf("drone top speed %d cells/s, want 18", got)
	}
	if got := cells(ShotSpeed); got != 36 {
		t.Fatalf("shot speed %d cells/s, want 36", got)
	}
	if turn := time.Duration(motion.Headings*glide.Unit/TurnSpeed) * TickPeriod; turn < 1200*time.Millisecond || turn > 1400*time.Millisecond {
		t.Fatalf("a full turn takes %v holding the key, want about 1.3s", turn)
	}
	if got := time.Duration(ShieldTicks) * TickPeriod; got.Round(time.Millisecond) != 2*time.Second {
		t.Fatalf("shield %v, want 2s", got)
	}
}

func TestShotsWrapAcrossTheEdgesAndHitWhatIsThere(t *testing.T) {
	g := newTestGame(Piece{Size: Medium, Body: motion.Body{Pos: motion.Vec{X: 1500, Y: 4 * CellH}}}, far(Large))
	g.shots = []Shot{{Body: motion.Body{Pos: motion.Vec{X: testCols*CellW - 500, Y: 4 * CellH}, Vel: motion.Vec{X: ShotSpeed}}, Left: ShotTicks}}
	steps(g, 3, Input{})
	if g.Score() != PointsMedium {
		t.Fatalf("score %d, want the piece across the edge hit (%d)", g.Score(), PointsMedium)
	}
}

func TestShotPositionsStayInsideTheSpace(t *testing.T) {
	g := newTestGame(far(Large))
	g.shots = []Shot{{Body: motion.Body{Pos: motion.Vec{X: testCols*CellW - 100, Y: 100}, Vel: motion.Vec{X: ShotSpeed, Y: -ShotSpeed}}, Left: ShotTicks}}
	g.Step(Input{})
	if p := g.Shots()[0].Pos; p.X < 0 || p.X >= testCols*CellW || p.Y < 0 || p.Y >= testRows*CellH {
		t.Fatalf("shot at %v, want wrapped inside the space", p)
	}
}
