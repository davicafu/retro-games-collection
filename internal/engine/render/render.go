package render

// Cell is one character on screen, in a colour.
type Cell struct {
	Rune  rune
	Color Color
}

// Color is a foreground colour from the terminal's bright ANSI palette; the
// zero Color is the terminal's default.
type Color uint8

const (
	ColorDefault Color = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

type Screen interface {
	Size() (int, int)
	Clear()
	SetCell(x, y int, cell Cell)
	Present() error
}
