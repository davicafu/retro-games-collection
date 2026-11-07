// Package drone is the Flappy-inspired game: a drone flies through the gaps
// between antenna masts that grow from the ceiling and the ground.
package drone

import "github.com/davidbarreiro/cli-games-collection/internal/engine/scroll"

// Physics uses integer sub-cells so the simulation stays deterministic.
// Positions grow downwards: row 0 is the top of the sky.
const (
	SubCells = 100 // sub-cells per cell

	Gravity = 5  // sub-cells per tick added to the velocity every tick
	Thrust  = 45 // upward velocity a thrust sets, in sub-cells per tick
	MaxFall = 60 // terminal fall velocity, in sub-cells per tick
)

// Speed is measured in cells per scroll.Unit ticks and must stay below
// scroll.Unit so masts move at most one cell per tick.
const (
	BaseSpeed     = 40
	SpeedStep     = 5
	SpeedDistance = 30
	MaxSpeed      = 70
)

var speedRamp = scroll.Ramp{Base: BaseSpeed, Step: SpeedStep, Every: SpeedDistance, Max: MaxSpeed}

const (
	AntennaWidth = 2
	GapSize      = 5
)

// Antenna is a pair of masts covering columns X..X+AntennaWidth-1, open only
// between rows GapTop and GapTop+GapSize-1.
type Antenna struct {
	X      int
	GapTop int
}

func (a Antenna) blocks(row int) bool {
	return row < a.GapTop || row >= a.GapTop+GapSize
}

func (a Antenna) covers(x int) bool {
	return x >= a.X && x < a.X+AntennaWidth
}

// Config builds a Game. Height is the number of sky rows; the ground lies
// just below them. Antennas seeds the world; Spawner, when set, adds new ones.
type Config struct {
	Width    int
	Height   int
	DroneX   int
	Antennas []Antenna
	Spawner  Spawner
}

// Game is the headless Drone simulation, advanced one tick at a time.
type Game struct {
	width, height int
	droneX        int
	antennas      []Antenna
	spawner       Spawner

	position int // sub-cells from the top
	velocity int // sub-cells per tick, positive is down
	started  bool

	scroller scroll.Scroller
	distance int
	score    int
	dead     bool
}

func NewGame(cfg Config) *Game {
	return &Game{
		width:    cfg.Width,
		height:   cfg.Height,
		droneX:   cfg.DroneX,
		antennas: append([]Antenna(nil), cfg.Antennas...),
		spawner:  cfg.Spawner,
		position: cfg.Height / 2 * SubCells,
	}
}

// Step advances the simulation by one tick. Before the first thrust the drone
// hovers and the world waits; a dead game does not advance.
func (g *Game) Step(thrust bool) {
	if g.dead || (!g.started && !thrust) {
		return
	}
	g.started = true

	g.fly(thrust)
	if g.Row() >= g.height {
		g.dead = true
		return
	}
	g.scroll()
	if g.collides() {
		g.dead = true
	}
}

func (g *Game) fly(thrust bool) {
	if thrust {
		g.velocity = -Thrust
	} else {
		g.velocity = min(g.velocity+Gravity, MaxFall)
	}
	g.position += g.velocity
	if g.position < 0 {
		g.position = 0
		g.velocity = 0
	}
}

func (g *Game) scroll() {
	for range g.scroller.Advance(g.Speed()) {
		g.distance++
		kept := g.antennas[:0]
		for _, a := range g.antennas {
			a.X--
			if a.X+AntennaWidth == g.droneX {
				g.score++
			}
			if a.X+AntennaWidth > 0 {
				kept = append(kept, a)
			}
		}
		g.antennas = kept
		if g.spawner != nil {
			g.antennas = g.spawner.Spawn(g.antennas, g.width, g.height)
		}
	}
}

func (g *Game) collides() bool {
	for _, a := range g.antennas {
		if a.covers(g.droneX) && a.blocks(g.Row()) {
			return true
		}
	}
	return false
}

// Position is the drone's height in sub-cells from the top of the sky.
func (g *Game) Position() int { return g.position }

// Row is the sky row the drone occupies.
func (g *Game) Row() int { return g.position / SubCells }

func (g *Game) Velocity() int { return g.velocity }
func (g *Game) Started() bool { return g.started }
func (g *Game) Dead() bool    { return g.dead }
func (g *Game) Width() int    { return g.width }
func (g *Game) Height() int   { return g.height }
func (g *Game) DroneX() int   { return g.droneX }
func (g *Game) Score() int    { return g.score }
func (g *Game) Speed() int    { return speedRamp.At(g.distance) }
func (g *Game) Distance() int { return g.distance }

func (g *Game) Antennas() []Antenna {
	return append([]Antenna(nil), g.antennas...)
}
