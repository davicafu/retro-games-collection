package defrag

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

// round adapts Game to session.Round: left and right (and A/D) shift, up (W)
// rotates, down (S) drops faster, space drops at once and G turns the guides
// on and off. The guides setting is shared by every round of a session.
type round struct {
	*Game
	guides *bool
}

func (r round) Update(_ context.Context, frame input.FrameInput) error {
	var in Input
	for _, a := range frame.Actions {
		switch a {
		case input.ActionLeft:
			in.Shift = -1
		case input.ActionRight:
			in.Shift = 1
		case input.ActionUp:
			in.Rotate = true
		case input.ActionDown:
			in.SoftDrop = true
		case input.ActionSelect:
			in.HardDrop = true
		case input.ActionGuides:
			*r.guides = !*r.guides
		}
	}
	r.Step(in)
	return nil
}

func (r round) Render(_ context.Context, screen render.Screen) error {
	Draw(screen, r.Game, *r.guides)
	return nil
}

func (r round) Over() bool { return r.Game.Over() }

// OverTitle names the end of the round: no room left for a fragment.
func (r round) OverTitle() string { return "DISK FULL" }

func (r round) Summary() []overlay.Section {
	return []overlay.Section{overlay.Text(
		fmt.Sprintf("Score: %d", r.Score()),
		fmt.Sprintf("Level: %d", r.Level()),
		fmt.Sprintf("Sectors: %d", r.Sectors()),
	)}
}

// Module wires Defrag into the runtime with injected clock, screen, input and RNG.
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
	return rt.Run(ctx, session.New(newRounds(m.rng)))
}

// newRounds makes a session's rounds, with the guides on and the setting
// kept from one round to the next.
func newRounds(rng Rand) func() session.Round {
	guides := true
	return func() session.Round { return round{Game: NewGame(Config{Rand: rng}), guides: &guides} }
}
