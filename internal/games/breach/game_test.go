package breach

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

const (
	testWidth  = 20
	testHeight = 12
	paddleRow  = testHeight - 1
)

// testLevel is a single level at base speed; tests place their own blocks.
var testLevel = Level{Speed: 100, DropOdds: 5}

func newTestGame(bricks ...Brick) *Game {
	return newGameWith(nil, bricks...)
}

func newGameWith(rng Rand, bricks ...Brick) *Game {
	g := NewGame(Config{Width: testWidth, Height: testHeight, Levels: []Level{testLevel}, Rand: rng})
	g.bricks = append([]Brick(nil), bricks...)
	return g
}

// farBrick keeps a round alive without being in the way of most tests.
var farBrick = Brick{X: 0, Y: 0, HP: 1}

// place puts the virus at a cell centre with the given velocity, already launched.
func place(g *Game, col, row, vx, vy int) {
	g.launched = true
	g.bx, g.by = col*SubCells+SubCells/2, row*SubCells+SubCells/2
	g.vx, g.vy = vx, vy
}

func TestVirusRidesTheInjectorUntilLaunch(t *testing.T) {
	g := newTestGame(farBrick)
	for i := 0; i < 3; i++ {
		g.Step(Input{Move: 1})
	}
	x, w := g.Paddle()
	col, row := g.BallCell()
	if g.Launched() || col != x+w/2 || row != paddleRow-1 {
		t.Fatalf("virus at (%d,%d), want riding the injector at (%d,%d)", col, row, x+w/2, paddleRow-1)
	}
}

// glideOffsets presses right once, then records the injector's offset from
// its start after that tick and each of the following idle ticks.
func glideOffsets(g *Game, ticks int) []int {
	start, _ := g.Paddle()
	g.Step(Input{Move: 1})
	x, _ := g.Paddle()
	got := []int{x - start}
	for i := 1; i < ticks; i++ {
		g.Step(Input{})
		x, _ := g.Paddle()
		got = append(got, x-start)
	}
	return got
}

func TestInjectorMovesOnTheFirstTickThenGlidesAndStops(t *testing.T) {
	got := glideOffsets(newTestGame(farBrick), PaddleGlide+3)
	if got[0] != 1 {
		t.Fatalf("offset after the press tick = %d, want 1: the first tick must respond", got[0])
	}
	for i := 1; i < len(got); i++ {
		if step := got[i] - got[i-1]; step < 0 || step > 1 {
			t.Fatalf("offsets %v: the injector must glide at most one cell per tick", got)
		}
	}
	want := (paddleUnit - PaddleSpeed + PaddleGlide*PaddleSpeed) / paddleUnit
	if last := got[len(got)-1]; last != want || got[PaddleGlide-1] != want {
		t.Fatalf("offsets %v: a tap must glide %d cells and stop after %d ticks", got, want, PaddleGlide)
	}
}

func TestHoldingAMoveKeyGlidesAtPaddleSpeedAndStopsAtTheWalls(t *testing.T) {
	g := newTestGame(farBrick)
	start, _ := g.Paddle()
	for i := 0; i < 9; i++ {
		g.Step(Input{Move: 1})
	}
	if x, _ := g.Paddle(); x-start != (paddleUnit-PaddleSpeed+9*PaddleSpeed)/paddleUnit {
		t.Fatalf("moved %d cells in 9 held ticks, want %d", x-start, (paddleUnit-PaddleSpeed+9*PaddleSpeed)/paddleUnit)
	}
	for i := 0; i < 100; i++ {
		g.Step(Input{Move: 1})
	}
	if x, w := g.Paddle(); x != testWidth-w {
		t.Fatalf("injector x = %d, want stopped at %d", x, testWidth-w)
	}
	for i := 0; i < 100; i++ {
		g.Step(Input{Move: -1})
	}
	if x, _ := g.Paddle(); x != 0 {
		t.Fatalf("injector x = %d, want stopped at 0", x)
	}
}

func TestOppositePressReversesTheGlideAtOnce(t *testing.T) {
	g := newTestGame(farBrick)
	start, _ := g.Paddle()
	g.Step(Input{Move: 1})
	g.Step(Input{Move: -1})
	if x, _ := g.Paddle(); x != start {
		t.Fatalf("injector x = %d, want back at %d after reversing", x, start)
	}
}

// TestRealTimeSpeedsDoNotDependOnTheTickRate pins speeds per second, so a
// change of TickPeriod must rescale the per-tick constants with it.
func TestRealTimeSpeedsDoNotDependOnTheTickRate(t *testing.T) {
	if TickPeriod != time.Second/60 {
		t.Fatalf("tick period = %v, want 60 ticks per second", TickPeriod)
	}
	perSecond := int(time.Second / TickPeriod)
	if got := Diagonal.X * perSecond / SubCells; got != 9 {
		t.Fatalf("diagonal virus moves %d cells/s, want 9", got)
	}
	if got := DropSpeed * perSecond / SubCells; got != 6 {
		t.Fatalf("drops fall %d cells/s, want 6", got)
	}
	if got := PaddleSpeed * perSecond / paddleUnit; got != 40 {
		t.Fatalf("injector glides %d cells/s, want 40", got)
	}
	if got := time.Duration(WideTicks) * TickPeriod; got.Round(time.Millisecond) != 15*time.Second {
		t.Fatalf("Wide lasts %v, want 15s", got)
	}
}

func TestLaunchSendsTheVirusUpDiagonally(t *testing.T) {
	g := newTestGame(farBrick)
	g.Step(Input{Launch: true})
	if vx, vy := g.Velocity(); !g.Launched() || vx != Diagonal.X || vy != -Diagonal.Y {
		t.Fatalf("velocity = (%d,%d), want (%d,%d)", vx, vy, Diagonal.X, -Diagonal.Y)
	}
}

func TestWallsAndCeilingReflect(t *testing.T) {
	g := newTestGame(farBrick)
	place(g, testWidth-1, 5, 120, -60)
	g.Step(Input{})
	if vx, _ := g.Velocity(); vx != -120 {
		t.Fatalf("right wall: vx = %d, want -120", vx)
	}
	if col, _ := g.BallCell(); col != testWidth-1 {
		t.Fatalf("right wall: virus left the field at column %d", col)
	}

	place(g, 0, 5, -120, -60)
	g.Step(Input{})
	if vx, _ := g.Velocity(); vx != 120 {
		t.Fatalf("left wall: vx = %d, want 120", vx)
	}

	place(g, 10, 0, 50, -110)
	g.Step(Input{})
	if _, vy := g.Velocity(); vy != 110 {
		t.Fatalf("ceiling: vy = %d, want 110", vy)
	}
}

func TestHittingABlockFromBelowDamagesItAndReflects(t *testing.T) {
	g := newTestGame(Brick{X: 8, Y: 3, HP: 2}, farBrick)
	place(g, 9, 4, 0, -120)
	g.Step(Input{})

	if _, vy := g.Velocity(); vy != 120 {
		t.Fatalf("vy = %d, want reflected to 120", vy)
	}
	if b := g.Bricks()[0]; b.HP != 1 {
		t.Fatalf("block HP = %d, want 1", b.HP)
	}
	if g.Score() != PointsPerHit {
		t.Fatalf("score = %d, want %d", g.Score(), PointsPerHit)
	}
	if _, row := g.BallCell(); row != 4 {
		t.Fatalf("virus entered the block row: row %d", row)
	}
}

func TestDestroyedBlockIsRemoved(t *testing.T) {
	g := newTestGame(Brick{X: 8, Y: 3, HP: 1}, farBrick)
	place(g, 9, 4, 0, -120)
	g.Step(Input{})
	if len(g.Bricks()) != 1 {
		t.Fatalf("blocks = %+v, want the hit one removed", g.Bricks())
	}
}

func TestHittingABlockSidewaysReflectsHorizontally(t *testing.T) {
	g := newTestGame(Brick{X: 10, Y: 5, HP: 1}, farBrick)
	place(g, 9, 5, 120, 0)
	g.Step(Input{})
	if vx, _ := g.Velocity(); vx != -120 {
		t.Fatalf("vx = %d, want -120", vx)
	}
}

func TestInjectorZoneSetsTheExitAngle(t *testing.T) {
	// Injector of width 7: offsets 0..6 from its left edge.
	cases := []struct {
		offset, incomingVX int
		want               Velocity
	}{
		{0, 25, Velocity{-Shallow.X, -Shallow.Y}},
		{1, 25, Velocity{-Diagonal.X, -Diagonal.Y}},
		{2, 25, Velocity{-Steep.X, -Steep.Y}},
		{3, 25, Velocity{Steep.X, -Steep.Y}},   // centre keeps going right
		{3, -25, Velocity{-Steep.X, -Steep.Y}}, // centre keeps going left
		{4, -25, Velocity{Steep.X, -Steep.Y}},
		{5, -25, Velocity{Diagonal.X, -Diagonal.Y}},
		{6, -25, Velocity{Shallow.X, -Shallow.Y}},
	}
	for _, c := range cases {
		g := newTestGame(farBrick)
		x, w := g.Paddle()
		if w != PaddleWidth {
			t.Fatalf("injector width = %d, want %d", w, PaddleWidth)
		}
		place(g, x+c.offset, paddleRow-1, 0, 120)
		g.vx = c.incomingVX
		g.bx = (x+c.offset)*SubCells + SubCells/2 - c.incomingVX // lands on the same column after moving
		g.Step(Input{})
		if vx, vy := g.Velocity(); vx != c.want.X || vy != c.want.Y {
			t.Fatalf("offset %d, incoming vx %d: velocity (%d,%d), want (%d,%d)", c.offset, c.incomingVX, vx, vy, c.want.X, c.want.Y)
		}
	}
}

func TestMissingCostsALifeAndResetsTheVirus(t *testing.T) {
	g := newTestGame(farBrick)
	place(g, 0, paddleRow, 0, 120) // injector starts centred, far from column 0
	g.Step(Input{})

	if g.Lives() != StartLives-1 {
		t.Fatalf("lives = %d, want %d", g.Lives(), StartLives-1)
	}
	if g.Launched() || g.Over() {
		t.Fatalf("after a miss the virus must ride the injector again")
	}
}

func TestLosingTheLastLifeEndsTheRound(t *testing.T) {
	g := newTestGame(farBrick)
	for i := 0; i < StartLives; i++ {
		place(g, 0, paddleRow, 0, 120)
		g.Step(Input{})
	}
	if !g.Over() || g.Won() || g.Lives() != 0 {
		t.Fatalf("over=%v won=%v lives=%d, want lost with 0 lives", g.Over(), g.Won(), g.Lives())
	}
}

func TestDestroyingEveryBlockBreachesTheFirewall(t *testing.T) {
	g := newTestGame(Brick{X: 8, Y: 3, HP: 1})
	place(g, 9, 4, 0, -120)
	g.Step(Input{})
	if !g.Won() || !g.Over() {
		t.Fatalf("won=%v over=%v, want breached", g.Won(), g.Over())
	}
}

func TestOverGameDoesNotAdvance(t *testing.T) {
	g := newTestGame(Brick{X: 8, Y: 3, HP: 1})
	place(g, 9, 4, 0, -120)
	g.Step(Input{})
	bx, by := g.Ball()
	x, _ := g.Paddle()
	g.Step(Input{Move: 1})
	if nbx, nby := g.Ball(); nbx != bx || nby != by {
		t.Fatalf("virus moved after the round ended")
	}
	if nx, _ := g.Paddle(); nx != x {
		t.Fatalf("injector moved after the round ended")
	}
}

func TestEveryAngleStaysBelowOneCellPerTick(t *testing.T) {
	for _, v := range []Velocity{Steep, Diagonal, Shallow} {
		if v.X >= SubCells || v.Y >= SubCells || v.X <= 0 || v.Y <= 0 {
			t.Fatalf("angle %+v must have both components in 1..%d so the virus cannot tunnel", v, SubCells-1)
		}
	}
}

func TestRightWallIsTheFirstSubCellPastTheLastColumn(t *testing.T) {
	g := newTestGame(farBrick)
	place(g, testWidth-1, 5, SubCells/2, -60) // lands exactly on width*SubCells
	g.Step(Input{})
	if vx, _ := g.Velocity(); vx != -SubCells/2 {
		t.Fatalf("vx = %d, want reflected", vx)
	}
	if col, _ := g.BallCell(); col != testWidth-1 {
		t.Fatalf("virus at column %d, outside the field", col)
	}
}

// TestAutopilotClearsEveryLevel plays each campaign level on the real
// playfield: an injector that tracks the virus and picks a random landing
// point each time the virus starts falling must clear every destructible
// block, so no wall leaves a block the virus cannot reach. Lives are topped
// up and the injector is moved one cell per tick directly, bypassing the
// glide: this checks reachability, not the autopilot's skill. Exit angles
// are pinned by TestInjectorZoneSetsTheExitAngle.
func TestAutopilotClearsEveryLevel(t *testing.T) {
	for i, level := range Campaign {
		for seed := int64(0); seed < 3; seed++ {
			aimRand := rand.New(rand.NewSource(seed))
			g := NewGame(Config{Width: 40, Height: 16, Levels: []Level{level}, Rand: rand.New(rand.NewSource(seed))})
			g.Step(Input{Launch: true})
			aim, prevVY := 0, 0
			for tick := 0; tick < 150000 && !g.Over(); tick++ {
				g.lives = StartLives
				if _, vy := g.Velocity(); vy > 0 && prevVY <= 0 {
					aim = aimRand.Intn(5) - 2
				}
				_, prevVY = g.Velocity()
				col, _ := g.BallCell()
				x, w := g.Paddle()
				target := x + w/2 + aim
				move := 0
				switch {
				case col < target-1:
					move = -1
				case col > target+1:
					move = 1
				}
				g.paddleX += move
				g.clampPaddle()
				g.Step(Input{Launch: true})
			}
			if !g.Won() {
				t.Fatalf("level %d seed %d: not cleared, blocks left %+v", i+1, seed, g.Bricks())
			}
		}
	}
}

// TestVirusBesideTheInjectorIsNotCaughtByMovingUnderIt pins the rule that
// only a virus falling into the injector row from above bounces: one that
// already slipped past its edge keeps falling, instead of freezing in place
// while the injector slides under it.
func TestVirusBesideTheInjectorIsNotCaughtByMovingUnderIt(t *testing.T) {
	g := newTestGame(farBrick)
	x, w := g.Paddle()
	place(g, x+w, paddleRow, 0, 60)
	g.paddleX++ // the injector now covers the virus column
	for i := 0; i < 10 && g.Lives() == StartLives; i++ {
		g.Step(Input{})
	}
	if g.Lives() != StartLives-1 {
		t.Fatalf("lives = %d, virus at %v: a virus past the injector must be lost, not caught", g.Lives(), fmt.Sprint(g.BallCell()))
	}
}
