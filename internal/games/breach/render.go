package breach

import (
	"fmt"
	"strings"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

const (
	// The virus is drawn as a half block so it moves in half-cell steps.
	glyphVirusHigh = '▀'
	glyphVirusLow  = '▄'
	glyphInjector  = '▀'
	glyphHP3       = '█'
	glyphHP2       = '▓'
	glyphHP1       = '░'
	glyphSolid     = '╳'
	glyphBlockL    = '▐'
	glyphBlockR    = '▌'
	glyphWide      = 'W'
	glyphLife      = '♥'

	launchHint = "↑ to launch"
)

// Draw renders the HUD, firewall, drops, injector and virus.
func Draw(screen render.Screen, g *Game) {
	width, height := g.Width(), g.Height()
	f := render.CenterFrame(screen, width, height)

	hud := fmt.Sprintf("Level %d/%d   Score: %d   %s", g.Level(), g.LevelCount(), g.Score(), strings.Repeat(string(glyphLife), g.Lives()))
	if g.WideLeft() > 0 {
		hud += "   WIDE"
	}
	f.Draw(screen, hud)
	// The banner goes first so game objects stay visible on top of it.
	if !g.Launched() {
		centred(screen, f, width, height-5, fmt.Sprintf("LEVEL %d", g.Level()))
		centred(screen, f, width, height-4, launchHint)
	}

	for _, b := range g.Bricks() {
		drawBrick(screen, f, b)
	}
	for _, d := range g.Drops() {
		glyph := glyphLife
		if d.Kind == PowerUpWide {
			glyph = glyphWide
		}
		set(screen, f, d.Col, d.Y/SubCells, glyph)
	}
	x, w := g.Paddle()
	for col := x; col < x+w; col++ {
		set(screen, f, col, height-1, glyphInjector)
	}
	if col, row := g.BallCell(); row < height {
		glyph := glyphVirusHigh
		if g.BallInLowerHalf() {
			glyph = glyphVirusLow
		}
		set(screen, f, col, row, glyph)
	}
}

// drawBrick draws a block as ▐██▌, ▐▓▓▌ or ▐░░▌ for 3, 2 or 1 hits left,
// and ▐╳╳▌ when indestructible; the edges keep neighbouring blocks distinct.
func drawBrick(screen render.Screen, f render.Frame, b Brick) {
	fill := glyphHP1
	switch {
	case b.Solid:
		fill = glyphSolid
	case b.HP >= 3:
		fill = glyphHP3
	case b.HP == 2:
		fill = glyphHP2
	}
	for i := 0; i < BrickWidth; i++ {
		glyph := fill
		switch i {
		case 0:
			glyph = glyphBlockL
		case BrickWidth - 1:
			glyph = glyphBlockR
		}
		set(screen, f, b.X+i, b.Y, glyph)
	}
}

func set(screen render.Screen, f render.Frame, col, row int, glyph rune) {
	x, y := f.Cell(col, row)
	render.SetCell(screen, x, y, glyph)
}

// centred writes text centred on a playfield row.
func centred(screen render.Screen, f render.Frame, width, row int, text string) {
	x, y := f.Cell((width-len([]rune(text)))/2, row)
	render.WriteText(screen, x, y, text)
}
