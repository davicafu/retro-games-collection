package tcellscreen

import (
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/gdamore/tcell/v2"
)

type Screen struct{ screen tcell.Screen }

func New() (*Screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.Clear()
	return &Screen{screen: s}, nil
}

func (s *Screen) Close()                                { s.screen.Fini() }
func (s *Screen) Size() (int, int)                      { return s.screen.Size() }
func (s *Screen) Clear()                                { s.screen.Clear() }
func (s *Screen) SetCell(x, y int, cell render.Cell)   { s.screen.SetContent(x, y, cell.Rune, nil, tcell.StyleDefault) }
func (s *Screen) Present() error                        { s.screen.Show(); return nil }
func (s *Screen) Raw() tcell.Screen                     { return s.screen }

type simulationScreen struct {
	width    int
	height   int
	contents [][]rune
}

func newSimulationScreen(width, height int) *simulationScreen {
	rows := make([][]rune, height)
	for y := range rows {
		rows[y] = make([]rune, width)
		for x := range rows[y] {
			rows[y][x] = ' '
		}
	}
	return &simulationScreen{width: width, height: height, contents: rows}
}

func (s *simulationScreen) Size() (int, int) { return s.width, s.height }
func (s *simulationScreen) Clear() {
	for y := range s.contents {
		for x := range s.contents[y] {
			s.contents[y][x] = ' '
		}
	}
}
func (s *simulationScreen) SetCell(x, y int, cell render.Cell) {
	if x < 0 || y < 0 || x >= s.width || y >= s.height {
		return
	}
	s.contents[y][x] = cell.Rune
}
func (s *simulationScreen) Present() error                      { return nil }

func runeCell(r rune) render.Cell { return render.Cell{Rune: r} }

func NewSimulationScreenForTest(width, height int) *simulationScreen {
	return newSimulationScreen(width, height)
}

func (s *simulationScreen) Snapshot() [][]rune {
	clone := make([][]rune, len(s.contents))
	for y := range s.contents {
		clone[y] = append([]rune(nil), s.contents[y]...)
	}
	return clone
}
