package runner

import "github.com/davidbarreiro/cli-games-collection/internal/engine/scroll"

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

	// Speed is measured in cells per scroll.Unit ticks. MaxSpeed must stay below
	// scroll.Unit so the world scrolls at most one cell per tick and nothing
	// tunnels through the player.
	BaseSpeed     = 35
	SpeedStep     = 5
	SpeedDistance = 25
	MaxSpeed      = 99

	scrollUnit = scroll.Unit
)

var speedRamp = scroll.Ramp{Base: BaseSpeed, Step: SpeedStep, Every: SpeedDistance, Max: MaxSpeed}

const (
	// JumpHeight is how high a jump from the ground rises, in cells.
	JumpHeight = 4
	// DoubleJumpLift is how much higher a second jump in mid-air rises
	// from where the player is, up to MaxJumpHeight.
	DoubleJumpLift = 3
	MaxJumpHeight  = 6
)

// jumpArc is a jump's heights, one per tick, from height from: it rises a
// cell per tick to peak, holds it two more ticks, then falls a cell per
// tick to the ground.
func jumpArc(from, peak int) []int {
	var arc []int
	for h := from + 1; h <= peak; h++ {
		arc = append(arc, h)
	}
	arc = append(arc, peak, peak)
	for h := peak - 1; h > 0; h-- {
		arc = append(arc, h)
	}
	return arc
}

// jumpHeights is the arc of a jump from the ground.
var jumpHeights = jumpArc(0, JumpHeight)

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

	arc      []int // the current jump's heights; nil when grounded
	arcIndex int
	doubled  bool // the second jump of this airtime is spent
	duckLeft int

	scroller scroll.Scroller
	distance int
	dead     bool
}

func NewGame(cfg Config) *Game {
	return &Game{
		width:     cfg.Width,
		playerX:   cfg.PlayerX,
		obstacles: append([]Obstacle(nil), cfg.Obstacles...),
		spawner:   cfg.Spawner,
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
	case g.airborne() && action == ActionJump && !g.doubled:
		h := g.PlayerHeight()
		g.arc, g.arcIndex, g.doubled = jumpArc(h, min(h+DoubleJumpLift, MaxJumpHeight)), 0, true
		return
	case g.airborne():
		if g.arcIndex++; g.arcIndex >= len(g.arc) {
			g.arc, g.doubled = nil, false
		}
	case action == ActionJump:
		g.arc, g.arcIndex = jumpHeights, 0
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
	for range g.scroller.Advance(g.Speed()) {
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

func (g *Game) airborne() bool { return g.arc != nil }

// PlayerHeight is the height of the player's feet above the ground.
func (g *Game) PlayerHeight() int {
	if !g.airborne() {
		return 0
	}
	return g.arc[g.arcIndex]
}

func (g *Game) Ducking() bool { return g.duckLeft > 0 }
func (g *Game) Dead() bool    { return g.dead }
func (g *Game) Width() int    { return g.width }
func (g *Game) PlayerX() int  { return g.playerX }

// Score is the distance travelled, in cells.
func (g *Game) Score() int { return g.distance }

// Speed is the current scroll speed in cells per 100 ticks.
func (g *Game) Speed() int {
	return speedRamp.At(g.distance)
}

func (g *Game) Obstacles() []Obstacle {
	return append([]Obstacle(nil), g.obstacles...)
}
