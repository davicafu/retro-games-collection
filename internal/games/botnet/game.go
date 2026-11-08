// Package botnet is the Space Invaders-inspired game: an infected server
// fires packets at waves of antivirus scanners that descend in formation.
package botnet

import (
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/glide"
)

// TickPeriod is the real time one Step stands for; every per-tick constant
// below is tuned to it.
const TickPeriod = time.Second / 60

// SubCells is the number of fixed-point units per cell; shots move in
// sub-cells and never skip a cell.
const SubCells = 200

const (
	PacketSpeed = 100 // player packet, sub-cells per tick upwards
	ShotSpeed   = 40  // quarantine shot, sub-cells per tick downwards

	ServerWidth = 3
	// The server glides like Breach's injector: 40 cells/s for 100 ms after
	// each key press, so key repeats join into continuous motion.
	ServerSpeed = 67
	ServerGlide = 6

	StartLives = 3
)

// Rand is the randomness enemy fire needs; *math/rand.Rand satisfies it.
type Rand interface {
	Intn(n int) int
}

// Input is the player's intent for one tick: Move is -1, 0 or 1.
type Input struct {
	Move int
	Fire bool
}

// Shot is a quarantine shot falling straight down; Y is in sub-cells.
type Shot struct {
	Col int
	Y   int
}

// Config builds a Game. Height counts the rows from the top of the
// playfield down to the server row, which is the last one.
type Config struct {
	Width  int
	Height int
	Rand   Rand
}

// Game is the headless Botnet simulation, advanced one tick at a time.
type Game struct {
	width, height int
	rng           Rand

	serverX int
	glide   glide.Glide

	packet           bool
	packetCol        int
	packetY          int // sub-cells
	shots            []Shot
	scanners         []Scanner
	total            int // scanners at the start of the wave
	dir              int // formation direction, -1 or 1
	stepTimer, steps int
	cover            []Proxy

	wave      int
	waveTicks int // ticks since the wave started
	score     int
	lives     int
	lost      bool
}

func NewGame(cfg Config) *Game {
	g := &Game{
		width:   cfg.Width,
		height:  cfg.Height,
		rng:     cfg.Rand,
		serverX: (cfg.Width - ServerWidth) / 2,
		glide:   glide.Glide{Speed: ServerSpeed, Ticks: ServerGlide},
		cover:   startingCover(cfg.Width, cfg.Height),
		lives:   StartLives,
	}
	g.startWave(1)
	return g
}

// Step advances the simulation by one tick. A finished round does not advance.
func (g *Game) Step(in Input) {
	if g.lost {
		return
	}
	g.waveTicks++
	if in.Move != 0 {
		g.glide.Press(in.Move)
	}
	g.serverX = min(max(g.serverX+g.glide.Step(), 0), g.width-ServerWidth)

	if in.Fire && !g.packet {
		g.fire()
	} else if g.packet {
		g.movePacket()
	}
	g.moveShots()
	if g.lost {
		return
	}
	g.enemyFire()

	g.stepTimer++
	if g.stepTimer >= g.stepInterval() {
		g.stepTimer = 0
		g.formationStep()
	}
	if len(g.scanners) == 0 && !g.lost {
		g.startWave(g.wave + 1)
	}
}

func (g *Game) serverRow() int { return g.height - 1 }

// fire launches a packet from the server's centre, in the row above it.
func (g *Game) fire() {
	col, row := g.serverX+ServerWidth/2, g.serverRow()-1
	if g.hitByPacket(col, row) {
		return
	}
	g.packet, g.packetCol, g.packetY = true, col, row*SubCells+SubCells/2
}

func (g *Game) movePacket() {
	ny := g.packetY - PacketSpeed
	switch {
	case ny < 0:
		g.packet = false
	case ny/SubCells != g.packetY/SubCells && g.hitByPacket(g.packetCol, ny/SubCells):
		g.packet = false
	default:
		g.packetY = ny
	}
}

// hitByPacket destroys the scanner or erodes the proxy at the cell, if any.
func (g *Game) hitByPacket(col, row int) bool {
	for i, s := range g.scanners {
		if s.covers(col, row) {
			g.score += s.Kind.Points()
			g.scanners = append(g.scanners[:i], g.scanners[i+1:]...)
			return true
		}
	}
	return g.erode(col, row)
}

func (g *Game) moveShots() {
	kept := g.shots[:0]
	for _, s := range g.shots {
		ny := s.Y + ShotSpeed
		row := ny / SubCells
		if row == s.Y/SubCells {
			s.Y = ny
			kept = append(kept, s)
			continue
		}
		switch {
		case row >= g.height:
		case row == g.serverRow() && s.Col >= g.serverX && s.Col < g.serverX+ServerWidth:
			g.loseLife()
			return
		case g.erode(s.Col, row):
		default:
			s.Y = ny
			kept = append(kept, s)
		}
	}
	g.shots = kept
}

func (g *Game) loseLife() {
	g.lives--
	g.shots = nil
	if g.lives == 0 {
		g.lost = true
	}
}

// enemyFire lets the lowest scanner of a random column fire, now and then.
func (g *Game) enemyFire() {
	if g.rng == nil || len(g.scanners) == 0 || len(g.shots) >= g.shotCap() || g.rng.Intn(g.fireOdds()) != 0 {
		return
	}
	shooter := g.scanners[g.rng.Intn(len(g.scanners))]
	for _, s := range g.scanners {
		if s.X == shooter.X && s.Row > shooter.Row {
			shooter = s
		}
	}
	g.shots = append(g.shots, Shot{Col: shooter.X + ScannerWidth/2, Y: (shooter.Row + 1) * SubCells})
}

func (g *Game) Server() (x, width int) { return g.serverX, ServerWidth }

// Packet returns the cell of the packet in flight, if any.
func (g *Game) Packet() (col, row int, ok bool) {
	return g.packetCol, g.packetY / SubCells, g.packet
}

// PacketInLowerHalf reports that the packet is in the lower half of its cell.
func (g *Game) PacketInLowerHalf() bool { return g.packetY%SubCells >= SubCells/2 }

func (g *Game) Shots() []Shot       { return append([]Shot(nil), g.shots...) }
func (g *Game) Scanners() []Scanner { return append([]Scanner(nil), g.scanners...) }
func (g *Game) Cover() []Proxy      { return append([]Proxy(nil), g.cover...) }
func (g *Game) Lives() int          { return g.lives }
func (g *Game) Score() int          { return g.score }
func (g *Game) Wave() int           { return g.wave }
func (g *Game) Over() bool          { return g.lost }
func (g *Game) Width() int          { return g.width }
func (g *Game) Height() int         { return g.height }
func (g *Game) FormationSteps() int { return g.steps }

// WaveTicks is the number of ticks since the current wave started.
func (g *Game) WaveTicks() int { return g.waveTicks }
