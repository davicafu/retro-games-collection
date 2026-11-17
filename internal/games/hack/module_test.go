package hack

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

// typed is a frame as the shell sends text: each rune with its action.
func typed(text string) input.FrameInput {
	f := input.FrameInput{Text: []rune(text)}
	for _, r := range text {
		if a := input.RuneAction(r); a != "" {
			f.Actions = append(f.Actions, a)
		}
	}
	return f
}

func update(t *testing.T, s *session.Session, frames ...input.FrameInput) {
	t.Helper()
	for _, f := range frames {
		if err := s.Update(context.Background(), f); err != nil {
			t.Fatalf("update: %v", err)
		}
	}
}

func sessionOf(g *Game) *session.Session {
	return session.New(func() session.Round { return round{g} })
}

func TestTheTerminalTakesQAndSpaceAsText(t *testing.T) {
	g := newTestGame(Won)
	s := sessionOf(g)
	update(t, s, typed("cat readme.txt q"))
	if s.ShouldQuit() || g.Buffer() != "cat readme.txt q" {
		t.Fatalf("quit %v buffer %q: typed keys must be text", s.ShouldQuit(), g.Buffer())
	}
	update(t, s, press(input.ActionBackspace), press(input.ActionBackspace), press(input.ActionSelect))
	if !strings.Contains(output(g), "Project Lantern") {
		t.Fatalf("enter must run the line:\n%s", g.Text())
	}
	update(t, s, press(input.ActionUp))
	if g.Buffer() != "cat readme.txt" {
		t.Fatalf("up must recall the line, buffer %q", g.Buffer())
	}
}

func TestOnTheMapArrowsMoveAndEnterConnects(t *testing.T) {
	g := newTestGame(Won)
	s := sessionOf(g)
	update(t, s, typed("map"), press(input.ActionSelect), press(input.ActionRight))
	if g.Mode() != ModeMap || g.Selected() != "mail" {
		t.Fatalf("mode %v selected %q", g.Mode(), g.Selected())
	}
	update(t, s, press(input.ActionSelect), press(), press())
	if g.Node() != "mail" {
		t.Fatalf("enter on the map must connect, at %q", g.Node())
	}
}

func TestTheEndTitlesAndSummary(t *testing.T) {
	g := newTestGame(Won)
	g.trace = 36
	s := sessionOf(g)
	update(t, s, typed("connect mail"), press(input.ActionSelect), press(), press())
	update(t, s, typed("connect vault"), press(input.ActionSelect), typed("pixel"), press(input.ActionSelect))
	update(t, s, typed("decrypt core.key.enc ember"), press(input.ActionSelect))
	update(t, s, typed("connect core"), press(input.ActionSelect), press(), press())
	update(t, s, typed("cat lantern.txt"), press(input.ActionSelect))
	m, open := s.Overlay()
	if !open || m.Title() != "ROOT ACCESS" {
		t.Fatalf("title %q, want ROOT ACCESS", m.Title())
	}
	content := strings.Join(m.Sections()[0].Lines, "\n")
	for _, want := range []string{"Trace: 36%", "Time: 00:00"} {
		if !strings.Contains(content, want) {
			t.Fatalf("summary %q missing %q", content, want)
		}
	}

	g = newTestGame(Lost)
	g.trace = 90
	s = sessionOf(g)
	update(t, s, typed("nope"), press(input.ActionSelect), typed("nope"), press(input.ActionSelect), typed("nope"), press(input.ActionSelect))
	if m, open := s.Overlay(); !open || m.Title() != "TRACE COMPLETE" {
		t.Fatalf("a full trace must end with TRACE COMPLETE")
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
	if snap := screen.Snapshot(); !strings.Contains(snap, "ghost@gate:~$") || !strings.Contains(snap, "Trace") {
		t.Fatalf("expected the prompt and the HUD on screen, got:\n%s", snap)
	}
}
