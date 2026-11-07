package runner

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

// newCollisionSession returns a session whose first round hits a standing
// player on tick 3; restarts produce an empty world.
func newCollisionSession() *session.Session {
	first := true
	return session.New(func() session.Round {
		if first {
			first = false
			return round{newTestGame(Obstacle{X: testPlayerX + 1, Width: 1, Tall: 1, Kind: KindGround})}
		}
		return round{newTestGame()}
	})
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

func TestRoundJumpKeys(t *testing.T) {
	for _, action := range []input.Action{input.ActionUp, input.ActionSelect} {
		s := newCollisionSession()
		update(t, s, press(action))
		if h := game(s).PlayerHeight(); h != 1 {
			t.Fatalf("%s: height = %d, want 1", action, h)
		}
	}
}

func TestRoundDownDucks(t *testing.T) {
	s := newCollisionSession()
	update(t, s, press(input.ActionDown))
	if !game(s).Ducking() {
		t.Fatalf("down must duck")
	}
}

func TestRoundJumpWinsOverDuckInTheSameFrame(t *testing.T) {
	s := newCollisionSession()
	update(t, s, press(input.ActionDown, input.ActionUp))
	if game(s).Ducking() || game(s).PlayerHeight() != 1 {
		t.Fatalf("jump must win over duck in one frame")
	}
}

func TestCollisionShowsGameOverWithScore(t *testing.T) {
	s := newCollisionSession()
	update(t, s, press(), press(), press())

	m, open := s.Overlay()
	if !open || m.Title() != "Game Over" {
		t.Fatalf("expected Game Over overlay after collision")
	}
	if content := strings.Join(m.Sections()[0].Lines, "\n"); !strings.Contains(content, "Score: 1") {
		t.Fatalf("overlay content %q missing score", content)
	}
}

func TestGameOverRestartBuildsFreshGame(t *testing.T) {
	s := newCollisionSession()
	update(t, s, press(), press(), press(), press(input.ActionSelect))

	g := game(s)
	if g.Dead() || g.Score() != 0 || len(g.Obstacles()) != 0 {
		t.Fatalf("restart must start a fresh game: dead=%v score=%d", g.Dead(), g.Score())
	}
}

func TestModuleRunRendersAndQuits(t *testing.T) {
	manual := clock.NewManual()
	screen := render.NewBufferScreen(80, 24)
	provider := &scriptedProvider{frames: []input.FrameInput{press(), press(input.ActionQuit)}}
	module := NewModule(manual, screen, provider, &scriptedRand{values: []int{0}})

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
	if !strings.Contains(screen.Snapshot(), "Score: 0") {
		t.Fatalf("expected HUD on screen, got:\n%s", screen.Snapshot())
	}
}
