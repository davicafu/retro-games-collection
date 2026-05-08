package sokoban

import (
	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
)

type Game struct {
	Width     int
	Height    int
	Player    grid.Point
	Walls     []grid.Point
	Boxes     []grid.Point
	Targets   []grid.Point
	undoStack *undoStack
}

func NewGame(width, height int, player grid.Point, walls, boxes, targets []grid.Point) *Game {
	return &Game{
		Width:     width,
		Height:    height,
		Player:    player,
		Walls:     append([]grid.Point(nil), walls...),
		Boxes:     append([]grid.Point(nil), boxes...),
		Targets:   append([]grid.Point(nil), targets...),
		undoStack: newUndoStack(),
	}
}

func (g *Game) initUndoStack() {
	if g.undoStack == nil {
		g.undoStack = newUndoStack()
	}
}

func (g *Game) Move(action input.Action) bool {
	g.initUndoStack()
	dx, dy := actionToDir(action)
	if dx == 0 && dy == 0 {
		return false
	}

	newPlayer := grid.Point{X: g.Player.X + dx, Y: g.Player.Y + dy}

	// Check wall collision
	if g.isWall(newPlayer) {
		return false
	}

	// Check if there's a box to push
	boxIdx := g.boxAt(newPlayer)
	if boxIdx >= 0 {
		newBox := grid.Point{X: newPlayer.X + dx, Y: newPlayer.Y + dy}

		// Can't push into wall
		if g.isWall(newBox) {
			return false
		}

		// Can't push into another box
		if g.boxAt(newBox) >= 0 {
			return false
		}

		// Save state for undo
		g.undoStack.Push(undoEntry{
			player: g.Player,
			boxes:  append([]grid.Point(nil), g.Boxes...),
		})

		// Push the box
		g.Boxes[boxIdx] = newBox
		g.Player = newPlayer
		return true
	}

	// Just move player
	g.undoStack.Push(undoEntry{
		player: g.Player,
		boxes:  append([]grid.Point(nil), g.Boxes...),
	})
	g.Player = newPlayer
	return true
}

func (g *Game) Undo() bool {
	g.initUndoStack()
	entry := g.undoStack.Pop()
	if entry == nil {
		return false
	}
	g.Player = entry.player
	g.Boxes = entry.boxes
	return true
}

func (g *Game) Won() bool {
	for _, target := range g.Targets {
		if g.boxAt(target) < 0 {
			return false
		}
	}
	return len(g.Targets) > 0
}

func (g *Game) isWall(p grid.Point) bool {
	for _, w := range g.Walls {
		if w == p {
			return true
		}
	}
	return false
}

func (g *Game) boxAt(p grid.Point) int {
	for i, b := range g.Boxes {
		if b == p {
			return i
		}
	}
	return -1
}

func actionToDir(action input.Action) (dx, dy int) {
	switch action {
	case input.ActionUp:
		return 0, -1
	case input.ActionDown:
		return 0, 1
	case input.ActionLeft:
		return -1, 0
	case input.ActionRight:
		return 1, 0
	}
	return 0, 0
}
