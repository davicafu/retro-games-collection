package records

import (
	"encoding/json"
	"sync"
)

type savesFile struct {
	Version int                        `json:"version"`
	Games   map[string]json.RawMessage `json:"games"`
}

// Saves holds one saved run per game, in a shape each game decides. It is
// safe for concurrent use.
type Saves struct {
	mu   sync.Mutex
	path string
	file savesFile
}

// OpenSaves reads the saves at path, as Open reads a book. An empty path
// keeps them in memory only.
func OpenSaves(path string) (*Saves, error) {
	s := &Saves{path: path}
	if path != "" {
		f, err := readJSON(path, func(f savesFile) int { return f.Version })
		if err != nil {
			return nil, err
		}
		s.file = f
	}
	s.file.Version = FormatVersion
	if s.file.Games == nil {
		s.file.Games = map[string]json.RawMessage{}
	}
	return s, nil
}

func (s *Saves) save() error {
	if s.path == "" {
		return nil
	}
	return writeJSON(s.path, s.file)
}

// Slot is one game's saved run.
type Slot struct {
	saves *Saves
	game  string
}

// Slot binds the saves to game.
func (s *Saves) Slot(game string) Slot { return Slot{saves: s, game: game} }

// Exists reports whether the game has a saved run.
func (sl Slot) Exists() bool {
	sl.saves.mu.Lock()
	defer sl.saves.mu.Unlock()
	_, ok := sl.saves.file.Games[sl.game]
	return ok
}

// Load fills v with the saved run and reports whether there was one. A run
// that does not decode into v counts as none.
func (sl Slot) Load(v any) (bool, error) {
	sl.saves.mu.Lock()
	defer sl.saves.mu.Unlock()
	raw, ok := sl.saves.file.Games[sl.game]
	if !ok {
		return false, nil
	}
	if json.Unmarshal(raw, v) != nil {
		return false, nil
	}
	return true, nil
}

// Store saves v as the game's run, replacing any other.
func (sl Slot) Store(v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	sl.saves.mu.Lock()
	defer sl.saves.mu.Unlock()
	sl.saves.file.Games[sl.game] = raw
	return sl.saves.save()
}

// Clear deletes the game's run.
func (sl Slot) Clear() error {
	sl.saves.mu.Lock()
	defer sl.saves.mu.Unlock()
	if _, ok := sl.saves.file.Games[sl.game]; !ok {
		return nil
	}
	delete(sl.saves.file.Games, sl.game)
	return sl.saves.save()
}
