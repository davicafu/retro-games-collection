package runtime

import (
	"context"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

type Screen = render.Screen

type Game interface {
	Init(context.Context) error
	Update(context.Context, input.FrameInput) error
	Render(context.Context, Screen) error
	ShouldQuit() bool
}

type Config struct {
	Clock  clock.Clock
	Input  input.Provider
	Screen Screen
}

type Runtime struct{ cfg Config }

func New(cfg Config) *Runtime { return &Runtime{cfg: cfg} }

func (r *Runtime) Run(ctx context.Context, game Game) error {
	if err := game.Init(ctx); err != nil {
		return err
	}

	for !game.ShouldQuit() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-r.cfg.Clock.Ticks():
			if err := game.Update(ctx, r.cfg.Input.Poll()); err != nil {
				return err
			}
			r.cfg.Screen.Clear()
			if err := game.Render(ctx, r.cfg.Screen); err != nil {
				return err
			}
			if err := r.cfg.Screen.Present(); err != nil {
				return err
			}
		}
	}

	return nil
}
