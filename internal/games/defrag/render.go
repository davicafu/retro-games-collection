package defrag

import (
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// The disk is drawn 2 characters per cell, with its right side and the next
// fragment beside it.
const (
	playfieldCols = 2*Width + 1 + panelCols
	panelCols     = 11

	spriteEmpty    = " ·"
	spriteBlank    = "  "
	spriteLocked   = "▓▓"
	spriteFragment = "██"
	spriteShadow   = "░░"
	spriteFlash    = "██"
	glyphSide      = '│'

	// FlashPeriod is how many ticks clearing rows stay shown or hidden.
	FlashPeriod = 3
)

// Every kind of fragment has its own colour; clearing rows flash white.
var kindColors = [kinds]render.Color{
	Bar:    render.ColorRed,
	Square: render.ColorCyan,
	Tee:    render.ColorYellow,
	Ess:    render.ColorBlue,
	Zed:    render.ColorMagenta,
	Jay:    render.ColorGreen,
	Ell:    render.ColorWhite,
}

const (
	colorEmpty  = render.ColorDefault
	colorShadow = render.ColorDefault
	colorFlash  = render.ColorWhite
)

// Draw renders the score, the disk with the falling fragment and beside it
// the next fragment, the level and the sectors. With guides on, empty cells
// show a dot and a shadow shows where the fragment will land.
func Draw(screen render.Screen, g *Game, guides bool) {
	f := render.CenterFrame(screen, playfieldCols, Height)
	f.Draw(screen, fmt.Sprintf("Score: %d", g.Score()))
	set := func(col, row int, sprite string, color render.Color) {
		x, y := f.Cell(col, row)
		render.WriteColoredText(screen, x, y, sprite, color)
	}
	clearing := map[int]bool{}
	for _, y := range g.Clearing() {
		clearing[y] = true
	}
	flashOn := g.ClearClock()/FlashPeriod%2 == 0
	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			v := g.At(grid.Point{X: x, Y: y})
			switch {
			case clearing[y] && flashOn:
				set(2*x, y, spriteFlash, colorFlash)
			case (clearing[y] || v == 0) && guides:
				set(2*x, y, spriteEmpty, colorEmpty)
			case clearing[y] || v == 0:
				set(2*x, y, spriteBlank, colorEmpty)
			default:
				set(2*x, y, spriteLocked, kindColors[v-1])
			}
		}
		set(2*Width, y, string(glyphSide), render.ColorDefault)
	}
	if len(clearing) == 0 {
		fr := g.Fragment()
		ghost := g.Ghost()
		for _, c := range fr.Shape.Cells {
			if p := ghost.Add(c); p.Y >= 0 && guides {
				set(2*p.X, p.Y, spriteShadow, colorShadow)
			}
		}
		for _, c := range fr.Shape.Cells {
			if p := fr.Pos.Add(c); p.Y >= 0 {
				set(2*p.X, p.Y, spriteFragment, kindColors[fr.Kind])
			}
		}
	}
	left := 2*Width + 2
	set(left+1, 0, "NEXT", render.ColorDefault)
	next := g.Next()
	for _, c := range g.Shape(next).Cells {
		set(left+1+2*c.X, 2+c.Y, spriteFragment, kindColors[next])
	}
	set(left+1, 7, "LEVEL", render.ColorDefault)
	set(left+1, 8, fmt.Sprint(g.Level()), render.ColorDefault)
	set(left+1, 10, "SECTORS", render.ColorDefault)
	set(left+1, 11, fmt.Sprint(g.Sectors()), render.ColorDefault)
	if b, ok := g.Bonus(); ok {
		if title := bonusTitles[b.Rows]; title != "" {
			set(left+1, 13, title, render.ColorYellow)
		}
		set(left+1, 14, fmt.Sprintf("+%d", b.Points), render.ColorYellow)
	}
}

// bonusTitles name clears of several rows at once.
var bonusTitles = map[int]string{2: "DOUBLE!", 3: "TRIPLE!", 4: "DEFRAG!"}
