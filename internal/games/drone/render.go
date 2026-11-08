package drone

import (
	"fmt"

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

// set draws a glyph at playfield column x and sky row; the row equal to the
// sky height is the ground line, the box's bottom border.
func set(screen render.Screen, f render.Frame, x, row int, glyph rune) {
	sx, sy := f.Cell(x, row)
	render.SetCell(screen, sx, sy, glyph)
}

// Draw renders the HUD, playfield, antennas and drone.
func Draw(screen render.Screen, g *Game) {
	width, height := g.Width(), g.Height()
	f := render.CenterFrame(screen, width, height)
	hud := fmt.Sprintf("Score: %d", g.Score())
	if !g.Started() {
		hud += "   " + takeOffHint
	}
	f.Draw(screen, hud)

	// The box's bottom edge doubles as the ground line.
	for x := 0; x < width; x++ {
		set(screen, f, x, height, glyphGround)
	}

	for _, a := range g.Antennas() {
		drawAntenna(screen, f, a, width, height)
	}
	if g.Row() < height {
		set(screen, f, g.DroneX(), g.Row(), glyphDrone)
	}
}

func drawAntenna(screen render.Screen, f render.Frame, a Antenna, width, height int) {
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
			case a.GapTop + a.Gap:
				glyph = glyphLowerTip
			}
			set(screen, f, x, row, glyph)
		}
	}
}
