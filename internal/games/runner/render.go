package runner

import (
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

const (
	// skyHeight is the number of playable rows above the ground: the highest
	// double jump plus the player's body and head.
	skyHeight = MaxJumpHeight + 2

	glyphHead       = '●'
	glyphBody       = '█'
	glyphDuck       = '▄'
	glyphGround     = '▓'
	glyphFlying     = '▒'
	glyphGroundLine = '═'
)

// set draws a glyph at playfield column x and height h above the ground.
// Height -1 is the ground line, the box's bottom border.
func set(screen render.Screen, f render.Frame, x, h int, glyph rune) {
	sx, sy := f.Cell(x, skyHeight-1-h)
	render.SetCell(screen, sx, sy, glyph)
}

// Draw renders the HUD, playfield, obstacles and player.
func Draw(screen render.Screen, g *Game) {
	width := g.Width()
	f := render.CenterFrame(screen, width, skyHeight)
	f.Draw(screen, fmt.Sprintf("Score: %d   Speed x%d.%d", g.Score(), g.Speed()*10/BaseSpeed/10, g.Speed()*10/BaseSpeed%10))

	// The box's bottom edge doubles as the ground line.
	for x := 0; x < width; x++ {
		set(screen, f, x, -1, glyphGroundLine)
	}

	for _, o := range g.Obstacles() {
		drawObstacle(screen, f, o, width)
	}
	drawPlayer(screen, f, g)
}

func drawObstacle(screen render.Screen, f render.Frame, o Obstacle, width int) {
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
				set(screen, f, x, h, glyph)
			}
		}
	}
}

func drawPlayer(screen render.Screen, f render.Frame, g *Game) {
	if g.Ducking() {
		set(screen, f, g.PlayerX(), 0, glyphDuck)
		return
	}
	set(screen, f, g.PlayerX(), g.PlayerHeight(), glyphBody)
	set(screen, f, g.PlayerX(), g.PlayerHeight()+1, glyphHead)
}
