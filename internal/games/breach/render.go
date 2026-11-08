package breach

import (
	"fmt"
	"strings"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

const (
	glyphVirus    = '●'
	glyphInjector = '▀'
	glyphHP3      = '█'
	glyphHP2      = '▓'
	glyphHP1      = '░'
	glyphSolid    = '╳'
	glyphBlockL   = '▐'
	glyphBlockR   = '▌'
	glyphWide     = 'W'
	glyphLife     = '♥'

	launchHint = "↑ to launch"
)

// layout places the frame (HUD + box) centred on the screen, so overlays,
// which centre on the screen too, sit on top of the playfield.
type layout struct {
	originX int // left border column
	originY int // HUD row; the box starts one row below
}

func newLayout(screenWidth, screenHeight, width, height int) layout {
	frameHeight := 1 + height + 2 // HUD, box with its borders
	return layout{
		originX: max(0, (screenWidth-(width+2))/2),
		originY: max(0, (screenHeight-frameHeight)/2),
	}
}

// toScreen maps playfield column and row to screen coordinates.
func (l layout) toScreen(col, row int) grid.Point {
	return grid.Point{X: l.originX + 1 + col, Y: l.originY + 2 + row}
}

// Draw renders the HUD, firewall, drops, injector and virus.
func Draw(screen render.Screen, g *Game) {
	width, height := g.Width(), g.Height()
	sw, sh := screen.Size()
	l := newLayout(sw, sh, width, height)

	hud := fmt.Sprintf("Level %d/%d   Score: %d   %s", g.Level(), g.LevelCount(), g.Score(), strings.Repeat(string(glyphLife), g.Lives()))
	if g.WideLeft() > 0 {
		hud += "   WIDE"
	}
	render.WriteText(screen, l.originX+1, l.originY, render.PadRight(hud, width))
	render.DrawBox(screen, l.originX, l.originY+1, width, height)
	// The banner goes first so game objects stay visible on top of it.
	if !g.Launched() {
		centred(screen, l, width, height-5, fmt.Sprintf("LEVEL %d", g.Level()))
		centred(screen, l, width, height-4, launchHint)
	}

	for _, b := range g.Bricks() {
		drawBrick(screen, l, b)
	}
	for _, d := range g.Drops() {
		glyph := glyphLife
		if d.Kind == PowerUpWide {
			glyph = glyphWide
		}
		set(screen, l, d.Col, d.Y/SubCells, glyph)
	}
	x, w := g.Paddle()
	for col := x; col < x+w; col++ {
		set(screen, l, col, height-1, glyphInjector)
	}
	if col, row := g.BallCell(); row < height {
		set(screen, l, col, row, glyphVirus)
	}
}

// drawBrick draws a block as ▐██▌, ▐▓▓▌ or ▐░░▌ for 3, 2 or 1 hits left,
// and ▐╳╳▌ when indestructible; the edges keep neighbouring blocks distinct.
func drawBrick(screen render.Screen, l layout, b Brick) {
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
		set(screen, l, b.X+i, b.Y, glyph)
	}
}

func set(screen render.Screen, l layout, col, row int, glyph rune) {
	p := l.toScreen(col, row)
	render.SetCell(screen, p.X, p.Y, glyph)
}

// centred writes text centred on a playfield row.
func centred(screen render.Screen, l layout, width, row int, text string) {
	p := l.toScreen((width-len([]rune(text)))/2, row)
	render.WriteText(screen, p.X, p.Y, text)
}
