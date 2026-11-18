package snake

import (
	"context"
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/records"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/runtime"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/session"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"
)

// round adapts Game to session.Round: arrow keys steer, every tick moves.
type round struct {
	*Game
}

func (r round) Update(_ context.Context, frame input.FrameInput) error {
	for _, action := range frame.Actions {
		switch action {
		case input.ActionUp:
			r.direction = DirectionUp
		case input.ActionDown:
			r.direction = DirectionDown
		case input.ActionLeft:
			r.direction = DirectionLeft
		case input.ActionRight:
			r.direction = DirectionRight
		}
	}
	return r.Game.Update(r.direction)
}

func (r round) Render(_ context.Context, screen render.Screen) error {
	layout := defaultPlayfieldLayout(12, 12)
	if shouldUseProjectedLayout(screen, layout) {
		scoreStr := fmt.Sprintf("Score: %d", r.Score())
		render.WriteText(screen, 1, 0, render.PadRight(scoreStr, layout.totalWidth()-1))
	}
	Draw(screen, r.Body(), r.Food())
	return nil
}

func (r round) Over() bool { return r.Dead() }

// Result is the round's entry for the high score table.
func (r round) Result() records.Entry { return records.Entry{Score: r.Score()} }

var _ session.Ranked = round{}

func (r round) Summary() []overlay.Section {
	return []overlay.Section{overlay.Text(fmt.Sprintf("Score: %d", r.Score()))}
}

type Module struct {
	clock    clock.Clock
	screen   render.Screen
	provider input.Provider
	selector FoodSelector
	board    session.Board
}

func NewModule(clk clock.Clock, screen render.Screen, provider input.Provider, selector FoodSelector) Module {
	return Module{clock: clk, screen: screen, provider: provider, selector: selector}
}

// WithBoard gives the module its game's high score table.
func (m Module) WithBoard(b session.Board) Module {
	m.board = b
	return m
}

func (m Module) Run(ctx context.Context) error {
	rt := runtime.New(runtime.Config{Clock: m.clock, Input: m.provider, Screen: m.screen})
	return rt.Run(ctx, session.New(func() session.Round {
		return round{NewGame(Config{
			Width: 12, Height: 12, Start: Position{X: 1, Y: 1},
			Food:     []Position{{X: 4, Y: 1}},
			Selector: m.selector,
		})}
	}, session.WithBoard(m.board)))
}
