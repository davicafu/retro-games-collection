package snake

import (
	"fmt"
	"strings"

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

func DrawOverlay(screen render.Screen, overlay *OverlayModel) {
	layout := defaultPlayfieldLayout(12, 12)
	if !shouldUseProjectedLayout(screen, layout) {
		return
	}

	boxX, boxY := 2, 4
	boxW := 10

	titleLines := strings.SplitN(overlay.Title(), "\n", 2)
	numHeader := len(titleLines)
	numOptions := len(overlay.Options())
	contentHeight := numHeader + 1 + numOptions // header + separator + options
	boxHeight := contentHeight + 2              // + top/bottom borders

	drawBoxSized(screen, boxX, boxY, boxW, boxHeight)

	for i, line := range titleLines {
		writeStr(screen, boxX+1, boxY+1+i, padRight(line, boxW))
	}

	for i, opt := range overlay.Options() {
		marker := " "
		if i == overlay.SelectedIndex() {
			marker = ">"
		}
		text := fmt.Sprintf("%s %s", marker, opt.Label)
		writeStr(screen, boxX+1, boxY+1+numHeader+1+i, padRight(text, boxW))
	}
}

func drawBoxSized(screen render.Screen, x, y, width, height int) {
	setRune(screen, x, y, '┌')
	for i := 1; i <= width; i++ {
		setRune(screen, x+i, y, '─')
	}
	setRune(screen, x+width+1, y, '┐')

	bottomY := y + height - 1
	for row := y + 1; row < bottomY; row++ {
		setRune(screen, x, row, '│')
		setRune(screen, x+width+1, row, '│')
	}

	setRune(screen, x, bottomY, '└')
	for i := 1; i <= width; i++ {
		setRune(screen, x+i, bottomY, '─')
	}
	setRune(screen, x+width+1, bottomY, '┘')
}

func setRune(screen render.Screen, x, y int, r rune) {
	screen.SetCell(x, y, render.Cell{Rune: r})
}

func writeStr(screen render.Screen, x, y int, text string) {
	for i, r := range text {
		screen.SetCell(x+i, y, render.Cell{Rune: r})
	}
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	padded := make([]byte, n)
	for i := range padded {
		padded[i] = ' '
	}
	copy(padded, s)
	return string(padded)
}
