package sokoban

import (
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
)

func TestMovePlayer(t *testing.T) {
	game := &Game{
		Width:   5,
		Height:  5,
		Player:  grid.Point{X: 2, Y: 2},
		Walls:   []grid.Point{},
		Boxes:   []grid.Point{},
		Targets: []grid.Point{},
	}

	// Move up
	moved := game.Move(input.ActionUp)
	if !moved {
		t.Error("expected move to succeed")
	}
	if game.Player != (grid.Point{X: 2, Y: 1}) {
		t.Errorf("expected player at (2,1), got %v", game.Player)
	}
}

func TestMoveIntoWall(t *testing.T) {
	game := &Game{
		Width:   5,
		Height:  5,
		Player:  grid.Point{X: 2, Y: 2},
		Walls:   []grid.Point{{X: 2, Y: 1}},
		Boxes:   []grid.Point{},
		Targets: []grid.Point{},
	}

	moved := game.Move(input.ActionUp)
	if moved {
		t.Error("expected move into wall to fail")
	}
	if game.Player != (grid.Point{X: 2, Y: 2}) {
		t.Errorf("expected player unchanged, got %v", game.Player)
	}
}

func TestPushBox(t *testing.T) {
	game := &Game{
		Width:   5,
		Height:  5,
		Player:  grid.Point{X: 2, Y: 2},
		Walls:   []grid.Point{},
		Boxes:   []grid.Point{{X: 2, Y: 1}},
		Targets: []grid.Point{},
	}

	// Push box up
	moved := game.Move(input.ActionUp)
	if !moved {
		t.Error("expected box push to succeed")
	}
	if game.Player != (grid.Point{X: 2, Y: 1}) {
		t.Errorf("expected player at (2,1), got %v", game.Player)
	}
	if len(game.Boxes) != 1 || game.Boxes[0] != (grid.Point{X: 2, Y: 0}) {
		t.Errorf("expected box at (2,0), got %v", game.Boxes)
	}
}

func TestPushBoxIntoWall(t *testing.T) {
	game := &Game{
		Width:   5,
		Height:  5,
		Player:  grid.Point{X: 2, Y: 2},
		Walls:   []grid.Point{{X: 2, Y: 0}},
		Boxes:   []grid.Point{{X: 2, Y: 1}},
		Targets: []grid.Point{},
	}

	moved := game.Move(input.ActionUp)
	if moved {
		t.Error("expected box push into wall to fail")
	}
	if game.Player != (grid.Point{X: 2, Y: 2}) {
		t.Errorf("expected player unchanged, got %v", game.Player)
	}
}

func TestPushBoxIntoBox(t *testing.T) {
	game := &Game{
		Width:   5,
		Height:  5,
		Player:  grid.Point{X: 2, Y: 2},
		Walls:   []grid.Point{},
		Boxes:   []grid.Point{{X: 2, Y: 1}, {X: 2, Y: 0}},
		Targets: []grid.Point{},
	}

	moved := game.Move(input.ActionUp)
	if moved {
		t.Error("expected box push into box to fail")
	}
	if game.Player != (grid.Point{X: 2, Y: 2}) {
		t.Errorf("expected player unchanged, got %v", game.Player)
	}
}

func TestWon(t *testing.T) {
	game := &Game{
		Width:   5,
		Height:  5,
		Player:  grid.Point{X: 1, Y: 2},
		Walls:   []grid.Point{},
		Boxes:   []grid.Point{{X: 2, Y: 2}},
		Targets: []grid.Point{{X: 2, Y: 2}},
	}

	if !game.Won() {
		t.Error("expected game to be won when box on target")
	}

	// Move box off target
	game.Boxes[0] = grid.Point{X: 3, Y: 2}
	if game.Won() {
		t.Error("expected game not won when box off target")
	}
}

func TestUndo(t *testing.T) {
	game := &Game{
		Width:   5,
		Height:  5,
		Player:  grid.Point{X: 2, Y: 2},
		Walls:   []grid.Point{},
		Boxes:   []grid.Point{{X: 2, Y: 1}},
		Targets: []grid.Point{},
	}

	game.Move(input.ActionUp) // Push box to (2,0), player to (2,1)
	game.Undo()

	if game.Player != (grid.Point{X: 2, Y: 2}) {
		t.Errorf("expected player back at (2,2), got %v", game.Player)
	}
	if len(game.Boxes) != 1 || game.Boxes[0] != (grid.Point{X: 2, Y: 1}) {
		t.Errorf("expected box back at (2,1), got %v", game.Boxes)
	}
}

func TestUndoEmptyStack(t *testing.T) {
	game := &Game{
		Width:   5,
		Height:  5,
		Player:  grid.Point{X: 2, Y: 2},
		Walls:   []grid.Point{},
		Boxes:   []grid.Point{},
		Targets: []grid.Point{},
	}

	// Should not panic
	game.Undo()
}
