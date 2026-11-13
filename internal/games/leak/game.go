// Package leak is the Pac-Man-inspired game: a memory leak eats the free
// bytes of the RAM while garbage collectors chase it through the map.
package leak

import (
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// TickPeriod is the real time one Step stands for; every per-tick constant
// below is tuned to it.
const TickPeriod = time.Second / 60

// CellUnits is the progress an actor needs to cross one cell. Speeds are
// in units per tick and stay below it, so nobody skips a cell.
const CellUnits = 120

const (
	LeakSpeed = 16 // 8 cells/s
	FleeSpeed = 8  // 4 cells/s, collectors during a malloc

	StartLives   = 3
	PointsByte   = 10
	PointsMalloc = 50
	// PointsCollector is the first collector eaten in one malloc; each
	// further one doubles it.
	PointsCollector = 100

	// RespawnTicks is how long an eaten collector waits in the pool: 3 s.
	RespawnTicks = 180
	// SightCells is how close the leak must be for the Wanderer to chase it.
	SightCells = 4
	// LeadCells is how far ahead of the leak the Interceptor aims.
	LeadCells = 4
)

// Directions an actor can move in; the zero Point means standing still.
var (
	Up    = grid.Point{X: 0, Y: -1}
	Down  = grid.Point{X: 0, Y: 1}
	Left  = grid.Point{X: -1, Y: 0}
	Right = grid.Point{X: 1, Y: 0}
)

// Rand is the randomness the collectors need; *math/rand.Rand satisfies it.
type Rand interface {
	Intn(n int) int
}

// Input is the player's intent for one tick: Dir is a direction or zero.
type Input struct {
	Dir grid.Point
}

// Actor is a moving piece: it left Cell and is Progress units of CellUnits
// on its way to Cell+Dir.
type Actor struct {
	Cell     grid.Point
	Dir      grid.Point
	Progress int
}

// Occupied is the cell the actor counts as being in: the nearer of the two.
func (a Actor) Occupied() grid.Point {
	if a.Progress >= CellUnits/2 {
		return a.Cell.Add(a.Dir)
	}
	return a.Cell
}

// reverse turns the actor back without a jump: it now heads to the cell it
// left, from the cell it was heading to.
func (a *Actor) reverse() {
	if a.Progress > 0 {
		a.Cell = a.Cell.Add(a.Dir)
		a.Progress = CellUnits - a.Progress
	}
	a.Dir = neg(a.Dir)
}

// advance moves the actor speed units. At each cell centre, choose picks the
// next direction; zero stops it there.
func (a *Actor) advance(speed int, choose func() grid.Point) {
	if a.Progress == 0 {
		a.Dir = choose()
		if a.Dir == (grid.Point{}) {
			return
		}
	}
	a.Progress += speed
	if a.Progress < CellUnits {
		return
	}
	a.Progress -= CellUnits
	a.Cell = a.Cell.Add(a.Dir)
	if a.Dir = choose(); a.Dir == (grid.Point{}) {
		a.Progress = 0
	}
}

// Config builds a Game from its maps (see Maps): level n plays map
// (n-1) mod len(Maps).
type Config struct {
	Maps [][]string
	Rand Rand
}

// Game is the headless Leak simulation, advanced one tick at a time.
type Game struct {
	maps    [][]string
	layout  layout
	width   int
	height  int
	rng     Rand
	exit    grid.Point // the corridor cell just outside the pool gate
	toExit  grid.DistanceField
	corners [kindCount]grid.DistanceField // where each kind pauses
	started bool

	leak       Actor
	wanted     grid.Point
	collectors []Collector
	lifeTicks  int // ticks since the leak's first move in this life
	fright     int // ticks of malloc left
	chain      int // collectors eaten during this malloc

	level     int
	score     int
	lives     int
	bytesLeft int
	lost      bool
}

func NewGame(cfg Config) *Game {
	g := &Game{maps: cfg.Maps, rng: cfg.Rand, lives: StartLives, level: 1}
	g.loadMap()
	return g
}

// loadMap loads the level's map, full of bytes and mallocs, and puts
// everyone at the start.
func (g *Game) loadMap() {
	g.layout = parseMap(g.maps[(g.level-1)%len(g.maps)])
	g.height, g.width = len(g.layout.tiles), len(g.layout.tiles[0])
	g.exit = g.layout.gate.Add(Up)
	g.toExit = grid.Distances(g.width, g.height, g.exit, g.collectorCanEnter(stateLeaving))
	for k := range g.corners {
		g.corners[k] = grid.Distances(g.width, g.height, g.cornerOf(Kind(k)), g.collectorCanEnter(stateActive))
	}
	g.bytesLeft = 0
	for _, row := range g.layout.tiles {
		for _, t := range row {
			if t == TileByte || t == TileMalloc {
				g.bytesLeft++
			}
		}
	}
	g.resetActors()
}

// resetActors puts the leak and the collectors back at their starts and
// waits for the player's next move.
func (g *Game) resetActors() {
	g.started = false
	g.lifeTicks, g.fright, g.chain = 0, 0, 0
	g.leak = Actor{Cell: g.layout.leakStart}
	g.wanted = grid.Point{}
	g.collectors = g.collectors[:0]
	for i, p := range g.layout.poolStart {
		g.collectors = append(g.collectors, Collector{
			Kind:      Kind(i) % kindCount,
			Actor:     Actor{Cell: p},
			home:      p,
			releaseAt: i * g.releaseInterval(),
		})
	}
}

// Step advances the simulation by one tick. A finished round does not
// advance, and nothing moves before the player's first move of a life.
func (g *Game) Step(in Input) {
	if g.lost {
		return
	}
	if in.Dir != (grid.Point{}) {
		g.wanted = in.Dir
		g.started = true
		if g.leak.Dir == neg(in.Dir) {
			g.leak.reverse()
		}
	}
	if !g.started {
		return
	}
	g.lifeTicks++
	if g.fright > 0 {
		g.fright--
		if g.fright == 0 {
			for i := range g.collectors {
				g.collectors[i].fleeing = false
			}
		}
	}

	leakBefore := g.leak.Occupied()
	g.leak.advance(LeakSpeed, g.leakChoice)
	g.eat(g.leak.Occupied())
	if g.bytesLeft == 0 {
		g.nextLevel()
		return
	}

	before := make([]grid.Point, len(g.collectors))
	for i := range g.collectors {
		before[i] = g.collectors[i].Occupied()
		g.moveCollector(&g.collectors[i])
	}
	g.resolveContacts(leakBefore, before)
}

// leakChoice keeps the wanted direction when open, else the current one,
// else stops.
func (g *Game) leakChoice() grid.Point {
	for _, d := range []grid.Point{g.wanted, g.leak.Dir} {
		if d != (grid.Point{}) && g.leakCanEnter(g.leak.Cell.Add(d)) {
			return d
		}
	}
	return grid.Point{}
}

func (g *Game) eat(p grid.Point) {
	switch g.layout.tiles[p.Y][p.X] {
	case TileByte:
		g.score += PointsByte
	case TileMalloc:
		g.score += PointsMalloc
		g.startMalloc()
	default:
		return
	}
	g.layout.tiles[p.Y][p.X] = TileFree
	g.bytesLeft--
}

// startMalloc turns the collectors outside the pool into prey: they turn
// back and flee.
func (g *Game) startMalloc() {
	g.fright = g.frightTicks()
	g.chain = 0
	for i := range g.collectors {
		c := &g.collectors[i]
		if c.state == stateActive {
			c.fleeing = true
			c.reverse()
		}
	}
}

// resolveContacts handles a leak and a collector in the same cell, or
// crossing each other during this tick.
func (g *Game) resolveContacts(leakBefore grid.Point, before []grid.Point) {
	leakNow := g.leak.Occupied()
	for i := range g.collectors {
		c := &g.collectors[i]
		if c.state == stateWaiting {
			continue
		}
		now := c.Occupied()
		if now != leakNow && (now != leakBefore || before[i] != leakNow) {
			continue
		}
		if c.fleeing {
			g.score += PointsCollector << g.chain
			g.chain++
			c.sendHome(g.lifeTicks + RespawnTicks)
			continue
		}
		g.loseLife()
		return
	}
}

func (g *Game) loseLife() {
	g.lives--
	if g.lives == 0 {
		g.lost = true
		return
	}
	g.resetActors()
}

func (g *Game) nextLevel() {
	g.level++
	g.loadMap()
}

func (g *Game) leakCanEnter(p grid.Point) bool {
	t, ok := g.tile(p)
	return ok && t != TileWall && t != TileGate && t != TilePool
}

func (g *Game) tile(p grid.Point) (Tile, bool) {
	if p.Y < 0 || p.Y >= g.height || p.X < 0 || p.X >= g.width {
		return TileWall, false
	}
	return g.layout.tiles[p.Y][p.X], true
}

// Leak returns the leak's position.
func (g *Game) Leak() Actor { return g.leak }

// Collectors returns a copy of the collectors.
func (g *Game) Collectors() []Collector { return append([]Collector(nil), g.collectors...) }

// Tile returns what the map holds at p; outside the map is a wall.
func (g *Game) Tile(p grid.Point) Tile {
	t, _ := g.tile(p)
	return t
}

// FrightLeft is the number of ticks the current malloc still lasts.
func (g *Game) FrightLeft() int { return g.fright }

func (g *Game) Started() bool  { return g.started }
func (g *Game) Width() int     { return g.width }
func (g *Game) Height() int    { return g.height }
func (g *Game) Level() int     { return g.level }
func (g *Game) Score() int     { return g.score }
func (g *Game) Lives() int     { return g.lives }
func (g *Game) BytesLeft() int { return g.bytesLeft }
func (g *Game) Over() bool     { return g.lost }

func neg(p grid.Point) grid.Point { return grid.Point{X: -p.X, Y: -p.Y} }
