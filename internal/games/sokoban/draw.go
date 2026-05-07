package sokoban

import (
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// Draw renders the game state to the screen in a single pass.
// Clears the screen and renders at (0,0).
func Draw(screen render.Screen, g *Game) {
	screen.Clear()
	DrawAt(screen, g, 0, 0)
}

// DrawAt renders the game state at an offset position on screen.
func DrawAt(screen render.Screen, g *Game, offsetX, offsetY int) {
	// Pre-compute box and target positions for quick lookup
	boxPositions := make(map[string]bool)
	for _, b := range g.Boxes {
		boxPositions[pointKey(b)] = true
	}

	targetPositions := make(map[string]bool)
	for _, t := range g.Targets {
		targetPositions[pointKey(t)] = true
	}

	width, height := screen.Size()

	// Render each cell within the game bounds
	for y := 0; y < height-offsetY; y++ {
		for x := 0; x < width-offsetX; x++ {
			p := grid.Point{X: x, Y: y}
			// Skip coordinates outside the game area
			if x >= g.Width || y >= g.Height {
				continue
			}

			r := ' ' // floor

			// Check for wall
			if g.isWall(p) {
				r = '#'
			}

			// Check for target
			if targetPositions[pointKey(p)] {
				if boxPositions[pointKey(p)] {
					r = '●'
				} else {
					r = '.'
				}
			} else if boxPositions[pointKey(p)] {
				r = '$'
			}

			// Check for player
			if g.Player == p {
				if targetPositions[pointKey(p)] {
					r = '+'
				} else {
					r = '@'
				}
			}

			screen.SetCell(offsetX+x, offsetY+y, render.Cell{Rune: r})
		}
	}
}

func pointKey(p grid.Point) string {
	return fmt.Sprintf("%d,%d", p.X, p.Y)
}
