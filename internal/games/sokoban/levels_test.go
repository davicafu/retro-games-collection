package sokoban

import (
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
