package render

import (
	"strconv"
	"strings"
)

// BufferScreen implements Screen backed by an in-memory rune buffer.
// Snapshot() returns the full frame as a string for Bubble Tea rendering.
type BufferScreen struct {
	width    int
	height   int
	contents [][]rune
	colors   [][]Color
}

func NewBufferScreen(width, height int) *BufferScreen {
	rows := make([][]rune, height)
	colors := make([][]Color, height)
	for y := range rows {
		rows[y] = make([]rune, width)
		colors[y] = make([]Color, width)
		for x := range rows[y] {
			rows[y][x] = ' '
		}
	}
	return &BufferScreen{width: width, height: height, contents: rows, colors: colors}
}

func (s *BufferScreen) Size() (int, int) { return s.width, s.height }

func (s *BufferScreen) Clear() {
	for y := range s.contents {
		row := s.contents[y]
		for x := range row {
			row[x] = ' '
			s.colors[y][x] = ColorDefault
		}
	}
}

func (s *BufferScreen) SetCell(x, y int, cell Cell) {
	if x < 0 || y < 0 || x >= s.width || y >= s.height {
		return
	}
	s.contents[y][x] = cell.Rune
	if cell.Color > ColorWhite {
		cell.Color = ColorDefault // not in the palette
	}
	s.colors[y][x] = cell.Color
}

func (s *BufferScreen) Present() error { return nil }

// Snapshot returns a string of the buffer contents, rows joined by newlines.
func (s *BufferScreen) Snapshot() string {
	var sb strings.Builder
	sb.Grow(s.width * (s.height + 1))
	for y := range s.contents {
		row := s.contents[y]
		for x := range row {
			sb.WriteRune(row[x])
		}
		if y < s.height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// Rows returns a deep copy of the buffer for test assertions.
func (s *BufferScreen) Rows() [][]rune {
	clone := make([][]rune, s.height)
	for y := range s.contents {
		clone[y] = append([]rune(nil), s.contents[y]...)
	}
	return clone
}

// StyledSnapshot is Snapshot with ANSI escape codes around the runs of
// coloured cells; every row ends in the default colour.
func (s *BufferScreen) StyledSnapshot() string {
	var sb strings.Builder
	sb.Grow(s.width * (s.height + 1))
	for y := range s.contents {
		current := ColorDefault
		for x, r := range s.contents[y] {
			if c := s.colors[y][x]; c != current {
				sb.WriteString(sgr(c))
				current = c
			}
			sb.WriteRune(r)
		}
		if current != ColorDefault {
			sb.WriteString(sgr(ColorDefault))
		}
		if y < s.height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// sgr is the escape code that switches to a colour: a bright ANSI
// foreground, or a reset for the default.
func sgr(c Color) string {
	if c == ColorDefault {
		return "\x1b[0m"
	}
	return "\x1b[" + strconv.Itoa(90+int(c)) + "m"
}

// Colors returns a deep copy of the cell colours for test assertions.
func (s *BufferScreen) Colors() [][]Color {
	clone := make([][]Color, s.height)
	for y := range s.colors {
		clone[y] = append([]Color(nil), s.colors[y]...)
	}
	return clone
}
