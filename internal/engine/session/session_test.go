package session

import (
	"context"
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"
)

// fakeRound records the frames it receives and ends when told to.
type fakeRound struct {
	id      int
	updates int
	frames  []input.FrameInput
	endOn   int // update count that ends the round; 0 = never
}

func (r *fakeRound) Update(_ context.Context, frame input.FrameInput) error {
	r.updates++
	r.frames = append(r.frames, frame)
	return nil
}

func (r *fakeRound) Render(_ context.Context, screen render.Screen) error {
	render.WriteText(screen, 0, 0, "ROUND")
	return nil
}

func (r *fakeRound) Over() bool { return r.endOn > 0 && r.updates >= r.endOn }

func (r *fakeRound) Summary() []overlay.Section {
	return []overlay.Section{overlay.Text("Score: 7")}
}

// newTestSession returns a session whose rounds end after endOn updates,
// plus a pointer to the number of rounds created so far.
func newTestSession(endOn int) (*Session, *int) {
	created := 0
	s := New(func() Round {
		created++
		return &fakeRound{id: created, endOn: endOn}
	})
	return s, &created
}

func press(actions ...input.Action) input.FrameInput {
	return input.FrameInput{Actions: actions}
}

func update(t *testing.T, s *Session, frames ...input.FrameInput) {
	t.Helper()
	for _, f := range frames {
		if err := s.Update(context.Background(), f); err != nil {
			t.Fatalf("update: %v", err)
		}
	}
}

func round(s *Session) *fakeRound { return s.Round().(*fakeRound) }

func optionIDs(m overlay.Model) []string {
	ids := make([]string, 0, len(m.Options()))
	for _, o := range m.Options() {
		ids = append(ids, o.ID)
	}
	return ids
}

func TestPlayingForwardsFramesToRound(t *testing.T) {
	s, _ := newTestSession(0)

	update(t, s, press(input.ActionUp), press())

	if got := round(s).updates; got != 2 {
		t.Fatalf("round updates = %d, want 2", got)
	}
	if _, open := s.Overlay(); open || s.ShouldQuit() {
		t.Fatalf("plain play must not open overlays or quit")
	}
}

func TestQuitWhilePlayingStopsWithoutUpdatingRound(t *testing.T) {
	s, _ := newTestSession(0)

	update(t, s, press(input.ActionUp, input.ActionQuit))

	if !s.ShouldQuit() {
		t.Fatalf("q must quit")
	}
	if round(s).updates != 0 {
		t.Fatalf("quit tick must not update the round")
	}
}

func TestRoundOverOpensGameOverWithSummaryAndNoContinue(t *testing.T) {
	s, _ := newTestSession(1)

	update(t, s, press())

	m, open := s.Overlay()
	if !open || m.Title() != "Game Over" {
		t.Fatalf("expected Game Over overlay")
	}
	if got := strings.Join(m.Sections()[0].Lines, "|"); got != "Score: 7" {
		t.Fatalf("summary = %q, want round summary", got)
	}
	if got := optionIDs(m); strings.Join(got, ",") != OptionRestart+","+OptionMenu {
		t.Fatalf("options = %v, want [restart menu]", got)
	}
	if s.ShouldQuit() {
		t.Fatalf("must not quit while Game Over is open")
	}
}

func TestOverlayBlocksRoundUpdatesAndKeepsOpenOnEmptyFrames(t *testing.T) {
	s, _ := newTestSession(1)
	update(t, s, press())

	update(t, s, press(), press())

	if round(s).updates != 1 {
		t.Fatalf("round updated while overlay open: %d", round(s).updates)
	}
	if _, open := s.Overlay(); !open {
		t.Fatalf("overlay must stay open on empty frames")
	}
}

func TestOverlayNavigation(t *testing.T) {
	s, _ := newTestSession(1)
	update(t, s, press())

	update(t, s, press(input.ActionDown))
	if m, _ := s.Overlay(); m.SelectedIndex() != 1 {
		t.Fatalf("down: selected = %d, want 1", m.SelectedIndex())
	}
	update(t, s, press(input.ActionUp))
	if m, _ := s.Overlay(); m.SelectedIndex() != 0 {
		t.Fatalf("up: selected = %d, want 0", m.SelectedIndex())
	}
}

func TestGameOverRestartStartsFreshRound(t *testing.T) {
	for _, confirm := range []input.Action{input.ActionSelect, input.ActionQuit} {
		s, created := newTestSession(1)
		update(t, s, press(), press(confirm))

		if *created != 2 || round(s).id != 2 {
			t.Fatalf("%s: expected a fresh round, created=%d", confirm, *created)
		}
		if _, open := s.Overlay(); open || s.ShouldQuit() {
			t.Fatalf("%s: restart must close the overlay and keep playing", confirm)
		}
	}
}

func TestGameOverMenuQuits(t *testing.T) {
	for _, confirm := range []input.Action{input.ActionSelect, input.ActionQuit} {
		s, _ := newTestSession(1)
		update(t, s, press(), press(input.ActionDown), press(confirm))

		if !s.ShouldQuit() {
			t.Fatalf("%s on Menu must end the session", confirm)
		}
	}
}

func TestPauseOpensPausedOverlayWithContinueFirst(t *testing.T) {
	s, _ := newTestSession(0)

	update(t, s, press(input.ActionPause))

	m, open := s.Overlay()
	if !open || m.Title() != "PAUSED" {
		t.Fatalf("expected PAUSED overlay")
	}
	if got := optionIDs(m); strings.Join(got, ",") != OptionContinue+","+OptionRestart+","+OptionMenu {
		t.Fatalf("options = %v, want [continue restart menu]", got)
	}
	if round(s).updates != 0 {
		t.Fatalf("pause tick must not update the round")
	}
}

func TestPauseResumes(t *testing.T) {
	cases := map[string][]input.FrameInput{
		"pause again":     {press(input.ActionPause)},
		"quit":            {press(input.ActionQuit)},
		"select continue": {press(input.ActionSelect)},
	}
	for name, frames := range cases {
		s, _ := newTestSession(0)
		update(t, s, press(input.ActionPause))
		update(t, s, frames...)

		if _, open := s.Overlay(); open {
			t.Fatalf("%s: expected overlay closed", name)
		}
		if s.ShouldQuit() {
			t.Fatalf("%s: resume must not quit", name)
		}
		update(t, s, press())
		if round(s).updates != 1 {
			t.Fatalf("%s: round must update again after resume", name)
		}
	}
}

func TestPauseRestartAndMenu(t *testing.T) {
	s, created := newTestSession(0)
	update(t, s, press(), press(input.ActionPause), press(input.ActionDown), press(input.ActionSelect))
	if *created != 2 {
		t.Fatalf("pause Restart must start a fresh round, created=%d", *created)
	}

	s, _ = newTestSession(0)
	update(t, s, press(input.ActionPause), press(input.ActionDown), press(input.ActionDown), press(input.ActionSelect))
	if !s.ShouldQuit() {
		t.Fatalf("pause Menu must end the session")
	}
}

func TestRenderDrawsRoundThenOverlayOnTop(t *testing.T) {
	s, _ := newTestSession(1)
	screen := render.NewBufferScreen(40, 12)

	if err := s.Render(context.Background(), screen); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.HasPrefix(screen.Snapshot(), "ROUND") {
		t.Fatalf("round not rendered:\n%s", screen.Snapshot())
	}

	update(t, s, press())
	screen.Clear()
	if err := s.Render(context.Background(), screen); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(screen.Snapshot(), "Game Over") {
		t.Fatalf("overlay not rendered:\n%s", screen.Snapshot())
	}
}

// titledRound ends at once and names its own end-of-round title.
type titledRound struct{ fakeRound }

func (r *titledRound) OverTitle() string { return "FIREWALL BREACHED" }

func TestRoundCanNameTheGameOverTitle(t *testing.T) {
	s := New(func() Round { return &titledRound{fakeRound{endOn: 1}} })
	update(t, s, press())

	m, open := s.Overlay()
	if !open || m.Title() != "FIREWALL BREACHED" {
		t.Fatalf("title = %q, want the round's title", m.Title())
	}
}

// typingRound is a fakeRound that takes text while typing is set.
type typingRound struct {
	fakeRound
	typing bool
}

func (r *typingRound) TakingText() bool { return r.typing }

func TestATypingRoundGetsLettersAsTextNotQuitOrPause(t *testing.T) {
	r := &typingRound{typing: true}
	s := New(func() Round { return r })
	update(t, s, input.FrameInput{Actions: []input.Action{input.ActionQuit, input.ActionPause}, Text: []rune{'q', 'p'}})
	if s.ShouldQuit() {
		t.Fatalf("typing q must not quit")
	}
	if _, open := s.Overlay(); open {
		t.Fatalf("typing p must not pause")
	}
	got := r.frames[0]
	if len(got.Actions) != 0 || string(got.Text) != "qp" {
		t.Fatalf("round got %+v, want the text without its actions", got)
	}
	update(t, s, press(input.ActionPause)) // Esc
	if _, open := s.Overlay(); !open {
		t.Fatalf("Esc must still pause a typing round")
	}
}

func TestARoundNotTypingKeepsLettersAsActions(t *testing.T) {
	r := &typingRound{typing: false}
	s := New(func() Round { return r })
	update(t, s, input.FrameInput{Actions: []input.Action{input.ActionQuit}, Text: []rune{'q'}})
	if !s.ShouldQuit() {
		t.Fatalf("q must quit a round that is not typing")
	}
}
