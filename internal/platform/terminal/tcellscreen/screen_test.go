package tcellscreen

import (
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/gdamore/tcell/v2"
)

func TestMapKeyEvent_MapsArrowAndQuitKeys(t *testing.T) {
	tests := []struct {
		name   string
		evt    *tcell.EventKey
		want   input.Action
		wantOK bool
	}{
		{name: "up", evt: tcell.NewEventKey(tcell.KeyUp, 0, 0), want: input.ActionUp, wantOK: true},
		{name: "down", evt: tcell.NewEventKey(tcell.KeyDown, 0, 0), want: input.ActionDown, wantOK: true},
		{name: "left", evt: tcell.NewEventKey(tcell.KeyLeft, 0, 0), want: input.ActionLeft, wantOK: true},
		{name: "right", evt: tcell.NewEventKey(tcell.KeyRight, 0, 0), want: input.ActionRight, wantOK: true},
		{name: "escape", evt: tcell.NewEventKey(tcell.KeyEscape, 0, 0), want: input.ActionPause, wantOK: true},
		{name: "w", evt: tcell.NewEventKey(tcell.KeyRune, 'w', 0), want: input.ActionUp, wantOK: true},
		{name: "a", evt: tcell.NewEventKey(tcell.KeyRune, 'a', 0), want: input.ActionLeft, wantOK: true},
		{name: "s", evt: tcell.NewEventKey(tcell.KeyRune, 's', 0), want: input.ActionDown, wantOK: true},
		{name: "d", evt: tcell.NewEventKey(tcell.KeyRune, 'd', 0), want: input.ActionRight, wantOK: true},
		{name: "q", evt: tcell.NewEventKey(tcell.KeyRune, 'q', 0), want: input.ActionQuit, wantOK: true},
		{name: "A", evt: tcell.NewEventKey(tcell.KeyRune, 'A', 0), want: input.ActionLeft, wantOK: true},
		{name: "W", evt: tcell.NewEventKey(tcell.KeyRune, 'W', 0), want: input.ActionUp, wantOK: true},
		{name: "S", evt: tcell.NewEventKey(tcell.KeyRune, 'S', 0), want: input.ActionDown, wantOK: true},
		{name: "D", evt: tcell.NewEventKey(tcell.KeyRune, 'D', 0), want: input.ActionRight, wantOK: true},
		{name: "Q", evt: tcell.NewEventKey(tcell.KeyRune, 'Q', 0), want: input.ActionQuit, wantOK: true},
		{name: "unmapped", evt: tcell.NewEventKey(tcell.KeyRune, 'x', 0), wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := mapKeyEvent(tt.evt)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}

			if !tt.wantOK {
				return
			}

			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestNewSimulationScreen_ImplementsRenderContract(t *testing.T) {
	s := newSimulationScreen(10, 4)

	if width, height := s.Size(); width != 10 || height != 4 {
		t.Fatalf("expected size 10x4, got %dx%d", width, height)
	}

	s.SetCell(2, 1, runeCell('X'))

	if got := s.contents[1][2]; got != 'X' {
		t.Fatalf("expected simulation screen cell to be X, got %q", got)
	}

	// Out-of-bounds writes should be ignored to mirror tcell behavior.
	s.SetCell(-1, 1, runeCell('Y'))
	s.SetCell(2, -1, runeCell('Y'))
	s.SetCell(10, 1, runeCell('Y'))
	s.SetCell(2, 4, runeCell('Y'))

	if got := s.contents[1][2]; got != 'X' {
		t.Fatalf("expected in-bounds value to remain X after out-of-bounds writes, got %q", got)
	}

	if err := s.Present(); err != nil {
		t.Fatalf("expected Present to succeed, got error: %v", err)
	}

	s.Clear()
	if got := s.contents[1][2]; got != ' ' {
		t.Fatalf("expected clear to reset cell, got %q", got)
	}
}

func TestSimulationScreen_SnapshotReturnsDeepCopy(t *testing.T) {
	s := NewSimulationScreenForTest(4, 2)
	s.SetCell(1, 1, runeCell('X'))

	snapshot := s.Snapshot()
	snapshot[1][1] = 'Z'

	if got := s.contents[1][1]; got != 'X' {
		t.Fatalf("expected snapshot mutation to not affect screen contents, got %q", got)
	}
}
