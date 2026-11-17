package hack

import (
	"context"
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
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
		return "ROOT ACCESS"
	}
	return "TRACE COMPLETE"
}

func (r round) Summary() []overlay.Section {
	s := int(r.Elapsed().Seconds())
	return []overlay.Section{overlay.Text(
		fmt.Sprintf("Trace: %d%%", r.Trace()),
		fmt.Sprintf("Time: %02d:%02d", s/60, s%60),
	)}
}

// Module wires Hack into the runtime with injected clock, screen, input and RNG.
type Module struct {
	clock    clock.Clock
	screen   render.Screen
	provider input.Provider
	rng      Rand
}

func NewModule(clk clock.Clock, screen render.Screen, provider input.Provider, rng Rand) Module {
	return Module{clock: clk, screen: screen, provider: provider, rng: rng}
}

func (m Module) Run(ctx context.Context) error {
	rt := runtime.New(runtime.Config{Clock: m.clock, Input: m.provider, Screen: m.screen})
	return rt.Run(ctx, session.New(func() session.Round { return round{NewGame(Config{Rand: m.rng})} }))
}
