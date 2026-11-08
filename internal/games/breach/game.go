// Package breach is the Breakout-inspired game: the player is a virus
// bouncing off an injector to break through a firewall.
package breach

import "time"

// TickPeriod is the real time one Step stands for; every per-tick constant
// below is tuned to it.
const TickPeriod = time.Second / 60

// SubCells is the number of fixed-point units per cell; all motion is integer.
const SubCells = 200

// Velocity is a speed in sub-cells per tick. Positive Y is down.
type Velocity struct{ X, Y int }

// scaled returns v at the given speed percentage.
func (v Velocity) scaled(percent int) Velocity {
	return Velocity{X: v.X * percent / 100, Y: v.Y * percent / 100}
}

// Exit angles off the injector, as magnitudes. Every component stays below
// SubCells so the virus moves at most one cell per axis per tick.
var (
	Steep    = Velocity{X: 17, Y: 37}
	Diagonal = Velocity{X: 30, Y: 30}
	Shallow  = Velocity{X: 40, Y: 20}
)

const (
	PaddleWidth = 7
	// PaddleGlide is how many ticks the injector keeps sliding after a move
	// key. Terminals report presses but not releases, so key repeats refresh
	// the glide while the key is held; it outlasts a typical repeat interval.
	PaddleGlide = 6
	// PaddleSpeed is the glide speed in cells per paddleUnit ticks.
	PaddleSpeed  = 67
	StartLives   = 2
	MaxLives     = 2
	PointsPerHit = 10
)

// Input is the player's intent for one tick: Move is -1, 0 or 1.
type Input struct {
	Move   int
	Launch bool
}

// Config builds a Game. Height counts the rows from the top wall down to the
// injector row, which is the last one. Levels are played in order; Rand,
// when set, drives power-up drops.
type Config struct {
	Width  int
	Height int
	Levels []Level
	Rand   Rand
}

// Game is the headless Breach simulation, advanced one tick at a time.
type Game struct {
	width, height int
	levels        []Level
	level         int // index into levels
	bricks        []Brick
	rng           Rand

	paddleX   int
	wideTicks int

	glideDir  int
	glideLeft int
	glideAcc  int // glide progress towards the next cell, in paddleUnit points

	bx, by   int // virus position in sub-cells
	vx, vy   int
	launched bool

	drops []Drop
	lives int
	score int
	won   bool
	lost  bool
}

func NewGame(cfg Config) *Game {
	g := &Game{
		width:  cfg.Width,
		height: cfg.Height,
		levels: cfg.Levels,
		rng:    cfg.Rand,
		lives:  StartLives,
	}
	g.loadLevel(0)
	return g
}

// Step advances the simulation by one tick. A finished round does not advance.
func (g *Game) Step(in Input) {
	if g.Over() {
		return
	}
	g.movePaddle(in.Move)
	g.tickWide()
	g.fallDrops() // after tickWide, so a Wide caught this tick lasts WideTicks
	switch {
	case g.launched:
		g.moveBall()
	case in.Launch:
		g.launched = true
		v := Diagonal.scaled(g.speed())
		g.vx, g.vy = v.X, -v.Y
		g.moveBall()
	default:
		g.rideInjector()
	}
	if g.cleared() {
		g.nextLevel()
	}
}

// loadLevel builds level i's wall with a fresh, centred injector.
func (g *Game) loadLevel(i int) {
	g.level = i
	g.bricks = g.levels[i].Bricks()
	g.drops = nil
	g.wideTicks = 0
	g.glideLeft = 0
	g.paddleX = (g.width - g.paddleWidth()) / 2
	g.rideInjector()
}

func (g *Game) nextLevel() {
	if g.level+1 >= len(g.levels) {
		g.won = true
		return
	}
	g.loadLevel(g.level + 1)
}

// cleared reports that no breakable block is left.
func (g *Game) cleared() bool {
	for _, b := range g.bricks {
		if !b.Solid {
			return false
		}
	}
	return true
}

func (g *Game) speed() int { return g.levels[g.level].Speed }

// paddleUnit is the number of PaddleSpeed points that move the injector one cell.
const paddleUnit = 100

func (g *Game) movePaddle(move int) {
	if move != 0 {
		if g.glideLeft == 0 || move != g.glideDir {
			// A fresh press moves on this very tick.
			g.glideAcc = paddleUnit - PaddleSpeed
		}
		g.glideDir, g.glideLeft = move, PaddleGlide
	}
	if g.glideLeft == 0 {
		return
	}
	g.glideLeft--
	g.glideAcc += PaddleSpeed
	if g.glideAcc >= paddleUnit {
		g.glideAcc -= paddleUnit
		g.paddleX += g.glideDir
		g.clampPaddle()
	}
}

func (g *Game) clampPaddle() {
	g.paddleX = min(max(g.paddleX, 0), g.width-g.paddleWidth())
}

// rideInjector parks the virus on the cell above the injector's centre.
func (g *Game) rideInjector() {
	g.launched = false
	g.vx, g.vy = 0, 0
	g.bx = (g.paddleX+g.paddleWidth()/2)*SubCells + SubCells/2
	g.by = (g.paddleRow()-1)*SubCells + SubCells/2
}

// moveBall moves one axis at a time so each collision reflects the right component.
func (g *Game) moveBall() {
	nx := g.bx + g.vx
	switch {
	case nx < 0 || nx >= g.width*SubCells:
		g.vx = -g.vx
	case g.hitBrick(nx/SubCells, g.by/SubCells):
		g.vx = -g.vx
	default:
		g.bx = nx
	}

	ny := g.by + g.vy
	switch {
	case ny < 0:
		g.vy = -g.vy
	case ny/SubCells == g.paddleRow() && g.by/SubCells < g.paddleRow() && g.onPaddle(g.bx/SubCells):
		// Only a virus entering the injector row from above bounces; one that
		// already slipped past the injector's edge keeps falling.
		g.bounceOffPaddle()
	case ny/SubCells >= g.height:
		g.loseBall()
	case g.hitBrick(g.bx/SubCells, ny/SubCells):
		g.vy = -g.vy
	default:
		g.by = ny
	}
}

// bounceOffPaddle picks the exit angle from where the virus lands: the
// further from the centre, the shallower. The exact centre keeps the current
// horizontal direction.
func (g *Game) bounceOffPaddle() {
	w := g.paddleWidth()
	rel := 2*(g.bx/SubCells-g.paddleX) - (w - 1) // -(w-1)..w-1, 0 at the centre
	angle := Steep
	switch abs(rel) * 3 / (w - 1) {
	case 2:
		angle = Diagonal
	case 3:
		angle = Shallow
	}
	angle = angle.scaled(g.speed())
	dir := sign(rel)
	if dir == 0 {
		dir = sign(g.vx)
		if dir == 0 {
			dir = 1
		}
	}
	g.vx, g.vy = dir*angle.X, -angle.Y
}

func (g *Game) loseBall() {
	g.lives--
	if g.lives == 0 {
		g.lost = true
		return
	}
	g.rideInjector()
}

// hitBrick damages the block covering the cell, if any, and reports the hit.
func (g *Game) hitBrick(col, row int) bool {
	for i, b := range g.bricks {
		if !b.covers(col, row) {
			continue
		}
		if b.Solid {
			return true
		}
		g.score += PointsPerHit
		g.bricks[i].HP--
		if g.bricks[i].HP == 0 {
			g.bricks = append(g.bricks[:i], g.bricks[i+1:]...)
			g.maybeDrop(b)
		}
		return true
	}
	return false
}

func (g *Game) onPaddle(col int) bool {
	return col >= g.paddleX && col < g.paddleX+g.paddleWidth()
}

func (g *Game) paddleRow() int { return g.height - 1 }

func (g *Game) paddleWidth() int {
	if g.wideTicks > 0 {
		return WidePaddleWidth
	}
	return PaddleWidth
}

// Paddle returns the injector's left column and width.
func (g *Game) Paddle() (int, int) { return g.paddleX, g.paddleWidth() }

// Ball returns the virus position in sub-cells.
func (g *Game) Ball() (int, int) { return g.bx, g.by }

// BallCell returns the virus cell (column, row).
func (g *Game) BallCell() (int, int) { return g.bx / SubCells, g.by / SubCells }

// BallInLowerHalf reports that the virus is in the lower half of its cell.
func (g *Game) BallInLowerHalf() bool { return g.by%SubCells >= SubCells/2 }

func (g *Game) Velocity() (int, int) { return g.vx, g.vy }
func (g *Game) Launched() bool       { return g.launched }
func (g *Game) Lives() int           { return g.lives }
func (g *Game) Score() int           { return g.score }
func (g *Game) Won() bool            { return g.won }
func (g *Game) Over() bool           { return g.won || g.lost }
func (g *Game) Width() int           { return g.width }
func (g *Game) Height() int          { return g.height }

// WideLeft is the number of ticks the Wide power-up still lasts.
func (g *Game) WideLeft() int { return g.wideTicks }

// Level is the 1-based number of the level being played.
func (g *Game) Level() int      { return g.level + 1 }
func (g *Game) LevelCount() int { return len(g.levels) }

func (g *Game) Bricks() []Brick { return append([]Brick(nil), g.bricks...) }
func (g *Game) Drops() []Drop   { return append([]Drop(nil), g.drops...) }

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	}
	return 0
}
