package rack

import (
	"context"
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

func TestRoundMapsArrowsToMovesAndSpaceToJump(t *testing.T) {
	s := sessionOf(func() *Game { return newTestGame(room...) })
	g := s.Round().(round).Game
	update(t, s, press(input.ActionRight))
	if g.Player().Pos != pt(2, 4) {
		t.Fatalf("right: at %v, want (2,4)", g.Player().Pos)
	}
	update(t, s, press(input.ActionLeft))
	if g.Player().Pos != pt(1, 4) {
		t.Fatalf("left: at %v, want (1,4)", g.Player().Pos)
	}
	update(t, s, press(input.ActionSelect))
	if !g.Player().Jumping() {
		t.Fatalf("space must jump")
	}
	g.wrenches = 1
	update(t, s, press(input.ActionFire))
	if g.WrenchesLeft() != 0 {
		t.Fatalf("f must throw a wrench")
	}
	g.player = Player{Pos: pt(5, 4)}
	update(t, s, press(input.ActionUp))
	if g.Player().Pos != pt(5, 3) {
		t.Fatalf("up: at %v, want (5,3)", g.Player().Pos)
	}
	g.walk.Stop()
	update(t, s, press(input.ActionDown))
	if g.Player().Pos != pt(5, 4) {
		t.Fatalf("down: at %v, want (5,4)", g.Player().Pos)
	}
}

func TestDeniedAccessShowsTheEndTitleWithScoreAndLevel(t *testing.T) {
	s := sessionOf(func() *Game {
		g := newTestGame(room...)
		g.lives, g.score, g.level = 1, 1200, 2
		g.bots = []Bot{{Pos: pt(2, 4), Dir: -1}}
		return g
	})
	update(t, s, press(input.ActionRight))
	m, open := s.Overlay()
	if !open || m.Title() != "ACCESS DENIED" {
		t.Fatalf("title = %q, want ACCESS DENIED", m.Title())
	}
	if content := strings.Join(m.Sections()[0].Lines, "\n"); !strings.Contains(content, "Score: 1200") || !strings.Contains(content, "Level: 2") {
		t.Fatalf("summary %q missing score or level", content)
	}
}

func TestModuleRunRendersAndQuits(t *testing.T) {
	manual := clock.NewManual()
	screen := render.NewBufferScreen(80, 24)
	provider := &scriptedProvider{frames: []input.FrameInput{press(), press(input.ActionQuit)}}
	module := NewModule(manual, screen, provider, &scriptedRand{})

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
	if !strings.Contains(snap, "Level 1") || !strings.Contains(snap, spriteTechnician) || !strings.Contains(snap, spriteLadder) {
		t.Fatalf("expected HUD, technician and ladders on screen, got:\n%s", snap)
	}
}
