package app

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/records"
	"github.com/davidbarreiro/cli-games-collection/internal/shell"
)

// Launcher is the default entry point: runs the interactive shell.
type Launcher func() error

// Runner runs a specific game by name (used by the play command).
type Runner func(game string) error

// RunShell starts the interactive shell (menu → game → menu loop).
func RunShell() error {
	_, err := tea.NewProgram(shell.New(withRecords())).Run()
	return err
}

// RunGame starts a specific game directly, skipping the menu.
func RunGame(game string) error {
	_, err := tea.NewProgram(shell.NewWithGame(game, withRecords())).Run()
	return err
}

// withRecords keeps high scores and saved runs in the user's config
// directory (records.DefaultDir).
func withRecords() shell.Option {
	dir, err := records.DefaultDir()
	if err != nil {
		dir = ""
	}
	return shell.WithRecords(openRecords(dir))
}

// openRecords opens records.json and saves.json in dir. A file that cannot
// be opened, or an empty dir, is kept in memory instead, so playing never
// depends on the disk.
func openRecords(dir string) (*records.Book, *records.Saves) {
	path := func(name string) string {
		if dir == "" {
			return ""
		}
		return filepath.Join(dir, name)
	}
	book, err := records.Open(path("records.json"))
	if err != nil {
		book, _ = records.Open("")
	}
	saves, err := records.OpenSaves(path("saves.json"))
	if err != nil {
		saves, _ = records.OpenSaves("")
	}
	return book, saves
}
