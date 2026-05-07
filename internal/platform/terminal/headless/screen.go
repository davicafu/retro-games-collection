package headless

import "github.com/davidbarreiro/cli-games-collection/internal/engine/render"

type Screen struct {
	width  int
	height int
	Cells  map[[2]int]render.Cell
}

func NewScreen(width, height int) *Screen {
	return &Screen{width: width, height: height, Cells: map[[2]int]render.Cell{}}
}

func (s *Screen) Size() (int, int) { return s.width, s.height }
func (s *Screen) Clear()           { s.Cells = map[[2]int]render.Cell{} }
func (s *Screen) SetCell(x, y int, cell render.Cell) {
	s.Cells[[2]int{x, y}] = cell
}
func (s *Screen) Present() error { return nil }
