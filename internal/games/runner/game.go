package runner

// Action is the player's intent for one tick.
type Action int

const (
	ActionNone Action = iota
	ActionJump
	ActionDuck
)

// Kind distinguishes obstacles the player must jump over from those it must duck under.
type Kind int

const (
	KindGround Kind = iota
	KindFlying
)

// Obstacle occupies columns X..X+Width-1. Ground obstacles fill heights 0..Tall-1;
// flying obstacles fill height 1, the standing player's head.
type Obstacle struct {
	X     int
	Width int
	Tall  int
	Kind  Kind
}

func (o Obstacle) occupies(height int) bool {
	if o.Kind == KindFlying {
		return height == flyingHeight
	}
	return height >= 0 && height < o.Tall
}

const (
	flyingHeight = 1

	// DuckTicks is how long one duck press lasts; terminals report no key release.
	DuckTicks = 12

	// Speed is measured in cells per 100 ticks. MaxSpeed must stay below 100 so the
	// world scrolls at most one cell per tick and nothing tunnels through the player.
	BaseSpeed     = 35
	SpeedStep     = 5
	SpeedDistance = 20
	MaxSpeed      = 90

	scrollUnit = 100
)

// jumpHeights is the scripted jump arc, one entry per tick, in cells above the ground.
var jumpHeights = []int{1, 2, 3, 4, 4, 4, 3, 2, 1}

// Config builds a Game. Obstacles seeds the initial world; Spawner, when set,
// generates new obstacles as the world scrolls.
type Config struct {
	Width     int
	PlayerX   int
	Obstacles []Obstacle
	Spawner   Spawner
}

// Game is the headless Endless Runner simulation, advanced one tick at a time.
type Game struct {
	width     int
	playerX   int
	obstacles []Obstacle
	spawner   Spawner

	jumpIndex int // index into jumpHeights; -1 when grounded
	duckLeft  int

	accumulator int
	distance    int
	dead        bool
}

func NewGame(cfg Config) *Game {
	return &Game{
		width:     cfg.Width,
		playerX:   cfg.PlayerX,
		obstacles: append([]Obstacle(nil), cfg.Obstacles...),
		spawner:   cfg.Spawner,
		jumpIndex: -1,
	}
}

// Step advances the simulation by one tick. A dead game does not advance.
func (g *Game) Step(action Action) {
	if g.dead {
		return
	}
	g.advancePlayer(action)
	g.scroll()
	if g.collides() {
		g.dead = true
	}
}

func (g *Game) advancePlayer(action Action) {
	switch {
	case g.airborne():
		g.jumpIndex++
		if g.jumpIndex >= len(jumpHeights) {
			g.jumpIndex = -1
		}
	case action == ActionJump:
		g.jumpIndex = 0
		g.duckLeft = 0
		return
	}

	if g.duckLeft > 0 {
		g.duckLeft--
	}
	if action == ActionDuck && !g.airborne() {
		g.duckLeft = DuckTicks
	}
}

func (g *Game) scroll() {
	g.accumulator += g.Speed()
	for g.accumulator >= scrollUnit {
		g.accumulator -= scrollUnit
		g.distance++
		kept := g.obstacles[:0]
		for _, o := range g.obstacles {
			o.X--
			if o.X+o.Width > 0 {
				kept = append(kept, o)
			}
		}
		g.obstacles = kept
		if g.spawner != nil {
			g.obstacles = g.spawner.Spawn(g.obstacles, g.width, g.Speed())
		}
	}
}

func (g *Game) collides() bool {
	low, high := g.bodyHeights()
	for _, o := range g.obstacles {
		if g.playerX < o.X || g.playerX >= o.X+o.Width {
			continue
		}
		for h := low; h <= high; h++ {
			if o.occupies(h) {
				return true
			}
		}
	}
	return false
}

// bodyHeights returns the lowest and highest cells the player occupies.
func (g *Game) bodyHeights() (int, int) {
	if g.Ducking() {
		return 0, 0
	}
	h := g.PlayerHeight()
	return h, h + 1
}

func (g *Game) airborne() bool { return g.jumpIndex >= 0 }

// PlayerHeight is the height of the player's feet above the ground.
func (g *Game) PlayerHeight() int {
	if !g.airborne() {
		return 0
	}
	return jumpHeights[g.jumpIndex]
}

func (g *Game) Ducking() bool { return g.duckLeft > 0 }
func (g *Game) Dead() bool    { return g.dead }
func (g *Game) Width() int    { return g.width }
func (g *Game) PlayerX() int  { return g.playerX }

// Score is the distance travelled, in cells.
func (g *Game) Score() int { return g.distance }

// Speed is the current scroll speed in cells per 100 ticks.
func (g *Game) Speed() int {
	speed := BaseSpeed + SpeedStep*(g.distance/SpeedDistance)
	if speed > MaxSpeed {
		return MaxSpeed
	}
	return speed
}

func (g *Game) Obstacles() []Obstacle {
	return append([]Obstacle(nil), g.obstacles...)
}
