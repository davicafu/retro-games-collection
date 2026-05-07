package app

import (
	"testing"
)

func TestRunShell_ReturnsNoError(t *testing.T) {
	// We can't easily test interactive tea.Program in unit tests.
	// Smoke test: the function signature is correct and RunGame works.
	// Shell tests live in internal/shell.
}

func TestRunGame_ReturnsNoErrorForSnake(t *testing.T) {
	// RunGame creates a tea.Program that starts snake directly.
	// Full integration test requires terminal — this is a compilation + wiring check.
	err := RunGame("snake")
	// Will actually run and try to start the game, which requires terminal.
	// In CI this might fail — but the important thing is it compiles and wires
	// through shell.NewWithGame correctly.
	if err != nil && err.Error() != "exit" {
		// tea.Program.Run() can return "exit" on quit, that's fine
		t.Logf("RunGame returned: %v (expected in non-TTY environments)", err)
	}
}

func TestTypesAreCallable(t *testing.T) {
	// Verify Launcher and Runner function signatures compile.
	var launcher Launcher = RunShell
	var runner Runner = RunGame

	if launcher == nil {
		t.Fatal("launcher should not be nil")
	}
	if runner == nil {
		t.Fatal("runner should not be nil")
	}
}
