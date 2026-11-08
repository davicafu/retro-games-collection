package breach

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

func TestRoundMovesTheInjector(t *testing.T) {
	s := sessionOf(func() *Game { return newTestGame(farBrick) })
	start, _ := game(s).Paddle()

	update(t, s, press(input.ActionRight))
	if x, _ := game(s).Paddle(); x != start+PaddleStep {
		t.Fatalf("right: x = %d, want %d", x, start+PaddleStep)
	}
	update(t, s, press(input.ActionLeft), press(input.ActionLeft))
	if x, _ := game(s).Paddle(); x != start-PaddleStep {
		t.Fatalf("left: x = %d, want %d", x, start-PaddleStep)
	}
	update(t, s, press(input.ActionLeft, input.ActionRight))
	if x, _ := game(s).Paddle(); x != start-PaddleStep {
		t.Fatalf("left+right in one frame must cancel, x = %d", x)
	}
}

func TestRoundLaunchKeys(t *testing.T) {
	for _, action := range []input.Action{input.ActionUp, input.ActionSelect} {
		s := sessionOf(func() *Game { return newTestGame(farBrick) })
		update(t, s, press(action))
		if !game(s).Launched() {
			t.Fatalf("%s must launch", action)
		}
	}
}

func TestBreachingShowsTheWinTitleWithScore(t *testing.T) {
	s := sessionOf(func() *Game {
		g := newTestGame(Brick{X: 8, Y: 3, HP: 1})
		place(g, 9, 4, 0, -60)
		return g
	})
	update(t, s, press())

	m, open := s.Overlay()
	if !open || m.Title() != "FIREWALL BREACHED" {
		t.Fatalf("title = %q, want FIREWALL BREACHED", m.Title())
	}
	if content := strings.Join(m.Sections()[0].Lines, "\n"); !strings.Contains(content, "Score: 10") || !strings.Contains(content, "Level: 1/1") {
		t.Fatalf("summary %q missing score or level", content)
	}
}

func TestLosingShowsGameOver(t *testing.T) {
	s := sessionOf(func() *Game {
		g := newTestGame(farBrick)
		g.lives = 1
		place(g, 0, paddleRow, 0, 60)
		return g
	})
	update(t, s, press())

	if m, open := s.Overlay(); !open || m.Title() != "Game Over" {
		t.Fatalf("expected Game Over after losing the last life")
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
	if !strings.Contains(snap, "Score: 0") || !strings.Contains(snap, "▐░░▌") || !strings.Contains(snap, string(glyphVirus)) {
		t.Fatalf("expected HUD, firewall and virus on screen, got:\n%s", snap)
	}
}
