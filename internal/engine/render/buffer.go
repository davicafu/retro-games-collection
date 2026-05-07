package render

import "strings"

// BufferScreen implements Screen backed by an in-memory rune buffer.
// Snapshot() returns the full frame as a string for Bubble Tea rendering.
type BufferScreen struct {
	width    int
	height   int
	contents [][]rune
}

func NewBufferScreen(width, height int) *BufferScreen {
	rows := make([][]rune, height)
	for y := range rows {
		rows[y] = make([]rune, width)
		for x := range rows[y] {
			rows[y][x] = ' '
		}
	}
	return &BufferScreen{width: width, height: height, contents: rows}
}

func (s *BufferScreen) Size() (int, int) { return s.width, s.height }

func (s *BufferScreen) Clear() {
	for y := range s.contents {
		row := s.contents[y]
		for x := range row {
			row[x] = ' '
		}
	}
}

func (s *BufferScreen) SetCell(x, y int, cell Cell) {
	if x < 0 || y < 0 || x >= s.width || y >= s.height {
		return
	}
	s.contents[y][x] = cell.Rune
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
