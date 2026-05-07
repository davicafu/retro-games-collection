package render

// DrawBox draws a bordered box at (x, y) with the given width and height.
// width and height are the INTERIOR dimensions (excluding borders).
func DrawBox(screen Screen, x, y, width, height int) {
	// Corners
	SetCell(screen, x, y, '┌')
	SetCell(screen, x+width+1, y, '┐')
	SetCell(screen, x, y+height+1, '└')
	SetCell(screen, x+width+1, y+height+1, '┘')

	// Top and bottom edges
	for i := 1; i <= width; i++ {
		SetCell(screen, x+i, y, '─')
		SetCell(screen, x+i, y+height+1, '─')
	}

	// Left and right edges
	for row := 1; row <= height; row++ {
		SetCell(screen, x, y+row, '│')
		SetCell(screen, x+width+1, y+row, '│')
	}
}

// WriteText writes a string starting at position (x, y) on the screen.
func WriteText(screen Screen, x, y int, text string) {
	for i, r := range text {
		SetCell(screen, x+i, y, r)
	}
}

// PadRight pads a string with spaces to reach exactly n visible characters.
// Truncates if the string is longer than n.
func PadRight(s string, n int) string {
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

// FillRect fills a rectangular area with the given rune.
// x, y is the top-left corner; width and height are the dimensions.
func FillRect(screen Screen, x, y, width, height int, r rune) {
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			SetCell(screen, x+col, y+row, r)
		}
	}
}

// SetCell is a convenience wrapper that writes a rune to the screen.
func SetCell(screen Screen, x, y int, r rune) {
	screen.SetCell(x, y, Cell{Rune: r})
}
