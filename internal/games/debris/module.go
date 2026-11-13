package debris

import (
	"context"
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/runtime"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/session"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"
)

const (
	playfieldCols = 72
	playfieldRows = 18
)

// round adapts Game to session.Round: left and right (and A/D) turn, up (W)
// thrusts and space fires; a frame may hold several of them, and of two
// turns the last one wins.
type round struct {
	*Game
}

func (r round) Update(_ context.Context, frame input.FrameInput) error {
	var in Input
	for _, a := range frame.Actions {
		switch a {
		case input.ActionLeft:
			in.Turn = -1
		case input.ActionRight:
			in.Turn = 1
		case input.ActionUp:
			in.Thrust = true
		case input.ActionSelect:
			in.Fire = true
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

// OverTitle names the end of the round: the debris broke the drone's hull.
func (r round) OverTitle() string { return "HULL BREACHED" }

func (r round) Summary() []overlay.Section {
	return []overlay.Section{overlay.Text(
		fmt.Sprintf("Score: %d", r.Score()),
		fmt.Sprintf("Wave: %d", r.Wave()),
	)}
}

// Module wires Debris into the runtime with injected clock, screen, input and RNG.
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
	newGame := func() *Game {
		return NewGame(Config{Cols: playfieldCols, Rows: playfieldRows, Rand: m.rng})
	}
	rt := runtime.New(runtime.Config{Clock: m.clock, Input: m.provider, Screen: m.screen})
	return rt.Run(ctx, session.New(func() session.Round { return round{newGame()} }))
}
