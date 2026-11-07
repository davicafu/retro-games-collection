package snake

import (
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

type playfieldLayout struct {
	boardWidth   int
	boardHeight  int
	boardOriginX int
	boardOriginY int
}

func defaultPlayfieldLayout(boardWidth, boardHeight int) playfieldLayout {
	return playfieldLayout{
		boardWidth:   boardWidth,
		boardHeight:  boardHeight,
		boardOriginX: 1,
		boardOriginY: 2,
	}
}

func (l playfieldLayout) toScreen(p Position) Position {
	return Position{X: l.boardOriginX + p.X, Y: l.boardOriginY + p.Y}
}

func (l playfieldLayout) totalWidth() int  { return l.boardWidth + 2 }
func (l playfieldLayout) totalHeight() int { return l.boardHeight + 3 }

func Draw(screen render.Screen, body []Position, food []Position) {
	layout := defaultPlayfieldLayout(12, 12)
	if shouldUseProjectedLayout(screen, layout) {
		drawBorder(screen, layout)

		for _, segment := range body {
			p := layout.toScreen(segment)
			screen.SetCell(p.X, p.Y, render.Cell{Rune: '█'})
		}

		for _, item := range food {
			p := layout.toScreen(item)
			screen.SetCell(p.X, p.Y, render.Cell{Rune: '•'})
		}

		return
	}

	drawLegacy(screen, body, food)
}

func shouldUseProjectedLayout(screen render.Screen, layout playfieldLayout) bool {
	width, height := screen.Size()
	return width >= layout.totalWidth() && height >= layout.totalHeight()
}

func drawLegacy(screen render.Screen, body []Position, food []Position) {
	for _, segment := range body {
		screen.SetCell(segment.X, segment.Y, render.Cell{Rune: '█'})
	}

	for _, item := range food {
		screen.SetCell(item.X, item.Y, render.Cell{Rune: '•'})
	}
}

func drawBorder(screen render.Screen, layout playfieldLayout) {
	width, height := layout.totalWidth(), layout.totalHeight()
	for x := 0; x < width; x++ {
		top := '-'
		bottom := '-'
		if x == 0 || x == width-1 {
			top = '+'
			bottom = '+'
		}
		screen.SetCell(x, 1, render.Cell{Rune: top})
		screen.SetCell(x, height-1, render.Cell{Rune: bottom})
	}
	for y := 2; y < height-1; y++ {
		screen.SetCell(0, y, render.Cell{Rune: '|'})
		screen.SetCell(width-1, y, render.Cell{Rune: '|'})
	}
}
