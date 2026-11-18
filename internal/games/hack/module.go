package hack

import (
	"context"
	"errors"
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/records"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/runtime"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/session"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"
)

// round adapts Game to session.Round. At the terminal typed keys are text
// (the session strips their actions), enter runs the line, backspace
// deletes and up/down walk the history. On the map and in a breach the
// arrows (and WASD) move, and enter or space act.
type round struct{ *Game }

func (r round) Update(_ context.Context, frame input.FrameInput) error {
	in := Input{Text: frame.Text}
	for _, a := range frame.Actions {
		switch a {
		case input.ActionUp:
			in.Move, in.History = grid.Point{Y: -1}, -1
		case input.ActionDown:
			in.Move, in.History = grid.Point{Y: 1}, 1
		case input.ActionLeft:
			in.Move = grid.Point{X: -1}
		case input.ActionRight:
			in.Move = grid.Point{X: 1}
		case input.ActionSelect:
			in.Enter, in.Act = true, true
		case input.ActionBackspace:
			in.Backspace = true
		}
	}
	r.Step(in)
	return nil
}

func (r round) Render(_ context.Context, screen render.Screen) error {
	Draw(screen, r.Game)
	return nil
}

func (r round) Over() bool { return r.Game.Over() }

// OverTitle names the end: the core's secret read, or the trace complete.
func (r round) OverTitle() string {
	if r.Won() {
		return "ACCESO ROOT"
	}
	return "RASTREO COMPLETO"
}

// Result is the round's entry for the high score table.
func (r round) Result() records.Entry { return records.Entry{Score: r.Score()} }

var _ session.Ranked = round{}

func (r round) Summary() []overlay.Section {
	s := int(r.Elapsed().Seconds())
	return []overlay.Section{overlay.Text(
		fmt.Sprintf("Puntuación: %d", r.Score()),
		fmt.Sprintf("Nivel: %d/%d", r.Level()+1, len(levels)),
		fmt.Sprintf("Rastreo: %d%%", r.Trace()),
		fmt.Sprintf("Pistas: %d", r.Hints()),
		fmt.Sprintf("Tiempo: %02d:%02d", s/60, s%60),
	)}
}

// Slot is where a run is saved between levels (records.Slot).
type Slot interface {
	Load(v any) (bool, error)
	Store(v any) error
	Clear() error
}

// savedRound is a round whose run is saved: each cleared level stores a
// checkpoint and the end of the run, won or lost, deletes it.
type savedRound struct {
	round
	slot Slot
	err  *error // the last save error, shown at the end
}

func (r savedRound) Update(ctx context.Context, frame input.FrameInput) error {
	wasCleared := r.cleared
	if err := r.round.Update(ctx, frame); err != nil {
		return err
	}
	switch cp, ok := r.Checkpoint(); {
	case r.Over():
		*r.err = errors.Join(*r.err, r.slot.Clear())
	case ok && !wasCleared:
		*r.err = r.slot.Store(cp)
	}
	return nil
}

func (r savedRound) Summary() []overlay.Section {
	s := r.round.Summary()
	if *r.err != nil {
		s = append(s, overlay.Text("No se pudo guardar la partida."))
	}
	return s
}

// Module wires Hack into the runtime with injected clock, screen, input and RNG.
type Module struct {
	clock    clock.Clock
	screen   render.Screen
	provider input.Provider
	rng      Rand
	board    session.Board
	save     Slot
	resume   bool
}

func NewModule(clk clock.Clock, screen render.Screen, provider input.Provider, rng Rand) Module {
	return Module{clock: clk, screen: screen, provider: provider, rng: rng}
}

// WithSave saves the run in slot. With resume the first round continues
// the saved run, if any; otherwise it deletes it. Restart always begins a
// new game.
func (m Module) WithSave(slot Slot, resume bool) Module {
	m.save, m.resume = slot, resume
	return m
}

// rounds makes the session's rounds.
func (m Module) rounds() func() session.Round {
	resume := m.resume
	return func() session.Round {
		cfg := Config{Rand: m.rng}
		if m.save == nil {
			return round{NewGame(cfg)}
		}
		var err error
		if resume {
			var cp Checkpoint
			if ok, _ := m.save.Load(&cp); ok {
				cfg.Resume = &cp
			}
		} else {
			err = m.save.Clear()
		}
		resume = false
		return savedRound{round: round{NewGame(cfg)}, slot: m.save, err: &err}
	}
}

// WithBoard gives the module its game's high score table.
func (m Module) WithBoard(b session.Board) Module {
	m.board = b
	return m
}

func (m Module) Run(ctx context.Context) error {
	rt := runtime.New(runtime.Config{Clock: m.clock, Input: m.provider, Screen: m.screen})
	return rt.Run(ctx, session.New(m.rounds(), session.WithBoard(m.board)))
}
