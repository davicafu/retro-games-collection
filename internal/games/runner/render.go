package runner

import (
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

const (
	// skyHeight is the number of playable rows above the ground: the highest
	// jump (4) plus the player's head.
	skyHeight = 6

	// frameHeight is the HUD row plus the bordered box (sky rows + top border + ground line).
	frameHeight = 1 + skyHeight + 2

	glyphHead       = '●'
	glyphBody       = '█'
	glyphDuck       = '▄'
	glyphGround     = '▓'
	glyphFlying     = '▒'
	glyphGroundLine = '═'
)

// layout places the frame (HUD + box) centered on the screen, so overlays,
// which center on the screen too, sit on top of the playfield.
type layout struct {
	originX int // left border column
	originY int // HUD row; the box starts one row below
}

func newLayout(screenWidth, screenHeight, playfieldWidth int) layout {
	return layout{
		originX: max(0, (screenWidth-(playfieldWidth+2))/2),
		originY: max(0, (screenHeight-frameHeight)/2),
	}
}

// toScreen maps playfield column x and height h above the ground to screen
// coordinates. Height -1 is the ground line.
func (l layout) toScreen(x, h int) grid.Point {
	return grid.Point{X: l.originX + 1 + x, Y: l.originY + 2 + skyHeight - 1 - h}
}

// Draw renders the HUD, playfield, obstacles and player.
func Draw(screen render.Screen, g *Game) {
	width := g.Width()
	sw, sh := screen.Size()
	l := newLayout(sw, sh, width)

	hud := fmt.Sprintf("Score: %d   Speed x%d.%d", g.Score(), g.Speed()*10/BaseSpeed/10, g.Speed()*10/BaseSpeed%10)
	render.WriteText(screen, l.originX+1, l.originY, render.PadRight(hud, width))

	// The box's bottom edge doubles as the ground line.
	render.DrawBox(screen, l.originX, l.originY+1, width, skyHeight)
	ground := l.toScreen(0, -1)
	for x := 0; x < width; x++ {
		render.SetCell(screen, ground.X+x, ground.Y, glyphGroundLine)
	}

	for _, o := range g.Obstacles() {
		drawObstacle(screen, l, o, width)
	}
	drawPlayer(screen, l, g)
}

func drawObstacle(screen render.Screen, l layout, o Obstacle, width int) {
	glyph := glyphGround
	if o.Kind == KindFlying {
		glyph = glyphFlying
	}
	for x := o.X; x < o.X+o.Width; x++ {
		if x < 0 || x >= width {
			continue
		}
		for h := 0; h < skyHeight; h++ {
			if o.occupies(h) {
				p := l.toScreen(x, h)
				render.SetCell(screen, p.X, p.Y, glyph)
			}
		}
	}
}

func drawPlayer(screen render.Screen, l layout, g *Game) {
	if g.Ducking() {
		p := l.toScreen(g.PlayerX(), 0)
		render.SetCell(screen, p.X, p.Y, glyphDuck)
		return
	}
	feet := l.toScreen(g.PlayerX(), g.PlayerHeight())
	head := l.toScreen(g.PlayerX(), g.PlayerHeight()+1)
	render.SetCell(screen, feet.X, feet.Y, glyphBody)
	render.SetCell(screen, head.X, head.Y, glyphHead)
}
