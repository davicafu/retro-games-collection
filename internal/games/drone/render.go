package drone

import (
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

const (
	glyphDrone    = '◆'
	glyphMast     = '║'
	glyphUpperTip = '╨' // lowest cell of the mast hanging from the ceiling
	glyphLowerTip = '╥' // highest cell of the mast growing from the ground
	glyphGround   = '═'

	takeOffHint = "Thrust to take off"
)

// layout places the frame (HUD + box) centered on the screen, so overlays,
// which center on the screen too, sit on top of the playfield.
type layout struct {
	originX int // left border column
	originY int // HUD row; the box starts one row below
}

func newLayout(screenWidth, screenHeight, width, height int) layout {
	frameHeight := 1 + 1 + height + 1 // HUD, top border, sky, ground line
	return layout{
		originX: max(0, (screenWidth-(width+2))/2),
		originY: max(0, (screenHeight-frameHeight)/2),
	}
}

// toScreen maps playfield column x and sky row to screen coordinates. The row
// equal to the sky height is the ground line.
func (l layout) toScreen(x, row int) grid.Point {
	return grid.Point{X: l.originX + 1 + x, Y: l.originY + 2 + row}
}

// Draw renders the HUD, playfield, antennas and drone.
func Draw(screen render.Screen, g *Game) {
	width, height := g.Width(), g.Height()
	sw, sh := screen.Size()
	l := newLayout(sw, sh, width, height)

	hud := fmt.Sprintf("Score: %d", g.Score())
	if !g.Started() {
		hud += "   " + takeOffHint
	}
	render.WriteText(screen, l.originX+1, l.originY, render.PadRight(hud, width))

	// The box's bottom edge doubles as the ground line.
	render.DrawBox(screen, l.originX, l.originY+1, width, height)
	ground := l.toScreen(0, height)
	for x := 0; x < width; x++ {
		render.SetCell(screen, ground.X+x, ground.Y, glyphGround)
	}

	for _, a := range g.Antennas() {
		drawAntenna(screen, l, a, width, height)
	}
	if g.Row() < height {
		p := l.toScreen(g.DroneX(), g.Row())
		render.SetCell(screen, p.X, p.Y, glyphDrone)
	}
}

func drawAntenna(screen render.Screen, l layout, a Antenna, width, height int) {
	for x := a.X; x < a.X+AntennaWidth; x++ {
		if x < 0 || x >= width {
			continue
		}
		for row := 0; row < height; row++ {
			if !a.blocks(row) {
				continue
			}
			glyph := glyphMast
			switch row {
			case a.GapTop - 1:
				glyph = glyphUpperTip
			case a.GapTop + GapSize:
				glyph = glyphLowerTip
			}
			p := l.toScreen(x, row)
			render.SetCell(screen, p.X, p.Y, glyph)
		}
	}
}
