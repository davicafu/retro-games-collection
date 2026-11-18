// Package session runs a real-time game on the runtime loop and owns the flow
// around each playthrough: pause, game over, restart and leaving to the menu.
// Games only provide a Round.
package session

import (
	"context"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/records"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/ui/highscore"
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

// Ranked is optionally implemented by a Round whose result goes on its
// game's high score table; the name is asked by the Session.
type Ranked interface {
	Result() records.Entry
}

// Board is a game's high score table.
type Board = highscore.Board

// Option configures a Session.
type Option func(*Session)

// WithBoard gives the session its game's high score table: a Ranked round
// whose result enters it asks for a name before the end, and the end shows
// the table.
func WithBoard(b Board) Option { return func(s *Session) { s.board = b } }

// TitleHighScore titles the overlay that asks for a name.
const TitleHighScore = highscore.Title

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

	board Board
	// naming asks for the name of a result that entered the table.
	naming *highscore.Prompt
	// rank is the last result's place on the table, from 1, or 0; saveErr
	// is why it could not be saved.
	rank    int
	saveErr error
}

// New starts a session with a first round from newRound; Restart calls it again.
func New(newRound func() Round, opts ...Option) *Session {
	s := &Session{newRound: newRound, round: newRound()}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (s *Session) Init(context.Context) error { return nil }

func (s *Session) Update(ctx context.Context, frame input.FrameInput) error {
	if s.naming != nil {
		s.updateName(frame)
		return nil
	}
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
		s.end()
	}
	return nil
}

// end asks for a name when the round's result enters the table, and
// otherwise opens the end overlay.
func (s *Session) end() {
	s.rank, s.saveErr = 0, nil
	if r, ok := s.round.(Ranked); ok && s.board != nil {
		if e := r.Result(); s.board.Qualifies(e) {
			s.naming = highscore.NewPrompt(e)
			s.openName()
			return
		}
	}
	s.openGameOver()
}

func (s *Session) openName() {
	m := s.naming.Overlay()
	s.overlay = &m
}

// updateName types the name; enter or esc saves the result.
func (s *Session) updateName(frame input.FrameInput) {
	if !s.naming.Update(frame) {
		s.openName()
		return
	}
	e := s.naming.Entry()
	s.naming = nil
	s.rank, s.saveErr = s.board.Add(e)
	s.openGameOver()
}

// table is the end overlay's high score section.
func (s *Session) table() []overlay.Section {
	if _, ok := s.round.(Ranked); !ok || s.board == nil {
		return nil
	}
	return []overlay.Section{overlay.Text(highscore.Lines(s.board.Top(), s.rank, s.saveErr)...)}
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
	m := overlay.NewAction(title, append(s.round.Summary(), s.table()...), []overlay.Option{
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
