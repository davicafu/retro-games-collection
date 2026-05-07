package sokoban

import "github.com/davidbarreiro/cli-games-collection/internal/engine/grid"

const maxUndo = 50

type undoEntry struct {
	player grid.Point
	boxes  []grid.Point
}

type undoStack struct {
	entries []undoEntry
}

func newUndoStack() *undoStack {
	return &undoStack{
		entries: make([]undoEntry, 0, maxUndo),
	}
}

func (s *undoStack) Push(entry undoEntry) {
	s.entries = append(s.entries, entry)
	if len(s.entries) > maxUndo {
		s.entries = s.entries[1:]
	}
}

func (s *undoStack) Pop() *undoEntry {
	if len(s.entries) == 0 {
		return nil
	}
	idx := len(s.entries) - 1
	entry := s.entries[idx]
	s.entries = s.entries[:idx]
	return &entry
}

func (s *undoStack) Len() int {
	return len(s.entries)
}
