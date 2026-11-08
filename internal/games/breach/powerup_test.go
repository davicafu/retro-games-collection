package breach

import "testing"

// scriptedRand returns values in order, then zeros.
type scriptedRand struct {
	values []int
	idx    int
}

func (r *scriptedRand) Intn(n int) int {
	if r.idx >= len(r.values) {
		return 0
	}
	v := r.values[r.idx] % n
	r.idx++
	return v
}

// breakBrick destroys a one-hit block at (8..11, 3) from below and returns the game.
func breakBrick(rng Rand) *Game {
	g := newGameWith(rng, Brick{X: 8, Y: 3, HP: 1}, farBrick)
	place(g, 9, 4, 0, -120)
	g.Step(Input{})
	return g
}

func TestDestroyedBlockDropsAPowerUpWhenTheRandomRollHits(t *testing.T) {
	g := breakBrick(&scriptedRand{values: []int{0, int(PowerUpLife)}})
	drops := g.Drops()
	if len(drops) != 1 || drops[0].Kind != PowerUpLife || drops[0].Col != 8+BrickWidth/2 {
		t.Fatalf("drops = %+v, want one Life drop under the block centre", drops)
	}
}

func TestNoDropWhenTheRollMissesOrWithoutRand(t *testing.T) {
	if g := breakBrick(&scriptedRand{values: []int{1}}); len(g.Drops()) != 0 {
		t.Fatalf("drops = %+v, want none when the roll misses", g.Drops())
	}
	if g := breakBrick(nil); len(g.Drops()) != 0 {
		t.Fatalf("drops = %+v, want none without a Rand", g.Drops())
	}
}

func TestDamagedButStandingBlockDoesNotDrop(t *testing.T) {
	g := newGameWith(&scriptedRand{}, Brick{X: 8, Y: 3, HP: 2})
	place(g, 9, 4, 0, -120)
	g.Step(Input{})
	if len(g.Drops()) != 0 {
		t.Fatalf("a block that survives the hit must not drop")
	}
}

// catchDrop lets a drop fall onto an injector placed under it.
func catchDrop(t *testing.T, g *Game, d Drop) {
	t.Helper()
	g.drops = []Drop{d}
	g.paddleX = d.Col - g.paddleWidth()/2
	for i := 0; i < 100 && len(g.drops) > 0; i++ {
		g.Step(Input{})
	}
	if len(g.drops) != 0 {
		t.Fatalf("drop never reached the injector")
	}
}

func TestDropsFallAtDropSpeedEvenWhileTheVirusRides(t *testing.T) {
	g := newTestGame(farBrick)
	g.drops = []Drop{{Col: 0, Y: 2*SubCells + SubCells/2, Kind: PowerUpLife}}
	g.Step(Input{})
	if d := g.Drops(); len(d) != 1 || d[0].Y != 2*SubCells+SubCells/2+DropSpeed {
		t.Fatalf("drops = %+v, want Y=%d", d, 2*SubCells+SubCells/2+DropSpeed)
	}
}

func TestCatchingLifeRestoresALostOneUpToTwo(t *testing.T) {
	g := newTestGame(farBrick)
	if g.Lives() != 2 {
		t.Fatalf("starting lives = %d, want 2", g.Lives())
	}
	catchDrop(t, g, Drop{Col: 10, Y: 5 * SubCells, Kind: PowerUpLife})
	if g.Lives() != 2 {
		t.Fatalf("lives = %d, a full virus must stay at 2", g.Lives())
	}
	g.lives = 1
	catchDrop(t, g, Drop{Col: 10, Y: 5 * SubCells, Kind: PowerUpLife})
	if g.Lives() != 2 {
		t.Fatalf("lives = %d, want a lost life restored to 2", g.Lives())
	}
}

func TestCatchingWideWidensTheInjectorAroundItsCentreForAWhile(t *testing.T) {
	g := newTestGame(farBrick)
	catchDrop(t, g, Drop{Col: 10, Y: 5 * SubCells, Kind: PowerUpWide})
	x, w := g.Paddle()
	if w != WidePaddleWidth || x+w/2 != 10 {
		t.Fatalf("injector x=%d width=%d, want width %d centred on 10", x, w, WidePaddleWidth)
	}

	for i := 0; i < WideTicks; i++ {
		g.Step(Input{})
	}
	x, w = g.Paddle()
	if w != PaddleWidth || x+w/2 != 10 {
		t.Fatalf("after %d ticks: injector x=%d width=%d, want width %d centred on 10", WideTicks, x, w, PaddleWidth)
	}
}

func TestMissedDropIsRemovedPastTheBottom(t *testing.T) {
	g := newTestGame(farBrick)
	g.paddleX = 0
	g.drops = []Drop{{Col: testWidth - 1, Y: 5 * SubCells, Kind: PowerUpLife}}
	for i := 0; i < 100; i++ {
		g.Step(Input{})
	}
	if len(g.Drops()) != 0 || g.Lives() != StartLives {
		t.Fatalf("drops=%+v lives=%d, want the missed drop gone and no effect", g.Drops(), g.Lives())
	}
}

func TestCatchingWideAgainRefreshesItWithoutMovingTheInjector(t *testing.T) {
	g := newTestGame(farBrick)
	catchDrop(t, g, Drop{Col: 10, Y: 5 * SubCells, Kind: PowerUpWide})
	for i := 0; i < 50; i++ {
		g.Step(Input{})
	}
	g.drops = []Drop{{Col: 10, Y: paddleRow*SubCells - DropSpeed, Kind: PowerUpWide}}
	g.Step(Input{})
	x, w := g.Paddle()
	if x+w/2 != 10 || g.WideLeft() != WideTicks {
		t.Fatalf("injector centre=%d wideLeft=%d, want centre 10 and a fresh %d", x+w/2, g.WideLeft(), WideTicks)
	}
}
