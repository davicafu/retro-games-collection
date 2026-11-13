package leak

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
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

func game(s *session.Session) *Game { return s.Round().(round).Game }

func update(t *testing.T, s *session.Session, frames ...input.FrameInput) {
	t.Helper()
	for _, f := range frames {
		if err := s.Update(context.Background(), f); err != nil {
			t.Fatalf("update: %v", err)
		}
	}
}

func TestRoundMapsArrowsToDirections(t *testing.T) {
	cases := map[input.Action]grid.Point{
		input.ActionUp:    Up,
		input.ActionDown:  Down,
		input.ActionLeft:  Left,
		input.ActionRight: Right,
	}
	for action, want := range cases {
		s := sessionOf(func() *Game { return newTestGame(RAM...) })
		update(t, s, press(action))
		if got := game(s).wanted; got != want {
			t.Fatalf("%s: wanted %v, want %v", action, got, want)
		}
	}
}

func TestRoundKeepsTheLastDirectionOfAFrame(t *testing.T) {
	s := sessionOf(func() *Game { return newTestGame(RAM...) })
	update(t, s, press(input.ActionUp, input.ActionLeft))
	if got := game(s).wanted; got != Left {
		t.Fatalf("wanted %v, want the last direction pressed (left)", got)
	}
}

func TestBeingCollectedShowsTheEndTitleWithScoreAndLevel(t *testing.T) {
	s := sessionOf(func() *Game {
		g := newTestGame(corridor...)
		g.lives, g.score, g.level = 1, 340, 2
		addCollector(g, KindTracer, pt(2, 1), Left)
		return g
	})
	for i := 0; i < 10; i++ {
		update(t, s, press(input.ActionRight))
	}
	m, open := s.Overlay()
	if !open || m.Title() != "COLLECTED" {
		t.Fatalf("title = %q, want COLLECTED", m.Title())
	}
	// The leak eats the byte at (2,1) as it runs into the collector.
	if content := strings.Join(m.Sections()[0].Lines, "\n"); !strings.Contains(content, "Score: 350") || !strings.Contains(content, "Level: 2") {
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
	if !strings.Contains(snap, "Level 1") || !strings.Contains(snap, spriteLeak) || !strings.Contains(snap, spriteCollector[KindTracer]) {
		t.Fatalf("expected HUD, leak and collectors on screen, got:\n%s", snap)
	}
}
