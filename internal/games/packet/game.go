// Package packet is the "Metro Survival" game: a network packet races along
// fiber links, switching links to dodge firewalls and taking bandwidth.
package packet

import (
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/scroll"
)

// TickPeriod is the real time one Step stands for; every per-tick constant
// below is tuned to it.
const TickPeriod = time.Second / 60

const (
	// Links is the number of fiber links (lanes), numbered from the top.
	Links = 4

	// Speeds are in scroll.Unit points per tick. MaxSpeed stays below
	// scroll.Unit, so the world scrolls at most one cell per tick and no
	// firewall tunnels through the packet.
	BaseSpeed     = 25 // 15 cells/s
	MaxSpeed      = 96 // 57.6 cells/s
	SpeedStep     = 3
	SpeedDistance = 60 // cells between speed steps

	// PointsToken is what a bandwidth token scores.
	PointsToken = 50
)

var speedRamp = scroll.Ramp{Base: BaseSpeed, Step: SpeedStep, Every: SpeedDistance, Max: MaxSpeed}

// Firewall blocks link Link over columns X..X+Length-1.
type Firewall struct {
	Link   int
	X      int
	Length int
}

func (f Firewall) covers(link, x int) bool {
	return f.Link == link && x >= f.X && x < f.X+f.Length
}

// Token is a bandwidth token on link Link at column X.
type Token struct {
	Link int
	X    int
}

// Input is the player's intent for one tick: Switch is -1 (up a link),
// 1 (down a link) or 0.
type Input struct {
	Switch int
}

// Spawner adds firewalls and tokens at the right edge as the world scrolls.
// It receives what is on the links, the playfield width and the speed the
// world will have when something placed at the right edge reaches the packet.
type Spawner interface {
	Spawn(firewalls []Firewall, tokens []Token, width, speed int) ([]Firewall, []Token)
}

// Config builds a Game. Firewalls and Tokens seed the world; Spawner, when
// set, adds more as it scrolls.
type Config struct {
	Width     int
	PlayerX   int
	Link      int
	Firewalls []Firewall
	Tokens    []Token
	Spawner   Spawner
}

// Game is the headless Packet simulation, advanced one tick at a time.
type Game struct {
	width     int
	playerX   int
	link      int
	firewalls []Firewall
	tokens    []Token
	spawner   Spawner

	scroller scroll.Scroller
	distance int
	taken    int
	dropped  bool
}

func NewGame(cfg Config) *Game {
	return &Game{
		width:     cfg.Width,
		playerX:   cfg.PlayerX,
		link:      cfg.Link,
		firewalls: append([]Firewall(nil), cfg.Firewalls...),
		tokens:    append([]Token(nil), cfg.Tokens...),
		spawner:   cfg.Spawner,
	}
}

// Step advances the simulation by one tick: the switch first, then the
// scroll. A dropped packet does not advance.
func (g *Game) Step(in Input) {
	if g.dropped {
		return
	}
	if in.Switch != 0 {
		g.link = min(max(g.link+in.Switch, 0), Links-1)
		if g.check() {
			return
		}
	}
	g.scroll()
	g.check()
}

// check drops the packet on a firewall or takes the token it is on, and
// reports a drop.
func (g *Game) check() bool {
	for _, f := range g.firewalls {
		if f.covers(g.link, g.playerX) {
			g.dropped = true
			return true
		}
	}
	kept := g.tokens[:0]
	for _, t := range g.tokens {
		if t.Link == g.link && t.X == g.playerX {
			g.taken++
			continue
		}
		kept = append(kept, t)
	}
	g.tokens = kept
	return false
}

func (g *Game) scroll() {
	for range g.scroller.Advance(g.Speed()) {
		g.distance++
		walls := g.firewalls[:0]
		for _, f := range g.firewalls {
			if f.X--; f.X+f.Length > 0 {
				walls = append(walls, f)
			}
		}
		g.firewalls = walls
		tokens := g.tokens[:0]
		for _, t := range g.tokens {
			if t.X--; t.X >= 0 {
				tokens = append(tokens, t)
			}
		}
		g.tokens = tokens
		if g.spawner != nil {
			arrival := speedRamp.At(g.distance + g.width - g.playerX)
			g.firewalls, g.tokens = g.spawner.Spawn(g.firewalls, g.tokens, g.width, arrival)
		}
	}
}

// Speed is the current scroll speed in scroll.Unit points per tick.
func (g *Game) Speed() int { return speedRamp.At(g.distance) }

// Score is the distance travelled in cells plus the tokens taken.
func (g *Game) Score() int { return g.distance + g.taken*PointsToken }

func (g *Game) Distance() int { return g.distance }
func (g *Game) Taken() int    { return g.taken }
func (g *Game) Dropped() bool { return g.dropped }
func (g *Game) Width() int    { return g.width }
func (g *Game) PlayerX() int  { return g.playerX }
func (g *Game) Link() int     { return g.link }
func (g *Game) Firewalls() []Firewall {
	return append([]Firewall(nil), g.firewalls...)
}
func (g *Game) Tokens() []Token { return append([]Token(nil), g.tokens...) }
