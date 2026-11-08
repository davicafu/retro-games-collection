package botnet

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

func game(s *session.Session) *Game { return s.Round().(round).Game }

func update(t *testing.T, s *session.Session, frames ...input.FrameInput) {
	t.Helper()
	for _, f := range frames {
		if err := s.Update(context.Background(), f); err != nil {
			t.Fatalf("update: %v", err)
		}
	}
}

func TestRoundMovesTheServer(t *testing.T) {
	s := sessionOf(func() *Game { return newTestGame([]Scanner{lone}, nil) })
	start, _ := game(s).Server()
	update(t, s, press(input.ActionRight))
	if x, _ := game(s).Server(); x != start+1 {
		t.Fatalf("right: x = %d, want %d", x, start+1)
	}
	update(t, s, press(input.ActionLeft))
	if x, _ := game(s).Server(); x != start {
		t.Fatalf("left: x = %d, want back at %d", x, start)
	}
	for i := 0; i < ServerGlide; i++ {
		update(t, s, press())
	}
	rest, _ := game(s).Server()
	update(t, s, press(input.ActionLeft, input.ActionRight))
	if x, _ := game(s).Server(); x != rest {
		t.Fatalf("left+right in one frame must cancel, x = %d, want %d", x, rest)
	}
}

func TestRoundFireKeys(t *testing.T) {
	for _, action := range []input.Action{input.ActionUp, input.ActionSelect} {
		s := sessionOf(func() *Game { return newTestGame([]Scanner{lone}, nil) })
		update(t, s, press(action))
		if _, _, ok := game(s).Packet(); !ok {
			t.Fatalf("%s must fire", action)
		}
	}
}

func TestQuarantineShowsTheEndTitleWithScoreAndWave(t *testing.T) {
	x, _ := newTestGame(nil, nil).Server()
	s := sessionOf(func() *Game {
		g := newTestGame([]Scanner{lone}, nil)
		g.lives, g.score, g.wave = 1, 70, 2
		g.shots = []Shot{{Col: x, Y: (serverRow-1)*SubCells + SubCells - 1}}
		return g
	})
	update(t, s, press())

	m, open := s.Overlay()
	if !open || m.Title() != "QUARANTINED" {
		t.Fatalf("title = %q, want QUARANTINED", m.Title())
	}
	if content := strings.Join(m.Sections()[0].Lines, "\n"); !strings.Contains(content, "Score: 70") || !strings.Contains(content, "Wave: 2") {
		t.Fatalf("summary %q missing score or wave", content)
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
	if !strings.Contains(snap, "Wave 1") || !strings.Contains(snap, scannerSprites[KindHigh][0]) || !strings.Contains(snap, spriteServer) {
		t.Fatalf("expected HUD, formation and server on screen, got:\n%s", snap)
	}
}
