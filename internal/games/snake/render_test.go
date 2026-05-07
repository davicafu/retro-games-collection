package snake

import (
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/tcellscreen"
)

func TestDraw_RendersBottomRightCellInsidePlayfield(t *testing.T) {
	screen := tcellscreen.NewSimulationScreenForTest(14, 15)
	Draw(screen, []Position{{X: 11, Y: 11}}, nil)

	snapshot := screen.Snapshot()
	if snapshot[13][12] != '█' {
		t.Fatalf("expected bottom-right logical cell inside playfield at (12,13), got %q", snapshot[13][12])
	}
	if snapshot[13][13] != '|' {
		t.Fatalf("expected right border outside playfield, got %q", snapshot[13][13])
	}
}

func TestPlayfieldLayout_ProjectsLogicalCellsIntoInterior(t *testing.T) {
	layout := defaultPlayfieldLayout(12, 12)

	topLeft := layout.toScreen(Position{X: 0, Y: 0})
	if topLeft != (Position{X: 1, Y: 2}) {
		t.Fatalf("expected logical (0,0) -> screen (1,2), got %+v", topLeft)
	}

	bottomRight := layout.toScreen(Position{X: 11, Y: 11})
	if bottomRight != (Position{X: 12, Y: 13}) {
		t.Fatalf("expected logical (11,11) -> screen (12,13), got %+v", bottomRight)
	}
}

func TestPlayfieldLayout_TotalSizeIncludesBorderAndTopMargin(t *testing.T) {
	layout := defaultPlayfieldLayout(12, 12)

	if got := layout.totalWidth(); got != 14 {
		t.Fatalf("expected total width 14, got %d", got)
	}

	if got := layout.totalHeight(); got != 15 {
		t.Fatalf("expected total height 15, got %d", got)
	}
}

func TestDraw_PreservesCurrentRuntimeCoordinateContract(t *testing.T) {
	screen := &stubScreen{width: 12, height: 12, cells: map[Position]rune{}}

	Draw(screen, []Position{{X: 1, Y: 3}}, []Position{{X: 5, Y: 7}})

	if got := screen.cells[Position{X: 1, Y: 3}]; got != '█' {
		t.Fatalf("expected snake segment rendered at logical coordinates, got %q", got)
	}

	if got := screen.cells[Position{X: 5, Y: 7}]; got != '•' {
		t.Fatalf("expected food rendered at logical coordinates, got %q", got)
	}
}

func TestDraw_ProjectsCoordinatesWhenSurfaceFitsFullLayout(t *testing.T) {
	screen := &stubScreen{width: 14, height: 15, cells: map[Position]rune{}}

	Draw(screen, []Position{{X: 1, Y: 3}}, []Position{{X: 5, Y: 7}})

	if got := screen.cells[Position{X: 2, Y: 5}]; got != '█' {
		t.Fatalf("expected projected snake segment at (2,5), got %q", got)
	}

	if got := screen.cells[Position{X: 6, Y: 9}]; got != '•' {
		t.Fatalf("expected projected food at (6,9), got %q", got)
	}
}

type stubScreen struct {
	width  int
	height int
	cells map[Position]rune
}

func (s *stubScreen) Size() (int, int) { return s.width, s.height }
func (s *stubScreen) Clear()           {}
func (s *stubScreen) Present() error   { return nil }
func (s *stubScreen) SetCell(x, y int, cell render.Cell) {
	s.cells[Position{X: x, Y: y}] = cell.Rune
}
