package packet

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

// fixedRand always returns 0.
type fixedRand struct{}

func (fixedRand) Intn(int) int { return 0 }

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

func TestRoundSwitchesLinksWithUpAndDownAndTheLastPressWins(t *testing.T) {
	s := sessionOf(func() *Game { return newTestGame(1, nil, nil) })
	g := s.Round().(round).Game
	update(t, s, press(input.ActionDown))
	if g.Link() != 2 {
		t.Fatalf("down: link %d, want 2", g.Link())
	}
	update(t, s, press(input.ActionUp))
	if g.Link() != 1 {
		t.Fatalf("up: link %d, want 1", g.Link())
	}
	update(t, s, press(input.ActionDown, input.ActionUp))
	if g.Link() != 0 {
		t.Fatalf("down then up in one frame: link %d, want 0 (the last press)", g.Link())
	}
}

func TestBeingDroppedShowsTheEndTitleWithScoreAndDistance(t *testing.T) {
	s := sessionOf(func() *Game {
		return newTestGame(1, []Firewall{{Link: 1, X: 10, Length: 1}}, []Token{{Link: 1, X: 7}})
	})
	for i := 0; i < 100; i++ {
		update(t, s, press())
	}
	m, open := s.Overlay()
	if !open || m.Title() != "PACKET DROPPED" {
		t.Fatalf("title = %q, want PACKET DROPPED", m.Title())
	}
	if content := strings.Join(m.Sections()[0].Lines, "\n"); !strings.Contains(content, "Score: 55") || !strings.Contains(content, "Distance: 5") {
		t.Fatalf("summary %q missing score or distance", content)
	}
}

func TestModuleRunRendersAndQuits(t *testing.T) {
	manual := clock.NewManual()
	screen := render.NewBufferScreen(80, 24)
	provider := &scriptedProvider{frames: []input.FrameInput{press(), press(input.ActionQuit)}}
	module := NewModule(manual, screen, provider, fixedRand{})

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
	if !strings.Contains(snap, "Score: 0") || !strings.Contains(snap, string(glyphPacket)) || !strings.Contains(snap, string(glyphSeparator)) {
		t.Fatalf("expected HUD, links and packet on screen, got:\n%s", snap)
	}
}
