// Package rack is the platformer inspired by Donkey Kong and Lode Runner: a
// technician climbs the racks of a data center collecting disks, dodging
// loose cables that roll down the floors and patrolling bots, to reach the
// server room at the top.
package rack

import (
	"maps"
	"slices"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/glide"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/platform"
)

// TickPeriod is the real time one Step stands for; every per-tick constant
// below is tuned to it.
const TickPeriod = time.Second / 60

const (
	// WalkSpeed is how fast the technician walks and climbs while a key is
	// held, in cells per glide.Unit ticks: about 8 cells/s. WalkGlide is how
	// long it keeps going after a press: a tap moves exactly one cell.
	WalkSpeed = 13
	WalkGlide = 7
	// FallTicks is how many ticks a fall takes per cell: 15 cells/s.
	FallTicks = 4
	// JumpTicks is how long a jump holds the technician a cell up: 0.35 s.
	JumpTicks = 21

	// BotTicks is how many ticks a bot takes per cell: 4 cells/s.
	BotTicks = 15
	// A cable rolls a cell every CableTicks ticks (6.7 cells/s) and the
	// spool drops one every CableInterval ticks (4 s) on level 1; each level
	// takes CableTicksStep and CableIntervalStep off, down to the minimums.
	CableTicks        = 9
	CableTicksStep    = 1
	MinCableTicks     = 6
	CableInterval     = 240
	CableIntervalStep = 30
	MinCableInterval  = 120
	// LadderChance is 1 in how many ladder tops a cable rolls down.
	LadderChance = 4

	// A toolbox gives WrenchesPerToolbox wrenches; a thrown wrench flies a
	// cell every WrenchTicks ticks (20 cells/s) along the technician's row,
	// over gaps, until it hits a wall, a bot or a cable.
	WrenchesPerToolbox = 3
	WrenchTicks        = 3

	StartLives = 3

	PointsDisk = 100
	PointsJump = 50
	PointsExit = 500
	// PointsBot and PointsCable are for disabling a bot and knocking a
	// cable away with a wrench.
	PointsBot   = 300
	PointsCable = 100
)

var (
	up   = grid.Point{Y: -1}
	down = grid.Point{Y: 1}
)

// Rand is the randomness the cables need; *math/rand.Rand satisfies it.
type Rand interface {
	Intn(n int) int
}

// Input is the player's intent for one tick: Move is a press of an arrow,
// one cell left, right, up or down, or none.
type Input struct {
	Move  grid.Point
	Jump  bool
	Throw bool
}

// Player is the technician.
type Player struct {
	Pos      grid.Point
	jumpLeft int
}

// Jumping reports whether the technician is in the air on a jump.
func (p Player) Jumping() bool { return p.jumpLeft > 0 }

// Cable is a loose cable rolling along Dir (-1 or 1). It falls where
// nothing holds it and, when descending, goes down a ladder.
type Cable struct {
	Pos        grid.Point
	Dir        int
	descending bool
	jumped     bool // already scored for a jump over it
}

// Bot is a bot patrolling its floor along Dir.
type Bot struct {
	Pos grid.Point
	Dir int
}

// Wrench is a thrown wrench flying along Dir (-1 or 1).
type Wrench struct {
	Pos grid.Point
	Dir int
}

// Cell is what a level shows at a cell.
type Cell int

const (
	CellEmpty Cell = iota
	CellFloor
	CellLadder
	CellDisk
	CellExit
	CellSpool
	CellToolbox
)

// level is one parsed map.
type level struct {
	tiles  [][]platform.Tile
	disks  map[grid.Point]bool
	tools  map[grid.Point]bool
	start  grid.Point
	exit   grid.Point
	spool  grid.Point
	bots   []grid.Point
	width  int
	height int
}

func parseLevel(rows []string) level {
	l := level{disks: map[grid.Point]bool{}, tools: map[grid.Point]bool{}, height: len(rows)}
	for y, row := range rows {
		line := make([]platform.Tile, 0, len(row))
		for x, r := range []rune(row) {
			p := grid.Point{X: x, Y: y}
			tile := platform.Empty
			switch r {
			case '=':
				tile = platform.Solid
			case 'H':
				tile = platform.Ladder
			case '$':
				l.disks[p] = true
			case 'T':
				l.tools[p] = true
			case 'P':
				l.start = p
			case 'E':
				l.exit = p
			case 'C':
				l.spool = p
			case 'B':
				l.bots = append(l.bots, p)
			}
			line = append(line, tile)
		}
		l.width = max(l.width, len(line))
		l.tiles = append(l.tiles, line)
	}
	return l
}

// Config builds a Game: Levels are the maps played in turn.
type Config struct {
	Levels [][]string
	Rand   Rand
}

// Game is the headless Rack simulation, advanced one tick at a time.
type Game struct {
	levels [][]string
	rng    Rand

	lvl    level
	disks  map[grid.Point]bool
	tools  map[grid.Point]bool
	player Player
	walk   glide.Glide
	facing grid.Point // the direction the walk glide moves
	fall   int        // ticks into the current cell of a fall
	aim    int        // the way the technician faces, -1 or 1
	cables []Cable
	bots   []Bot

	wrenches    int // in hand
	thrown      []Wrench
	wrenchClock int

	cableClock int
	botClock   int
	spawnClock int

	level int
	score int
	lives int
	over  bool
}

func NewGame(cfg Config) *Game {
	g := &Game{
		levels: cfg.Levels, rng: cfg.Rand,
		level: 1, lives: StartLives,
		walk: glide.Glide{Speed: WalkSpeed, Ticks: WalkGlide},
	}
	g.load()
	return g
}

// load starts the current level afresh.
func (g *Game) load() {
	g.lvl = parseLevel(g.levels[(g.level-1)%len(g.levels)])
	g.disks = maps.Clone(g.lvl.disks)
	g.tools = maps.Clone(g.lvl.tools)
	g.restart()
}

// restart puts the technician back at the start, clears the cables and the
// wrenches in flight and sends every bot, disabled ones too, back to its
// post; collected disks and toolboxes stay collected, wrenches in hand stay.
func (g *Game) restart() {
	g.player = Player{Pos: g.lvl.start}
	g.walk.Stop()
	g.fall = 0
	g.aim = 1
	g.cables = nil
	g.thrown = nil
	g.bots = nil
	for i, p := range g.lvl.bots {
		g.bots = append(g.bots, Bot{Pos: p, Dir: 1 - 2*(i%2)})
	}
	g.cableClock, g.botClock, g.spawnClock, g.wrenchClock = 0, 0, 0, 0
}

// Tile is the level's structure at p; outside the level is solid.
func (g *Game) Tile(p grid.Point) platform.Tile {
	if p.Y < 0 || p.Y >= len(g.lvl.tiles) || p.X < 0 || p.X >= len(g.lvl.tiles[p.Y]) {
		return platform.Solid
	}
	return g.lvl.tiles[p.Y][p.X]
}

// Step advances the simulation by one tick. A finished round does not
// advance.
func (g *Game) Step(in Input) {
	if g.over {
		return
	}
	g.movePlayer(in)
	if g.disks[g.player.Pos] {
		delete(g.disks, g.player.Pos)
		g.score += PointsDisk
	}
	if g.tools[g.player.Pos] {
		delete(g.tools, g.player.Pos)
		g.wrenches += WrenchesPerToolbox
	}
	if g.player.Pos == g.lvl.exit && g.DoorOpen() {
		g.score += PointsExit
		g.level++
		g.load()
		return
	}
	if in.Throw && g.wrenches > 0 && len(g.thrown) == 0 {
		g.wrenches--
		g.thrown = []Wrench{{Pos: g.player.Pos, Dir: g.aim}}
		g.wrenchClock = 0
		g.moveWrenches()
	} else if g.wrenchClock++; g.wrenchClock >= WrenchTicks {
		g.wrenchClock = 0
		g.moveWrenches()
	}
	g.scoreJumps()
	if g.caught() {
		return
	}
	if g.cableClock++; g.cableClock >= g.cableTicks() {
		g.cableClock = 0
		g.moveCables()
	}
	if g.botClock++; g.botClock >= BotTicks {
		g.botClock = 0
		g.moveBots()
	}
	if g.spawnClock++; g.spawnClock >= g.cableInterval() {
		g.spawnClock = 0
		g.spawnCable()
	}
	g.knockDown()
	g.scoreJumps()
	g.caught()
}

func (g *Game) movePlayer(in Input) {
	p := &g.player
	if p.jumpLeft > 0 {
		p.jumpLeft--
	}
	if in.Move.X != 0 {
		g.aim = in.Move.X
	}
	if in.Move != (grid.Point{}) {
		if in.Move != g.facing {
			g.walk.Stop()
			g.facing = in.Move
		}
		g.walk.Press(1)
	}
	if p.jumpLeft == 0 && !platform.Supported(g, p.Pos) {
		// Falling: no steering until it lands.
		if g.fall++; g.fall >= FallTicks {
			g.fall = 0
			p.Pos = p.Pos.Add(down)
		}
		return
	}
	g.fall = 0
	if in.Jump && p.jumpLeft == 0 && g.Tile(p.Pos) != platform.Ladder && platform.Open(g, p.Pos.Add(up)) {
		p.Pos = p.Pos.Add(up)
		p.jumpLeft = JumpTicks
		return
	}
	if g.walk.Step() == 0 {
		return
	}
	switch g.facing {
	case up:
		if platform.CanClimbUp(g, p.Pos) {
			p.Pos = p.Pos.Add(up)
		}
	case down:
		if platform.CanClimbDown(g, p.Pos) {
			p.Pos = p.Pos.Add(down)
		}
	default:
		if next := p.Pos.Add(g.facing); platform.Open(g, next) {
			p.Pos = next
		}
	}
}

// moveWrenches moves each wrench in flight a cell, dropping it at a wall,
// and knocks down what it hits.
func (g *Game) moveWrenches() {
	kept := g.thrown[:0]
	for _, w := range g.thrown {
		next := w.Pos.Add(grid.Point{X: w.Dir})
		if platform.Open(g, next) {
			w.Pos = next
			kept = append(kept, w)
		}
	}
	g.thrown = kept
	g.knockDown()
}

// knockDown disables each bot and knocks away each cable a wrench in flight
// is on; the wrench is spent on the first one it hits.
func (g *Game) knockDown() {
	kept := g.thrown[:0]
	for _, w := range g.thrown {
		if i := slices.IndexFunc(g.bots, func(b Bot) bool { return b.Pos == w.Pos }); i >= 0 {
			g.bots = slices.Delete(g.bots, i, i+1)
			g.score += PointsBot
			continue
		}
		if i := slices.IndexFunc(g.cables, func(c Cable) bool { return c.Pos == w.Pos }); i >= 0 {
			g.cables = slices.Delete(g.cables, i, i+1)
			g.score += PointsCable
			continue
		}
		kept = append(kept, w)
	}
	g.thrown = kept
}

// scoreJumps scores each cable passing under a jumping technician, once.
func (g *Game) scoreJumps() {
	if !g.player.Jumping() {
		return
	}
	for i := range g.cables {
		if c := &g.cables[i]; !c.jumped && c.Pos == g.player.Pos.Add(down) {
			c.jumped = true
			g.score += PointsJump
		}
	}
}

// caught costs a life when a cable or a bot touches the technician and
// reports whether one did.
func (g *Game) caught() bool {
	hit := false
	for _, c := range g.cables {
		hit = hit || c.Pos == g.player.Pos
	}
	for _, b := range g.bots {
		hit = hit || b.Pos == g.player.Pos
	}
	if !hit {
		return false
	}
	if g.lives--; g.lives == 0 {
		g.over = true
		return true
	}
	g.restart()
	return true
}

// cableTicks is how many ticks a cable takes per cell on the current level.
func (g *Game) cableTicks() int {
	return max(CableTicks-CableTicksStep*(g.level-1), MinCableTicks)
}

// cableInterval is how many ticks pass between cables on the current level.
func (g *Game) cableInterval() int {
	return max(CableInterval-CableIntervalStep*(g.level-1), MinCableInterval)
}

// spawnCable drops a cable at the spool, rolling right towards the top
// floor's gap.
func (g *Game) spawnCable() {
	g.cables = append(g.cables, Cable{Pos: g.lvl.spool, Dir: 1})
}

// moveCables moves every cable one cell: down a ladder it is descending,
// down where nothing holds it, otherwise along its floor. A cable turns
// back when it lands, at the foot of a ladder and at a wall, and leaves at
// the bottom floor's ends.
func (g *Game) moveCables() {
	kept := g.cables[:0]
	for _, c := range g.cables {
		if g.moveCable(&c) {
			kept = append(kept, c)
		}
	}
	g.cables = kept
}

func (g *Game) moveCable(c *Cable) bool {
	switch {
	case c.descending:
		if platform.CanClimbDown(g, c.Pos) {
			c.Pos = c.Pos.Add(down)
			return true
		}
		c.descending = false
		c.Dir = -c.Dir
		return true
	case !platform.Supported(g, c.Pos):
		c.Pos = c.Pos.Add(down)
		if platform.Supported(g, c.Pos) {
			c.Dir = -c.Dir
		}
		return true
	}
	next := c.Pos.Add(grid.Point{X: c.Dir})
	if next.X < 0 || next.X >= g.lvl.width {
		if c.Pos.Y == g.lvl.height-2 {
			return false
		}
		c.Dir = -c.Dir
		return true
	}
	if !platform.Open(g, next) {
		c.Dir = -c.Dir
		return true
	}
	c.Pos = next
	if g.Tile(c.Pos.Add(down)) == platform.Ladder && g.rng.Intn(LadderChance) == 0 {
		c.descending = true
	}
	return true
}

// moveBots moves every bot one cell along its floor, turning instead at a
// wall or before a gap. Bots walk over ladders and never climb them.
func (g *Game) moveBots() {
	for i := range g.bots {
		b := &g.bots[i]
		next := b.Pos.Add(grid.Point{X: b.Dir})
		if platform.Open(g, next) && platform.Supported(g, next) {
			b.Pos = next
		} else {
			b.Dir = -b.Dir
		}
	}
}

// Cell is what the level shows at p.
func (g *Game) Cell(p grid.Point) Cell {
	switch {
	case g.disks[p]:
		return CellDisk
	case g.tools[p]:
		return CellToolbox
	case p == g.lvl.exit:
		return CellExit
	case p == g.lvl.spool:
		return CellSpool
	}
	switch g.Tile(p) {
	case platform.Solid:
		return CellFloor
	case platform.Ladder:
		return CellLadder
	}
	return CellEmpty
}

func (g *Game) Width() int        { return g.lvl.width }
func (g *Game) Height() int       { return g.lvl.height }
func (g *Game) Player() Player    { return g.player }
func (g *Game) Cables() []Cable   { return append([]Cable(nil), g.cables...) }
func (g *Game) Bots() []Bot       { return append([]Bot(nil), g.bots...) }
func (g *Game) DisksLeft() int    { return len(g.disks) }
func (g *Game) WrenchesLeft() int { return g.wrenches }
func (g *Game) Thrown() []Wrench  { return append([]Wrench(nil), g.thrown...) }
func (g *Game) DoorOpen() bool    { return len(g.disks) == 0 }
func (g *Game) Score() int        { return g.score }
func (g *Game) Lives() int        { return g.lives }
func (g *Game) Level() int        { return g.level }
func (g *Game) Over() bool        { return g.over }
func (g *Game) Exit() grid.Point  { return g.lvl.exit }
func (g *Game) Spool() grid.Point { return g.lvl.spool }
