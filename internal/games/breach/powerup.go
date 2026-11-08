package breach

// Rand is the randomness power-ups need; *math/rand.Rand satisfies it.
type Rand interface {
	Intn(n int) int
}

// PowerUp is what a falling drop grants when the injector catches it.
type PowerUp int

const (
	PowerUpWide PowerUp = iota
	PowerUpLife
	powerUpCount
)

const (
	// DropOdds: one destroyed block in DropOdds drops a power-up.
	DropOdds = 5
	// DropSpeed is how fast drops fall, in sub-cells per tick.
	DropSpeed = 30

	WidePaddleWidth = 11
	// WideTicks is how long Wide lasts: 15 s at 50 ms per tick.
	WideTicks = 300
)

// Drop is a power-up falling from a destroyed block; Y is in sub-cells.
type Drop struct {
	Col  int
	Y    int
	Kind PowerUp
}

// maybeDrop rolls for a power-up under a destroyed block.
func (g *Game) maybeDrop(b Brick) {
	if g.rng == nil || g.rng.Intn(DropOdds) != 0 {
		return
	}
	kind := PowerUp(g.rng.Intn(int(powerUpCount)))
	g.drops = append(g.drops, Drop{Col: b.X + BrickWidth/2, Y: b.Y*SubCells + SubCells/2, Kind: kind})
}

// fallDrops moves every drop down; the injector catches the ones landing on it.
func (g *Game) fallDrops() {
	kept := g.drops[:0]
	for _, d := range g.drops {
		d.Y += DropSpeed
		row := d.Y / SubCells
		switch {
		case row == g.paddleRow() && g.onPaddle(d.Col):
			g.apply(d.Kind)
		case row < g.height:
			kept = append(kept, d)
		}
	}
	g.drops = kept
}

func (g *Game) apply(p PowerUp) {
	switch p {
	case PowerUpLife:
		g.lives = min(g.lives+1, MaxLives)
	case PowerUpWide:
		if g.wideTicks == 0 {
			g.paddleX -= (WidePaddleWidth - PaddleWidth) / 2
		}
		g.wideTicks = WideTicks
		g.clampPaddle()
	}
}

// tickWide counts Wide down and narrows the injector around its centre when it ends.
func (g *Game) tickWide() {
	if g.wideTicks == 0 {
		return
	}
	g.wideTicks--
	if g.wideTicks == 0 {
		g.paddleX += (WidePaddleWidth - PaddleWidth) / 2
		g.clampPaddle()
	}
}
