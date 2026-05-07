package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidbarreiro/cli-games-collection/internal/shell"
)

// Launcher is the default entry point: runs the interactive shell.
type Launcher func() error

// Runner runs a specific game by name (used by the play command).
type Runner func(game string) error

// RunShell starts the interactive shell (menu → game → menu loop).
func RunShell() error {
	_, err := tea.NewProgram(shell.New()).Run()
	return err
}

// RunGame starts a specific game directly, skipping the menu.
func RunGame(game string) error {
	_, err := tea.NewProgram(shell.NewWithGame(game)).Run()
	return err
}
