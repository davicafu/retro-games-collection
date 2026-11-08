package botnet

import (
	"fmt"
	"strings"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

const (
	spriteServer = "▟█▙"

	// Packets and shots are drawn in the half of the cell they are in, so
	// they move in half-cell steps.
	glyphPacketHigh = '╹'
	glyphPacketLow  = '╻'
	glyphShotHigh   = '╵'
	glyphShotLow    = '╷'

	glyphProxyFull = '▓'
	glyphProxyWorn = '░'
	glyphLife      = '♥'

	// BannerTicks is how long the "WAVE N" banner shows: 1.5 s.
	BannerTicks = 90
)

// scannerSprites holds two animation frames per kind; the formation flips
// between them on every step.
var scannerSprites = [...][2]string{
	KindLow:  {"[◉]", "]◉["},
	KindMid:  {"<◉>", ">◉<"},
	KindHigh: {"╓◉╖", "╙◉╜"},
}

// Draw renders the HUD, cover, scanners, shots, packet and server.
func Draw(screen render.Screen, g *Game) {
	width, height := g.Width(), g.Height()
	f := render.CenterFrame(screen, width, height)
	f.Draw(screen, fmt.Sprintf("Wave %d   Score: %d   %s", g.Wave(), g.Score(), strings.Repeat(string(glyphLife), g.Lives())))

	// The banner goes first so game objects stay visible on top of it.
	if g.WaveTicks() < BannerTicks {
		banner := fmt.Sprintf("WAVE %d", g.Wave())
		text(screen, f, (width-len(banner))/2, height-6, banner)
	}
	for _, p := range g.Cover() {
		glyph := glyphProxyWorn
		if p.HP >= ProxyHP {
			glyph = glyphProxyFull
		}
		set(screen, f, p.X, p.Row, glyph)
	}
	frame := g.FormationSteps() % 2
	for _, s := range g.Scanners() {
		text(screen, f, s.X, s.Row, scannerSprites[s.Kind][frame])
	}
	for _, s := range g.Shots() {
		glyph := glyphShotHigh
		if s.Y%SubCells >= SubCells/2 {
			glyph = glyphShotLow
		}
		set(screen, f, s.Col, s.Y/SubCells, glyph)
	}
	if col, row, ok := g.Packet(); ok {
		glyph := glyphPacketHigh
		if g.PacketInLowerHalf() {
			glyph = glyphPacketLow
		}
		set(screen, f, col, row, glyph)
	}
	x, _ := g.Server()
	text(screen, f, x, height-1, spriteServer)
}

func set(screen render.Screen, f render.Frame, col, row int, glyph rune) {
	x, y := f.Cell(col, row)
	render.SetCell(screen, x, y, glyph)
}

func text(screen render.Screen, f render.Frame, col, row int, s string) {
	x, y := f.Cell(col, row)
	render.WriteText(screen, x, y, s)
}
