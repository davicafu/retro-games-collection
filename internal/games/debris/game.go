// Package debris is the Asteroids-inspired game: a maintenance drone in orbit
// breaks dead satellites into ever smaller pieces, in a space whose edges
// wrap around.
package debris

import (
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/glide"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/motion"
)

// TickPeriod is the real time one Step stands for; every per-tick constant
// below is tuned to it.
const TickPeriod = time.Second / 60

// Physics runs in integer units where a cell is CellW wide and CellH tall:
// terminal cells are about twice as tall as wide, so motion looks the same
// in every direction. Speeds are in units per tick.
const (
	CellW = 1000
	CellH = 2000
)

const (
	// TurnSpeed is how fast the drone turns while a turn key is held, in
	// headings per glide.Unit ticks: a full turn in about 1.3 s. TurnGlide
	// is how long it keeps turning after a press: long enough that a held
	// key's repeats join into a steady turn, short enough that a tap turns
	// exactly one heading, for fine aim.
	TurnSpeed = 40
	TurnGlide = 3

	// ThrustTicks is how long one press of the thruster accelerates the
	// drone: terminals report no key release, and a held key's repeats join
	// bursts into continuous thrust. It is long enough to keep thrusting
	// while the other hand taps a turn, which stops the thrust key's repeats.
	ThrustTicks   = 24
	ThrustAccel   = 6
	MaxDroneSpeed = 300 // 18 cells/s
	// DragDen is the drag: every tick the drone loses 1/DragDen of its
	// velocity (half of it in about 1.4 s).
	DragDen = 120

	ShotSpeed    = 600 // 36 cells/s, added to the drone's velocity
	ShotTicks    = 45  // a shot's flight, 0.75 s
	MaxShots     = 4
	FireCooldown = 8

	StartLives = 3
	// ShieldTicks is how long a respawned drone cannot be hit: 2 s.
	ShieldTicks = 120
	// WaveDelay is the pause between clearing a wave and the next: 1.5 s.
	WaveDelay = 90
	// SafeDistance is how far from the drone a wave's pieces appear.
	SafeDistance = 15 * CellW

	FirstPieces = 4
	MaxPieces   = 8

	PointsLarge  = 20
	PointsMedium = 50
	PointsSmall  = 100

	droneRadius = 450
)

// Size is a piece of debris' size.
type Size int

const (
	Small Size = iota
	Medium
	Large
)

// radius is how far from its centre a piece can be hit, matching its drawing.
var radius = [...]int{Small: 600, Medium: 1000, Large: 1900}

var points = [...]int{Small: PointsSmall, Medium: PointsMedium, Large: PointsLarge}

// baseSpeed is each size's speed at wave 1: 10.2, 8.4 and 5.4 cells/s.
var baseSpeed = [...]int{Small: 170, Medium: 140, Large: 90}

// PiecesAtWave is how many large pieces wave n starts with.
func PiecesAtWave(n int) int { return min(FirstPieces+n-1, MaxPieces) }

// PieceSpeed is a size's speed in wave n: 8% faster each wave, up to 60%.
func PieceSpeed(size Size, wave int) int {
	return baseSpeed[size] * (100 + min(8*(wave-1), 60)) / 100
}

// Rand is the randomness the waves and splits need; *math/rand.Rand
// satisfies it.
type Rand interface {
	Intn(n int) int
}

// Input is the player's intent for one tick: Turn is a press of a turn key,
// -1 (anticlockwise), 1 (clockwise) or 0.
type Input struct {
	Turn   int
	Thrust bool
	Fire   bool
}

// Drone is the player's drone.
type Drone struct {
	motion.Body
	Heading int
}

// Piece is a piece of debris drifting in a straight line along Heading.
type Piece struct {
	motion.Body
	Size    Size
	Heading int
}

// Shot is a shot in flight with Left ticks to go.
type Shot struct {
	motion.Body
	Left int
}

// Config builds a Game. Pieces, when set, replace the first wave.
type Config struct {
	Cols, Rows int
	Rand       Rand
	Pieces     []Piece
}

// Game is the headless Debris simulation, advanced one tick at a time.
type Game struct {
	cols, rows int
	space      motion.Space
	rng        Rand

	drone      Drone
	turn       glide.Glide
	thrustLeft int
	cooldown   int
	shield     int
	shots      []Shot
	pieces     []Piece

	wave    int
	waiting int // ticks since the last wave was cleared
	score   int
	lives   int
	over    bool
}

func NewGame(cfg Config) *Game {
	g := &Game{
		cols: cfg.Cols, rows: cfg.Rows, rng: cfg.Rand,
		space: motion.Space{W: cfg.Cols * CellW, H: cfg.Rows * CellH},
		lives: StartLives, wave: 1,
		turn: glide.Glide{Speed: TurnSpeed, Ticks: TurnGlide},
	}
	g.respawn()
	g.shield = 0
	if cfg.Pieces != nil {
		g.pieces = append([]Piece(nil), cfg.Pieces...)
	} else {
		g.spawnWave()
	}
	return g
}

func (g *Game) centre() motion.Vec { return motion.Vec{X: g.space.W / 2, Y: g.space.H / 2} }

func (g *Game) respawn() {
	g.drone = Drone{Body: motion.Body{Pos: g.centre()}}
	g.thrustLeft = 0
	g.turn.Stop()
	g.shield = ShieldTicks
}

// spawnWave places the wave's large pieces away from the drone, each
// drifting along a random heading.
func (g *Game) spawnWave() {
	for range PiecesAtWave(g.wave) {
		var p motion.Vec
		for try := 0; try < 100; try++ {
			p = motion.Vec{X: g.rng.Intn(g.space.W), Y: g.rng.Intn(g.space.H)}
			if !g.space.Near(p, g.drone.Pos, SafeDistance) {
				break
			}
		}
		g.pieces = append(g.pieces, g.newPiece(Large, p, g.rng.Intn(motion.Headings)))
	}
}

func (g *Game) newPiece(size Size, pos motion.Vec, heading int) Piece {
	vel := motion.Heading(heading).Scale(PieceSpeed(size, g.wave), motion.Unit)
	return Piece{Body: motion.Body{Pos: pos, Vel: vel}, Size: size, Heading: heading}
}

// Step advances the simulation by one tick. A finished round does not
// advance.
func (g *Game) Step(in Input) {
	if g.over {
		return
	}
	g.steer(in)
	g.ageShots()
	for i := range g.pieces {
		g.pieces[i].Move(g.space)
	}
	g.shootDown()
	g.collide()
	if len(g.pieces) == 0 {
		if g.waiting++; g.waiting >= WaveDelay {
			g.waiting = 0
			g.wave++
			g.spawnWave()
		}
	}
}

func (g *Game) steer(in Input) {
	d := &g.drone
	if in.Turn != 0 {
		g.turn.Press(in.Turn)
	}
	d.Heading = ((d.Heading+g.turn.Step())%motion.Headings + motion.Headings) % motion.Headings
	if in.Thrust {
		g.thrustLeft = ThrustTicks
	}
	if g.thrustLeft > 0 {
		g.thrustLeft--
		d.Vel = d.Vel.Add(motion.Heading(d.Heading).Scale(ThrustAccel, motion.Unit))
	} else {
		d.Vel = motion.Damp(d.Vel, DragDen)
	}
	d.Vel = motion.Limit(d.Vel, MaxDroneSpeed)
	d.Move(g.space)

	if g.cooldown > 0 {
		g.cooldown--
	}
	if in.Fire && g.cooldown == 0 && len(g.shots) < MaxShots {
		vel := d.Vel.Add(motion.Heading(d.Heading).Scale(ShotSpeed, motion.Unit))
		g.shots = append(g.shots, Shot{Body: motion.Body{Pos: d.Pos, Vel: vel}, Left: ShotTicks})
		g.cooldown = FireCooldown
	}
	if g.shield > 0 {
		g.shield--
	}
}

// ageShots drops the shots at the end of their range.
func (g *Game) ageShots() {
	kept := g.shots[:0]
	for _, s := range g.shots {
		if s.Left--; s.Left > 0 {
			kept = append(kept, s)
		}
	}
	g.shots = kept
}

// shootDown moves each shot in two half steps, so a fast shot cannot pass
// through a small piece between ticks, and breaks the pieces they hit.
func (g *Game) shootDown() {
	kept := g.shots[:0]
	for _, s := range g.shots {
		half := s.Vel.Scale(1, 2)
		hit := false
		for step := 0; step < 2 && !hit; step++ {
			s.Pos = g.space.Wrap(s.Pos.Add(half))
			for i, p := range g.pieces {
				if g.space.Near(s.Pos, p.Pos, radius[p.Size]) {
					g.hit(i)
					hit = true
					break
				}
			}
		}
		if !hit {
			kept = append(kept, s)
		}
	}
	g.shots = kept
}

// hit scores piece i and breaks it into two smaller pieces flying apart,
// or destroys it when it is small.
func (g *Game) hit(i int) {
	p := g.pieces[i]
	g.score += points[p.Size]
	g.pieces = append(g.pieces[:i], g.pieces[i+1:]...)
	if p.Size == Small {
		return
	}
	spread := 2 + g.rng.Intn(2)
	for _, h := range []int{p.Heading - spread, p.Heading + spread} {
		g.pieces = append(g.pieces, g.newPiece(p.Size-1, p.Pos, h))
	}
}

func (g *Game) collide() {
	if g.shield > 0 {
		return
	}
	for _, p := range g.pieces {
		if g.space.Near(g.drone.Pos, p.Pos, radius[p.Size]+droneRadius) {
			if g.lives--; g.lives == 0 {
				g.over = true
				return
			}
			g.respawn()
			return
		}
	}
}

func (g *Game) Drone() Drone        { return g.drone }
func (g *Game) Pieces() []Piece     { return append([]Piece(nil), g.pieces...) }
func (g *Game) Shots() []Shot       { return append([]Shot(nil), g.shots...) }
func (g *Game) Shielded() bool      { return g.shield > 0 }
func (g *Game) ShieldLeft() int     { return g.shield }
func (g *Game) Score() int          { return g.score }
func (g *Game) Lives() int          { return g.lives }
func (g *Game) Wave() int           { return g.wave }
func (g *Game) Over() bool          { return g.over }
func (g *Game) Cols() int           { return g.cols }
func (g *Game) Rows() int           { return g.rows }
func (g *Game) Space() motion.Space { return g.space }
