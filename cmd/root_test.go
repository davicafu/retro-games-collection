package cmd

import (
	"errors"
	"os"
	"testing"
)

func TestNewRootCommand_UsesCollectionName(t *testing.T) {
	cmd := NewRootCommand(func() error { return nil }, nil)

	if cmd.Use != "retro-games" {
		t.Fatalf("expected root command use to be retro-games, got %q", cmd.Use)
	}

	if !cmd.SilenceErrors {
		t.Fatal("expected root command to silence cobra errors for single-source error output")
	}
}

func TestRootCommand_InvokesLauncherCallback(t *testing.T) {
	called := false

	cmd := NewRootCommand(func() error {
		called = true
		return nil
	}, func(string) error { return nil })

	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !called {
		t.Fatal("expected launcher callback to be called")
	}
}

func TestPlayCommand_ForwardsSelectedGame(t *testing.T) {
	var selected string

	cmd := NewRootCommand(func() error { return nil }, func(game string) error {
		selected = game
		return nil
	})

	cmd.SetArgs([]string{"play", "snake"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if selected != "snake" {
		t.Fatalf("expected selected game to be snake, got %q", selected)
	}
}

func TestPlayCommand_EnforcesArgValidation(t *testing.T) {
	cmd := NewRootCommand(func() error { return nil }, func(string) error { return nil })

	cmd.SetArgs([]string{"play"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error for missing game argument")
	}
}

func TestCallbacks_ErrorPropagation(t *testing.T) {
	t.Run("launcher error", func(t *testing.T) {
		launcherErr := errors.New("launcher failed")
		cmd := NewRootCommand(func() error { return launcherErr }, func(string) error { return nil })

		cmd.SetArgs(nil)
		err := cmd.Execute()
		if !errors.Is(err, launcherErr) {
			t.Fatalf("expected launcher error to propagate, got %v", err)
		}
	})

	t.Run("play callback error", func(t *testing.T) {
		runGameErr := errors.New("run game failed")
		cmd := NewRootCommand(func() error { return nil }, func(string) error { return runGameErr })

		cmd.SetArgs([]string{"play", "snake"})
		err := cmd.Execute()
		if !errors.Is(err, runGameErr) {
			t.Fatalf("expected run game error to propagate, got %v", err)
		}
	})
}

func TestExecute_InvokesLauncherCallback(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"retro-games"}
	t.Cleanup(func() { os.Args = originalArgs })

	called := false
	err := Execute(func() error {
		called = true
		return nil
	}, func(string) error { return nil })

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !called {
		t.Fatal("expected launcher callback to be called")
	}
}

func TestExecute_ForwardsPlaySelectedGame(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"retro-games", "play", "snake"}
	t.Cleanup(func() { os.Args = originalArgs })

	var selected string
	err := Execute(func() error { return nil }, func(game string) error {
		selected = game
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if selected != "snake" {
		t.Fatalf("expected selected game to be snake, got %q", selected)
	}
}

func TestExecute_PropagatesLauncherError(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"retro-games"}
	t.Cleanup(func() { os.Args = originalArgs })

	launcherErr := errors.New("launcher failed")
	err := Execute(func() error { return launcherErr }, func(string) error { return nil })

	if !errors.Is(err, launcherErr) {
		t.Fatalf("expected launcher error to propagate, got %v", err)
	}
}
