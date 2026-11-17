// Package session runs a real-time game on the runtime loop and owns the flow
// around each playthrough: pause, game over, restart and leaving to the menu.
// Games only provide a Round.
package session

import (
	"context"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"
)

// Round is one playthrough of a game. Update only receives frames while the
// round is being played: quit and pause are handled by the Session.
type Round interface {
	Update(ctx context.Context, frame input.FrameInput) error
	Render(ctx context.Context, screen render.Screen) error
	// Over reports that the playthrough ended, e.g. the player died.
	Over() bool
	// Summary is shown in the Game Over overlay, e.g. the final score.
	Summary() []overlay.Section
}

// TextTaker is optionally implemented by a Round that takes typed text,
// e.g. at a prompt. While TakingText is true the round gets typed
// characters as text only: Q and P do not quit or pause, while Esc still
// pauses.
type TextTaker interface {
	TakingText() bool
}

// OverTitler is optionally implemented by a Round to title the end-of-round
// overlay, e.g. for a win. Rounds without it get "Game Over".
type OverTitler interface {
	OverTitle() string
}

// Overlay option IDs.
const (
	OptionContinue = "continue"
	OptionRestart  = "restart"
	OptionMenu     = "menu"
)

// Session implements runtime.Game. Ending it (Q while playing, or Menu in an
// overlay) makes ShouldQuit true; the launcher then returns to its menu.
type Session struct {
	newRound func() Round
	round    Round
	overlay  *overlay.Model
	paused   bool
	quit     bool
}

// New starts a session with a first round from newRound; Restart calls it again.
func New(newRound func() Round) *Session {
	return &Session{newRound: newRound, round: newRound()}
}

func (s *Session) Init(context.Context) error { return nil }

func (s *Session) Update(ctx context.Context, frame input.FrameInput) error {
	if s.overlay != nil {
		s.updateOverlay(frame)
		return nil
	}

	if t, ok := s.round.(TextTaker); ok && t.TakingText() {
		frame = input.WithoutTyping(frame)
	}
	for _, a := range frame.Actions {
		switch a {
		case input.ActionQuit:
			s.quit = true
			return nil
		case input.ActionPause:
			s.openPause()
			return nil
		}
	}

	if err := s.round.Update(ctx, frame); err != nil {
		return err
	}
	if s.round.Over() {
		s.openGameOver()
	}
	return nil
}

func (s *Session) updateOverlay(frame input.FrameInput) {
	for _, a := range frame.Actions {
		switch a {
		case input.ActionUp:
			s.overlay.Apply(overlay.IntentMoveUp)
		case input.ActionDown:
			s.overlay.Apply(overlay.IntentMoveDown)
		case input.ActionSelect:
			s.resolve(s.overlay.Apply(overlay.IntentConfirm))
			return
		case input.ActionPause, input.ActionQuit:
			// While paused both keys resume. On Game Over, Q confirms the
			// selected option (pinned by the original Snake behavior).
			if s.paused {
				s.closeOverlay()
				return
			}
			if a == input.ActionQuit {
				s.resolve(s.overlay.Apply(overlay.IntentConfirm))
				return
			}
		}
	}
}

func (s *Session) resolve(result overlay.Result) {
	if result.Kind != overlay.ResultKindConfirmed {
		return
	}
	switch result.OptionID {
	case OptionContinue:
		s.closeOverlay()
	case OptionRestart:
		s.round = s.newRound()
		s.closeOverlay()
	case OptionMenu:
		s.quit = true
		s.closeOverlay()
	}
}

func (s *Session) openPause() {
	m := overlay.NewAction("PAUSED", nil, []overlay.Option{
		{ID: OptionContinue, Label: "Continue", Enabled: true},
		{ID: OptionRestart, Label: "Restart", Enabled: true},
		{ID: OptionMenu, Label: "Menu", Enabled: true},
	})
	s.overlay = &m
	s.paused = true
}

func (s *Session) openGameOver() {
	title := "Game Over"
	if t, ok := s.round.(OverTitler); ok {
		title = t.OverTitle()
	}
	m := overlay.NewAction(title, s.round.Summary(), []overlay.Option{
		{ID: OptionRestart, Label: "Restart", Enabled: true},
		{ID: OptionMenu, Label: "Menu", Enabled: true},
	})
	s.overlay = &m
}

func (s *Session) closeOverlay() {
	s.overlay = nil
	s.paused = false
}

// Render draws the round, then any open overlay on top of it.
func (s *Session) Render(ctx context.Context, screen render.Screen) error {
	if err := s.round.Render(ctx, screen); err != nil {
		return err
	}
	if s.overlay != nil {
		overlay.Draw(screen, *s.overlay)
	}
	return nil
}

func (s *Session) ShouldQuit() bool { return s.quit }

// Round returns the round being played.
func (s *Session) Round() Round { return s.round }

// Overlay returns the open overlay, if any.
func (s *Session) Overlay() (overlay.Model, bool) {
	if s.overlay == nil {
		return overlay.Model{}, false
	}
	return *s.overlay, true
}
