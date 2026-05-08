package sokoban

import (
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	lines := []string{
		"#####",
		"#@  #",
		"# $ #",
		"#  .#",
		"#####",
	}

	game, err := ParseLevel(lines)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if game.Width != 5 {
		t.Errorf("expected width 5, got %d", game.Width)
	}
	if game.Height != 5 {
		t.Errorf("expected height 5, got %d", game.Height)
	}
	if game.Player.X != 1 || game.Player.Y != 1 {
		t.Errorf("expected player at (1,1), got (%d,%d)", game.Player.X, game.Player.Y)
	}
	if len(game.Boxes) != 1 || game.Boxes[0].X != 2 || game.Boxes[0].Y != 2 {
		t.Errorf("expected box at (2,2), got %v", game.Boxes)
	}
	if len(game.Targets) != 1 || game.Targets[0].X != 3 || game.Targets[0].Y != 3 {
		t.Errorf("expected target at (3,3), got %v", game.Targets)
	}
	if len(game.Walls) != 16 { // all # characters
		t.Errorf("expected 16 walls, got %d", len(game.Walls))
	}
}

func TestParseLevelInvalid(t *testing.T) {
	// Missing player - should error
	lines := []string{
		"#####",
		"#   #",
		"# $ #",
		"#  .#",
		"#####",
	}

	_, err := ParseLevel(lines)
	if err == nil {
		t.Error("expected error for level without player")
	}
}

func TestAllLevelsParse(t *testing.T) {
	for i, level := range Levels {
		game, err := ParseLevelString(level)
		if err != nil {
			t.Errorf("level %d parse error: %v", i+1, err)
			continue
		}
		if game == nil {
			t.Errorf("level %d: game is nil", i+1)
			continue
		}
		if len(game.Boxes) != len(game.Targets) {
			t.Errorf("level %d: %d boxes but %d targets (must match)", i+1, len(game.Boxes), len(game.Targets))
		}
		if len(game.Boxes) == 0 {
			t.Errorf("level %d: no boxes", i+1)
		}
	}
}

func TestParseLevelMultiBoxTarget(t *testing.T) {
	lines := []string{
		"#######",
		"#  @  #",
		"# $  #",
		"# .$. #",
		"#  .  #",
		"#######",
	}

	game, err := ParseLevel(lines)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(game.Boxes) != 2 {
		t.Errorf("expected 2 boxes, got %d", len(game.Boxes))
	}
	if len(game.Targets) != 3 {
		t.Errorf("expected 3 targets, got %d", len(game.Targets))
	}
}

func TestAllLevelsAreRectangularAndPerimeterClosed(t *testing.T) {
	for i, level := range Levels {
		lines := strings.Split(strings.TrimSpace(level), "\n")
		if len(lines) == 0 {
			t.Fatalf("level %d: expected at least one row", i+1)
		}

		width := len(lines[0])
		for rowIndex, line := range lines {
			if len(line) != width {
				t.Fatalf("level %d row %d: expected width %d, got %d", i+1, rowIndex, width, len(line))
			}
		}

		for x := 0; x < width; x++ {
			if lines[0][x] != '#' {
				t.Fatalf("level %d: top border must be closed at x=%d", i+1, x)
			}
			if lines[len(lines)-1][x] != '#' {
				t.Fatalf("level %d: bottom border must be closed at x=%d", i+1, x)
			}
		}

		for y, line := range lines {
			if line[0] != '#' {
				t.Fatalf("level %d: left border must be closed at y=%d", i+1, y)
			}
			if line[len(line)-1] != '#' {
				t.Fatalf("level %d: right border must be closed at y=%d", i+1, y)
			}
		}
	}
}

func TestLevel7HasNoBoxesPinnedUnderTopWall(t *testing.T) {
	lines := strings.Split(strings.TrimSpace(Levels[6]), "\n")
	for x := 1; x < len(lines[1])-1; x++ {
		if lines[1][x] == '$' {
			t.Fatalf("level 7: box at top interior row x=%d is pinned under the top wall", x)
		}
	}
}
