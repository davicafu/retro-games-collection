package debris

import (
	"fmt"
	"strings"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/motion"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

const (
	glyphShot = '•'
	glyphLife = '♥'

	// Debris is red, what can hurt the drone; the drone and its shots are
	// in colours of their own.
	colorPiece = render.ColorRed
	colorDrone = render.ColorCyan
	colorShot  = render.ColorYellow

	// BlinkPeriod is how many ticks a shielded drone stays shown or hidden.
	BlinkPeriod = 8
)

// droneGlyphs shows the drone by heading, one glyph per eighth of a turn,
// starting at up and turning clockwise.
var droneGlyphs = []rune("▲◥▶◢▼◣◀◤")

// shapes draws each size centred on its position: a large piece is 4 by 2
// cells, a medium one 2 by 1, a small one a single cell.
var shapes = [...][]string{
	Small:  {"▪"},
	Medium: {"▓▓"},
	Large:  {"▟██▙", "▜██▛"},
}

// Draw renders the HUD, the pieces, the shots and the drone.
func Draw(screen render.Screen, g *Game) {
	f := render.CenterFrame(screen, g.Cols(), g.Rows())
	f.Draw(screen, fmt.Sprintf("Wave %d   Score: %d   %s", g.Wave(), g.Score(), strings.Repeat(string(glyphLife), g.Lives())))
	// set draws at a playfield cell, wrapping across the edges like the
	// space itself.
	set := func(col, row int, r rune, color render.Color) {
		col = ((col % g.Cols()) + g.Cols()) % g.Cols()
		row = ((row % g.Rows()) + g.Rows()) % g.Rows()
		x, y := f.Cell(col, row)
		screen.SetCell(x, y, render.Cell{Rune: r, Color: color})
	}
	for _, p := range g.Pieces() {
		shape := shapes[p.Size]
		left := firstCell(p.Pos.X, len([]rune(shape[0])), CellW)
		top := firstCell(p.Pos.Y, len(shape), CellH)
		for dy, line := range shape {
			for dx, r := range []rune(line) {
				set(left+dx, top+dy, r, colorPiece)
			}
		}
	}
	for _, s := range g.Shots() {
		col, row := cellOf(s.Pos.X, s.Pos.Y)
		set(col, row, glyphShot, colorShot)
	}
	if g.Shielded() && g.ShieldLeft()/BlinkPeriod%2 == 1 {
		return
	}
	d := g.Drone()
	col, row := cellOf(d.Pos.X, d.Pos.Y)
	set(col, row, droneGlyphs[droneGlyph(d.Heading)], colorDrone)
}

func cellOf(x, y int) (int, int) { return x / CellW, y / CellH }

// firstCell is the first of n cells of the given size centred on pos, the
// nearest the drawing can come to the shape the hits use.
func firstCell(pos, n, size int) int {
	a := pos - n*size/2 + size/2
	return (a - (a%size+size)%size) / size
}

// droneGlyph is the index in droneGlyphs of the eighth of a turn nearest
// to a heading.
func droneGlyph(heading int) int {
	per := motion.Headings / len(droneGlyphs)
	return (heading + per/2) / per % len(droneGlyphs)
}
