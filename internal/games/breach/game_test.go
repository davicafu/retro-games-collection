package breach

import "testing"

const (
	testWidth  = 20
	testHeight = 12
	paddleRow  = testHeight - 1
)

func newTestGame(bricks ...Brick) *Game {
	return NewGame(Config{Width: testWidth, Height: testHeight, Bricks: bricks})
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

func TestInjectorMovesByStepAndStopsAtTheWalls(t *testing.T) {
	g := newTestGame(farBrick)
	start, _ := g.Paddle()
	g.Step(Input{Move: 1})
	if x, _ := g.Paddle(); x != start+PaddleStep {
		t.Fatalf("injector x = %d, want %d", x, start+PaddleStep)
	}
	for i := 0; i < 50; i++ {
		g.Step(Input{Move: 1})
	}
	if x, w := g.Paddle(); x != testWidth-w {
		t.Fatalf("injector x = %d, want stopped at %d", x, testWidth-w)
	}
	for i := 0; i < 50; i++ {
		g.Step(Input{Move: -1})
	}
	if x, _ := g.Paddle(); x != 0 {
		t.Fatalf("injector x = %d, want stopped at 0", x)
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
	place(g, testWidth-1, 5, 60, -30)
	g.Step(Input{})
	if vx, _ := g.Velocity(); vx != -60 {
		t.Fatalf("right wall: vx = %d, want -60", vx)
	}
	if col, _ := g.BallCell(); col != testWidth-1 {
		t.Fatalf("right wall: virus left the field at column %d", col)
	}

	place(g, 0, 5, -60, -30)
	g.Step(Input{})
	if vx, _ := g.Velocity(); vx != 60 {
		t.Fatalf("left wall: vx = %d, want 60", vx)
	}

	place(g, 10, 0, 25, -55)
	g.Step(Input{})
	if _, vy := g.Velocity(); vy != 55 {
		t.Fatalf("ceiling: vy = %d, want 55", vy)
	}
}

func TestHittingABlockFromBelowDamagesItAndReflects(t *testing.T) {
	g := newTestGame(Brick{X: 8, Y: 3, HP: 2}, farBrick)
	place(g, 9, 4, 0, -60)
	g.Step(Input{})

	if _, vy := g.Velocity(); vy != 60 {
		t.Fatalf("vy = %d, want reflected to 60", vy)
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
	place(g, 9, 4, 0, -60)
	g.Step(Input{})
	if len(g.Bricks()) != 1 {
		t.Fatalf("blocks = %+v, want the hit one removed", g.Bricks())
	}
}

func TestHittingABlockSidewaysReflectsHorizontally(t *testing.T) {
	g := newTestGame(Brick{X: 10, Y: 5, HP: 1}, farBrick)
	place(g, 9, 5, 60, 0)
	g.Step(Input{})
	if vx, _ := g.Velocity(); vx != -60 {
		t.Fatalf("vx = %d, want -60", vx)
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
		place(g, x+c.offset, paddleRow-1, 0, 60)
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
	place(g, 0, paddleRow, 0, 60) // injector starts centred, far from column 0
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
		place(g, 0, paddleRow, 0, 60)
		g.Step(Input{})
	}
	if !g.Over() || g.Won() || g.Lives() != 0 {
		t.Fatalf("over=%v won=%v lives=%d, want lost with 0 lives", g.Over(), g.Won(), g.Lives())
	}
}

func TestDestroyingEveryBlockBreachesTheFirewall(t *testing.T) {
	g := newTestGame(Brick{X: 8, Y: 3, HP: 1})
	place(g, 9, 4, 0, -60)
	g.Step(Input{})
	if !g.Won() || !g.Over() {
		t.Fatalf("won=%v over=%v, want breached", g.Won(), g.Over())
	}
}

func TestOverGameDoesNotAdvance(t *testing.T) {
	g := newTestGame(Brick{X: 8, Y: 3, HP: 1})
	place(g, 9, 4, 0, -60)
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

func TestDefaultFirewallHasHardenedTopRow(t *testing.T) {
	bricks := Firewall(40)
	if len(bricks) != 5*40/BrickWidth {
		t.Fatalf("blocks = %d, want %d", len(bricks), 5*40/BrickWidth)
	}
	for _, b := range bricks {
		want := 1
		if b.Y == FirewallTop {
			want = 2
		}
		if b.HP != want {
			t.Fatalf("block %+v HP = %d, want %d", b, b.HP, want)
		}
	}
}

func TestRightWallIsTheFirstSubCellPastTheLastColumn(t *testing.T) {
	g := newTestGame(farBrick)
	place(g, testWidth-1, 5, SubCells/2, -30) // lands exactly on width*SubCells
	g.Step(Input{})
	if vx, _ := g.Velocity(); vx != -SubCells/2 {
		t.Fatalf("vx = %d, want reflected", vx)
	}
	if col, _ := g.BallCell(); col != testWidth-1 {
		t.Fatalf("virus at column %d, outside the field", col)
	}
}

// TestAutopilotBreachesTheDefaultFirewall plays full rounds on the real
// playfield: an injector that tracks the virus and varies where it lands
// must clear every block, so the default wall never leaves a block the
// virus cannot reach. Exit angles are pinned by TestInjectorZoneSetsTheExitAngle.
func TestAutopilotBreachesTheDefaultFirewall(t *testing.T) {
	for seed := 0; seed < 10; seed++ {
		rng := &scriptedRand{values: []int{seed, seed + 1, seed + 2}}
		g := NewGame(Config{Width: 40, Height: 16, Bricks: Firewall(40), Rand: rng})
		// Offset the injector so each run starts from a different angle.
		g.paddleX = seed * 3
		g.Step(Input{Launch: true})
		for tick := 0; tick < 20000 && !g.Over(); tick++ {
			// Like a player, vary where the virus lands after every hit.
			aim := g.Score()/PointsPerHit%3 - 1
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
			g.Step(Input{Move: move, Launch: true})
		}
		if !g.Won() {
			t.Fatalf("seed %d: not breached, lives=%d blocks left=%d", seed, g.Lives(), len(g.Bricks()))
		}
		if g.Lives() < StartLives {
			t.Fatalf("seed %d: autopilot lost a life", seed)
		}
	}
}
