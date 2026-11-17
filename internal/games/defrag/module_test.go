package defrag

import (
	"context"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/session"
)

type scriptedProvider struct {
	frames []input.FrameInput
	idx    int
}

func (p *scriptedProvider) Poll() input.FrameInput {
	if p.idx >= len(p.frames) {
		return input.FrameInput{}
	}
	frame := p.frames[p.idx]
	p.idx++
	return frame
}

func press(actions ...input.Action) input.FrameInput {
	return input.FrameInput{Actions: actions}
}

func sessionOf(newGame func() *Game) *session.Session {
	guides := true
	return session.New(func() session.Round { return round{Game: newGame(), guides: &guides} })
}

func update(t *testing.T, s *session.Session, frames ...input.FrameInput) {
	t.Helper()
	for _, f := range frames {
		if err := s.Update(context.Background(), f); err != nil {
			t.Fatalf("update: %v", err)
		}
	}
}

func TestRoundMapsKeysToShiftRotateAndDrops(t *testing.T) {
	s := sessionOf(func() *Game { return newTestGame(Tee, Square) })
	g := s.Round().(round).Game
	update(t, s, press(input.ActionLeft))
	if g.Fragment().Pos.X != 2 {
		t.Fatalf("left: at %v", g.Fragment().Pos)
	}
	g.shift.Stop()
	update(t, s, press(input.ActionRight))
	if g.Fragment().Pos.X != 3 {
		t.Fatalf("right: at %v", g.Fragment().Pos)
	}
	before := g.Fragment().Shape
	update(t, s, press(input.ActionUp))
	if sameCells(before, g.Fragment().Shape) {
		t.Fatalf("up must rotate")
	}
	update(t, s, press(input.ActionDown))
	if g.Score() != PointsSoftDrop {
		t.Fatalf("down must soft drop, score %d", g.Score())
	}
	update(t, s, press(input.ActionSelect))
	if g.Fragment().Kind != Square {
		t.Fatalf("space must hard drop and deal the next fragment")
	}
}

func TestAFullDiskShowsTheEndTitleWithScoreLevelAndSectors(t *testing.T) {
	s := sessionOf(func() *Game {
		g := newTestGame(Square, Square)
		g.score, g.level, g.sectors = 4200, 3, 25
		fill(g, 1, 0)
		return g
	})
	update(t, s, press(input.ActionSelect))
	m, open := s.Overlay()
	if !open || m.Title() != "DISK FULL" {
		t.Fatalf("title = %q, want DISK FULL", m.Title())
	}
	content := strings.Join(m.Sections()[0].Lines, "\n")
	for _, want := range []string{"Level: 3", "Sectors: 25", "Score: "} {
		if !strings.Contains(content, want) {
			t.Fatalf("summary %q missing %q", content, want)
		}
	}
}

func TestModuleRunRendersAndQuits(t *testing.T) {
	manual := clock.NewManual()
	screen := render.NewBufferScreen(80, 24)
	provider := &scriptedProvider{frames: []input.FrameInput{press(), press(input.ActionQuit)}}
	module := NewModule(manual, screen, provider, rand.New(rand.NewSource(1)))

	errCh := make(chan error, 1)
	go func() { errCh <- module.Run(context.Background()) }()
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("module did not quit")
	}
	snap := screen.Snapshot()
	if !strings.Contains(snap, "Score: 0") || !strings.Contains(snap, spriteFragment) || !strings.Contains(snap, "NEXT") {
		t.Fatalf("expected HUD, a fragment and the next panel on screen, got:\n%s", snap)
	}
}

func TestGTogglesTheGuidesForEveryRoundOfTheSession(t *testing.T) {
	rounds := newRounds(rand.New(rand.NewSource(1)))
	first := rounds().(round)
	if !*first.guides {
		t.Fatalf("guides must start on")
	}
	if err := first.Update(context.Background(), press(input.ActionGuides)); err != nil {
		t.Fatal(err)
	}
	if *first.guides {
		t.Fatalf("g must turn the guides off")
	}
	if next := rounds().(round); *next.guides {
		t.Fatalf("a restarted round must keep the guides off")
	}
}
