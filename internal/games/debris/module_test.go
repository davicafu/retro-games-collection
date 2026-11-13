package debris

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/motion"
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
	return session.New(func() session.Round { return round{newGame()} })
}

func update(t *testing.T, s *session.Session, frames ...input.FrameInput) {
	t.Helper()
	for _, f := range frames {
		if err := s.Update(context.Background(), f); err != nil {
			t.Fatalf("update: %v", err)
		}
	}
}

func TestRoundMapsKeysToTurnThrustAndFire(t *testing.T) {
	s := sessionOf(func() *Game { return newTestGame(far(Large)) })
	g := s.Round().(round).Game
	update(t, s, press(input.ActionRight))
	if g.Drone().Heading != 1 {
		t.Fatalf("right: heading %d, want 1", g.Drone().Heading)
	}
	update(t, s, press(input.ActionLeft))
	if g.Drone().Heading != 0 {
		t.Fatalf("left: heading %d, want 0", g.Drone().Heading)
	}
	update(t, s, press(input.ActionRight, input.ActionUp, input.ActionSelect))
	if g.Drone().Heading != 1 || g.Drone().Vel == (motion.Vec{}) || len(g.Shots()) != 1 {
		t.Fatalf("one frame with a turn, thrust and fire must do all three: %+v, %d shots", g.Drone(), len(g.Shots()))
	}
}

func TestBreachedHullShowsTheEndTitleWithScoreAndWave(t *testing.T) {
	centre := motion.Vec{X: testCols * CellW / 2, Y: testRows * CellH / 2}
	s := sessionOf(func() *Game {
		g := newTestGame(Piece{Size: Large, Body: motion.Body{Pos: centre}})
		g.lives, g.score, g.wave = 1, 270, 2
		return g
	})
	update(t, s, press())
	m, open := s.Overlay()
	if !open || m.Title() != "HULL BREACHED" {
		t.Fatalf("title = %q, want HULL BREACHED", m.Title())
	}
	if content := strings.Join(m.Sections()[0].Lines, "\n"); !strings.Contains(content, "Score: 270") || !strings.Contains(content, "Wave: 2") {
		t.Fatalf("summary %q missing score or wave", content)
	}
}

func TestModuleRunRendersAndQuits(t *testing.T) {
	manual := clock.NewManual()
	screen := render.NewBufferScreen(80, 24)
	provider := &scriptedProvider{frames: []input.FrameInput{press(), press(input.ActionQuit)}}
	module := NewModule(manual, screen, provider, zeroRand{})

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
	if !strings.Contains(snap, "Wave 1") || !strings.Contains(snap, "▲") || !strings.Contains(snap, "█") {
		t.Fatalf("expected HUD, drone and debris on screen, got:\n%s", snap)
	}
}
