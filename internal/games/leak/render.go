package leak

import (
	"fmt"
	"strings"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// Every map cell is drawn 2 characters wide so the map looks square. The
// map's outer ring of walls is not drawn: the box around the playfield
// stands for it.
const (
	spriteLeak    = "◖◗"
	spriteFleeing = "░░"
	spriteWearing = "▒▒" // a fleeing collector whose malloc is about to wear off

	glyphByte   = '·'
	glyphMalloc = '◆'
	glyphGate   = '╌'
	glyphLife   = '♥'

	bannerReady = "READY"

	// BlinkTicks is how long before a malloc wears off the fleeing
	// collectors start blinking: 2 s, flipping every BlinkPeriod ticks.
	BlinkTicks  = 120
	BlinkPeriod = 15
)

var spriteCollector = [...]string{
	KindTracer:      "▛▜",
	KindInterceptor: "▙▟",
	KindWanderer:    "▞▚",
}

// Colours: the leak is yellow and every collector kind has its own colour;
// fleeing collectors share one, and turn white while they blink.
const (
	colorLeak    = render.ColorYellow
	colorFleeing = render.ColorBlue
	colorWearing = render.ColorWhite
)

var colorCollector = [...]render.Color{
	KindTracer:      render.ColorRed,
	KindInterceptor: render.ColorMagenta,
	KindWanderer:    render.ColorGreen,
}

// wallGlyphs maps the walls a wall cell connects to (up 1, left 2, down 4,
// right 8) to a box-drawing character.
var wallGlyphs = [16]rune{
	0: '▪',
	1: '│', 4: '│', 1 | 4: '│',
	2: '─', 8: '─', 2 | 8: '─',
	4 | 8: '┌', 2 | 4: '┐', 1 | 8: '└', 1 | 2: '┘',
	1 | 4 | 8: '├', 1 | 2 | 4: '┤', 2 | 4 | 8: '┬', 1 | 2 | 8: '┴',
	1 | 2 | 4 | 8: '┼',
}

// Draw renders the HUD, the map, the READY banner and the actors.
func Draw(screen render.Screen, g *Game) {
	f := render.CenterFrame(screen, 2*(g.Width()-2), g.Height()-2)
	f.Draw(screen, fmt.Sprintf("Level %d   Score: %d   %s", g.Level(), g.Score(), strings.Repeat(string(glyphLife), g.Lives())))

	for y := 1; y < g.Height()-1; y++ {
		for x := 1; x < g.Width()-1; x++ {
			drawTile(screen, f, g, grid.Point{X: x, Y: y})
		}
	}
	if !g.Started() {
		row := g.bannerRow()
		x, y := f.Cell((2*(g.Width()-2)-len(bannerReady))/2, row-1)
		render.WriteText(screen, x, y, bannerReady)
	}
	for _, c := range g.Collectors() {
		sprite, color := collectorSprite(g, c)
		drawActor(screen, f, c.Actor, sprite, color)
	}
	drawActor(screen, f, g.Leak(), spriteLeak, colorLeak)
}

func drawTile(screen render.Screen, f render.Frame, g *Game, p grid.Point) {
	x, y := f.Cell(2*(p.X-1), p.Y-1)
	switch g.Tile(p) {
	case TileByte:
		render.SetCell(screen, x, y, glyphByte)
	case TileMalloc:
		render.SetCell(screen, x, y, glyphMalloc)
	case TileGate:
		render.WriteText(screen, x, y, string([]rune{glyphGate, glyphGate}))
	case TileWall:
		if solid(g, p) {
			return
		}
		mask := 0
		for i, d := range []grid.Point{Up, Left, Down, Right} {
			if n := p.Add(d); drawnAsWall(g, n) && !solid(g, n) {
				mask |= 1 << i
			}
		}
		render.SetCell(screen, x, y, wallGlyphs[mask])
		if mask&8 != 0 {
			render.SetCell(screen, x+1, y, wallGlyphs[2|8])
		}
	}
}

// drawnAsWall reports a wall or gate cell; outside the map counts as wall.
func drawnAsWall(g *Game, p grid.Point) bool {
	t := g.Tile(p)
	return t == TileWall || t == TileGate
}

// solid reports a wall cell surrounded by walls on all 8 sides: the inside
// of a block, left blank so only its outline shows.
func solid(g *Game, p grid.Point) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if !drawnAsWall(g, grid.Point{X: p.X + dx, Y: p.Y + dy}) {
				return false
			}
		}
	}
	return true
}

func collectorSprite(g *Game, c Collector) (string, render.Color) {
	if !c.Fleeing() {
		return spriteCollector[c.Kind], colorCollector[c.Kind]
	}
	if left := g.FrightLeft(); left <= BlinkTicks && left/BlinkPeriod%2 == 0 {
		return spriteWearing, colorWearing
	}
	return spriteFleeing, colorFleeing
}

// drawActor draws a sprite where the actor is: horizontal moves in
// half-cell (one character) steps, vertical ones on the nearer row.
func drawActor(screen render.Screen, f render.Frame, a Actor, sprite string, color render.Color) {
	col := 2*(a.Cell.X-1) + a.Dir.X*(2*a.Progress/CellUnits)
	row := a.Cell.Y - 1
	if a.Dir.Y != 0 && a.Progress >= CellUnits/2 {
		row += a.Dir.Y
	}
	x, y := f.Cell(col, row)
	render.WriteColoredText(screen, x, y, sprite, color)
}
