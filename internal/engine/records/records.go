// Package records keeps what outlives a playthrough: each game's high score
// table (Book) and saved runs (Saves), in JSON files under the user's
// config directory.
package records

import (
	"slices"
	"strings"
	"sync"
	"unicode"
)

const (
	// Size is how many results a table keeps.
	Size = 5
	// MaxName is the longest name a table keeps, in characters.
	MaxName = 12
	// DefaultName stands for a player who typed no name.
	DefaultName = "ANON"
)

// Entry is one result. A higher Score ranks first; on equal scores less
// time (Seconds) does, for games that measure it, and then the older
// result.
type Entry struct {
	Name    string `json:"name"`
	Score   int    `json:"score"`
	Seconds int    `json:"seconds,omitempty"`
}

func (e Entry) beats(o Entry) bool {
	if e.Score != o.Score {
		return e.Score > o.Score
	}
	return e.Seconds < o.Seconds
}

// CleanName is name as a table keeps it: without control characters or
// surrounding spaces, cut to MaxName, and DefaultName when nothing is left.
func CleanName(name string) string {
	name = strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name))
	if r := []rune(name); len(r) > MaxName {
		name = string(r[:MaxName])
	}
	if name == "" {
		return DefaultName
	}
	return name
}

type bookFile struct {
	Version int                `json:"version"`
	Games   map[string][]Entry `json:"games"`
}

// Book is every game's table. It is safe for concurrent use.
type Book struct {
	mu   sync.Mutex
	path string
	file bookFile
}

// Open reads the book at path; see readJSON for missing and unreadable
// files. An empty path keeps the book in memory only.
func Open(path string) (*Book, error) {
	b := &Book{path: path}
	if path != "" {
		f, err := readJSON(path, func(f bookFile) int { return f.Version })
		if err != nil {
			return nil, err
		}
		b.file = f
	}
	b.file.Version = FormatVersion
	if b.file.Games == nil {
		b.file.Games = map[string][]Entry{}
	}
	return b, nil
}

// Top is game's table, best first.
func (b *Book) Top(game string) []Entry {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.file.Games[game])
}

// Qualifies reports whether e would enter game's table.
func (b *Book) Qualifies(game string, e Entry) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.rank(game, e) > 0
}

// rank is where e would enter game's table, from 1, or 0 when it would not.
func (b *Book) rank(game string, e Entry) int {
	if e.Score <= 0 {
		return 0
	}
	t := b.file.Games[game]
	i := 0
	for i < len(t) && !e.beats(t[i]) {
		i++
	}
	if i >= Size {
		return 0
	}
	return i + 1
}

// Add enters e, with its name cleaned, in game's table and saves the book.
// It returns e's rank from 1, or 0 when e did not enter. On a save error
// the result stays in memory and the file keeps its previous content.
func (b *Book) Add(game string, e Entry) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	rank := b.rank(game, e)
	if rank == 0 {
		return 0, nil
	}
	e.Name = CleanName(e.Name)
	t := slices.Insert(b.file.Games[game], rank-1, e)
	b.file.Games[game] = t[:min(len(t), Size)]
	if b.path == "" {
		return rank, nil
	}
	return rank, writeJSON(b.path, b.file)
}

// Board is one game's table.
type Board struct {
	book *Book
	game string
}

// Board binds the book to game.
func (b *Book) Board(game string) Board { return Board{book: b, game: game} }

func (b Board) Top() []Entry             { return b.book.Top(b.game) }
func (b Board) Qualifies(e Entry) bool   { return b.book.Qualifies(b.game, e) }
func (b Board) Add(e Entry) (int, error) { return b.book.Add(b.game, e) }
