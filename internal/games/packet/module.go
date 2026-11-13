package packet

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
	playfieldWidth = 64
	playerColumn   = 6
	startLink      = 1
)

// round adapts Game to session.Round: up and down (and W/S) switch links;
// when a frame holds several, the last one wins.
type round struct {
	*Game
}

func (r round) Update(_ context.Context, frame input.FrameInput) error {
	var in Input
	for _, a := range frame.Actions {
		switch a {
		case input.ActionUp:
			in.Switch = -1
		case input.ActionDown:
			in.Switch = 1
		}
	}
	r.Step(in)
	return nil
}

func (r round) Render(_ context.Context, screen render.Screen) error {
	Draw(screen, r.Game)
	return nil
}

func (r round) Over() bool { return r.Dropped() }

// OverTitle names the end of the round: a firewall dropped the packet.
func (r round) OverTitle() string { return "PACKET DROPPED" }

func (r round) Summary() []overlay.Section {
	return []overlay.Section{overlay.Text(
		fmt.Sprintf("Score: %d", r.Score()),
		fmt.Sprintf("Distance: %d", r.Distance()),
	)}
}

// Module wires Packet into the runtime with injected clock, screen, input and RNG.
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
		return NewGame(Config{
			Width:   playfieldWidth,
			PlayerX: playerColumn,
			Link:    startLink,
			Spawner: NewRandomSpawner(m.rng),
		})
	}
	rt := runtime.New(runtime.Config{Clock: m.clock, Input: m.provider, Screen: m.screen})
	return rt.Run(ctx, session.New(func() session.Round { return round{newGame()} }))
}
