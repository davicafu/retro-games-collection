package drone

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

func newTestSession() *session.Session {
	return session.New(func() session.Round { return round{newTestGame()} })
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

func TestRoundThrustKeys(t *testing.T) {
	for _, action := range []input.Action{input.ActionUp, input.ActionSelect} {
		s := newTestSession()
		update(t, s, press(action))
		if g := game(s); !g.Started() || g.Velocity() != -Thrust {
			t.Fatalf("%s must thrust: started=%v velocity=%d", action, g.Started(), g.Velocity())
		}
	}
}

func TestRoundOtherKeysDoNotThrust(t *testing.T) {
	s := newTestSession()
	update(t, s, press(input.ActionDown), press(input.ActionLeft), press(input.ActionRight))
	if game(s).Started() {
		t.Fatalf("only thrust keys may take off")
	}
}

func TestCrashShowsGameOverWithScoreAndRestartIsFresh(t *testing.T) {
	s := newTestSession()
	update(t, s, press(input.ActionUp))
	for i := 0; i < 200; i++ {
		if _, open := s.Overlay(); open {
			break
		}
		update(t, s, press())
	}

	m, open := s.Overlay()
	if !open || m.Title() != "Game Over" {
		t.Fatalf("expected Game Over after falling to the ground")
	}
	if content := strings.Join(m.Sections()[0].Lines, "\n"); !strings.Contains(content, "Score: 0") {
		t.Fatalf("overlay content %q missing score", content)
	}

	update(t, s, press(input.ActionSelect))
	if g := game(s); g.Dead() || g.Started() {
		t.Fatalf("restart must start a fresh round")
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
	if snap := screen.Snapshot(); !strings.Contains(snap, "Score: 0") || !strings.Contains(snap, string(glyphDrone)) {
		t.Fatalf("expected HUD and drone on screen, got:\n%s", snap)
	}
}
