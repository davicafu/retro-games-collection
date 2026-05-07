package render

import (
	"strings"
	"testing"
)

func TestNewBufferScreen_CreatesCorrectDimensions(t *testing.T) {
	s := NewBufferScreen(10, 5)
	if s.width != 10 {
		t.Fatalf("expected width 10, got %d", s.width)
	}
	if s.height != 5 {
		t.Fatalf("expected height 5, got %d", s.height)
	}
	if len(s.contents) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(s.contents))
	}
	for i, row := range s.contents {
		if len(row) != 10 {
			t.Fatalf("expected row %d to have length 10, got %d", i, len(row))
		}
	}
}

func TestNewBufferScreen_InitializedToSpaces(t *testing.T) {
	s := NewBufferScreen(4, 3)
	for y := 0; y < 3; y++ {
		for x := 0; x < 4; x++ {
			if s.contents[y][x] != ' ' {
				t.Fatalf("expected space at (%d,%d), got %q", x, y, s.contents[y][x])
			}
		}
	}
}

func TestBufferScreen_Size(t *testing.T) {
	s := NewBufferScreen(20, 10)
	w, h := s.Size()
	if w != 20 {
		t.Fatalf("expected width 20, got %d", w)
	}
	if h != 10 {
		t.Fatalf("expected height 10, got %d", h)
	}
}

func TestBufferScreen_SetAndReadCell(t *testing.T) {
	s := NewBufferScreen(10, 5)
	s.SetCell(3, 2, Cell{Rune: 'X'})
	if s.contents[2][3] != 'X' {
		t.Fatalf("expected 'X' at (3,2), got %q", s.contents[2][3])
	}
}

func TestBufferScreen_SetCellOutOfBoundsIgnored(t *testing.T) {
	s := NewBufferScreen(10, 5)
	s.SetCell(-1, 2, Cell{Rune: 'X'})
	s.SetCell(10, 2, Cell{Rune: 'X'})
	s.SetCell(5, -1, Cell{Rune: 'X'})
	s.SetCell(5, 5, Cell{Rune: 'X'})

	// Verify all cells are still spaces
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			if s.contents[y][x] != ' ' {
				t.Fatalf("expected space at (%d,%d) after out-of-bounds writes, got %q", x, y, s.contents[y][x])
			}
		}
	}
}

func TestBufferScreen_ClearResetsToSpaces(t *testing.T) {
	s := NewBufferScreen(5, 3)
	s.SetCell(0, 0, Cell{Rune: 'A'})
	s.SetCell(4, 2, Cell{Rune: 'Z'})
	s.Clear()

	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			if s.contents[y][x] != ' ' {
				t.Fatalf("expected space at (%d,%d) after Clear, got %q", x, y, s.contents[y][x])
			}
		}
	}
}

func TestBufferScreen_PresentReturnsNil(t *testing.T) {
	s := NewBufferScreen(5, 3)
	if err := s.Present(); err != nil {
		t.Fatalf("expected Present to return nil, got %v", err)
	}
}

func TestBufferScreen_SnapshotEmpty(t *testing.T) {
	s := NewBufferScreen(4, 3)
	expected := "    \n    \n    "
	if got := s.Snapshot(); got != expected {
		t.Fatalf("expected empty snapshot %q, got %q", expected, got)
	}
}

func TestBufferScreen_SnapshotWithContent(t *testing.T) {
	s := NewBufferScreen(5, 3)
	s.SetCell(0, 0, Cell{Rune: 'A'})
	s.SetCell(4, 1, Cell{Rune: 'B'})
	s.SetCell(2, 2, Cell{Rune: 'C'})

	snapshot := s.Snapshot()
	lines := strings.Split(snapshot, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "A    " {
		t.Fatalf("expected first line 'A    ', got %q", lines[0])
	}
	if string(lines[1][4]) != "B" {
		t.Fatalf("expected 'B' at column 4 of line 2, got %q", lines[1][4])
	}
	if string(lines[2][2]) != "C" {
		t.Fatalf("expected 'C' at column 2 of line 3, got %q", lines[2][2])
	}
}

func TestBufferScreen_RowsReturnsDeepCopy(t *testing.T) {
	s := NewBufferScreen(3, 2)
	s.SetCell(0, 0, Cell{Rune: 'X'})

	rows := s.Rows()
	// Mutate the returned copy
	rows[0][0] = 'Z'

	// Original should be unchanged
	if s.contents[0][0] != 'X' {
		t.Fatalf("expected original to remain 'X', got %q", s.contents[0][0])
	}
}

func TestBufferScreen_RowsIsolation(t *testing.T) {
	s := NewBufferScreen(3, 2)
	rows := s.Rows()
	rows[1][1] = 'Y'

	if s.contents[1][1] != ' ' {
		t.Fatalf("expected original to be space after rows mutation, got %q", s.contents[1][1])
	}
}
