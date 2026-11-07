package drone

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
	playfieldWidth  = 48
	playfieldHeight = 14
	droneColumn     = 8
)

// round adapts Game to session.Round: Up and Select thrust.
type round struct {
	*Game
}

func (r round) Update(_ context.Context, frame input.FrameInput) error {
	thrust := false
	for _, a := range frame.Actions {
		if a == input.ActionUp || a == input.ActionSelect {
			thrust = true
		}
	}
	r.Step(thrust)
	return nil
}

func (r round) Render(_ context.Context, screen render.Screen) error {
	Draw(screen, r.Game)
	return nil
}

func (r round) Over() bool { return r.Dead() }

func (r round) Summary() []overlay.Section {
	return []overlay.Section{overlay.Text(fmt.Sprintf("Score: %d", r.Score()))}
}

// Module wires Drone into the runtime with injected clock, screen, input and RNG.
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
			Height:  playfieldHeight,
			DroneX:  droneColumn,
			Spawner: NewRandomSpawner(m.rng),
		})
	}
	rt := runtime.New(runtime.Config{Clock: m.clock, Input: m.provider, Screen: m.screen})
	return rt.Run(ctx, session.New(func() session.Round { return round{newGame()} }))
}
