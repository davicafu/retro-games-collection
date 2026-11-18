package breach

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

const (
	playfieldWidth  = 40
	playfieldHeight = 16
)

// round adapts Game to session.Round: Left/Right move the injector, Up and
// Select launch the virus.
type round struct {
	*Game
}

func (r round) Update(_ context.Context, frame input.FrameInput) error {
	var in Input
	for _, a := range frame.Actions {
		switch a {
		case input.ActionLeft:
			in.Move--
		case input.ActionRight:
			in.Move++
		case input.ActionUp, input.ActionSelect:
			in.Launch = true
		}
	}
	in.Move = sign(in.Move)
	r.Step(in)
	return nil
}

func (r round) Render(_ context.Context, screen render.Screen) error {
	Draw(screen, r.Game)
	return nil
}

func (r round) Over() bool { return r.Game.Over() }

// OverTitle names the end of the round: a breach is a win.
func (r round) OverTitle() string {
	if r.Won() {
		return "FIREWALL BREACHED"
	}
	return "Game Over"
}

// Result is the round's entry for the high score table.
func (r round) Result() records.Entry { return records.Entry{Score: r.Score()} }

var _ session.Ranked = round{}

func (r round) Summary() []overlay.Section {
	return []overlay.Section{overlay.Text(
		fmt.Sprintf("Score: %d", r.Score()),
		fmt.Sprintf("Level: %d/%d", r.Level(), r.LevelCount()),
	)}
}

// Module wires Breach into the runtime with injected clock, screen, input and RNG.
type Module struct {
	clock    clock.Clock
	screen   render.Screen
	provider input.Provider
	rng      Rand
	board    session.Board
}

func NewModule(clk clock.Clock, screen render.Screen, provider input.Provider, rng Rand) Module {
	return Module{clock: clk, screen: screen, provider: provider, rng: rng}
}

// WithBoard gives the module its game's high score table.
func (m Module) WithBoard(b session.Board) Module {
	m.board = b
	return m
}

func (m Module) Run(ctx context.Context) error {
	newGame := func() *Game {
		return NewGame(Config{
			Width:  playfieldWidth,
			Height: playfieldHeight,
			Levels: Campaign,
			Rand:   m.rng,
		})
	}
	rt := runtime.New(runtime.Config{Clock: m.clock, Input: m.provider, Screen: m.screen})
	return rt.Run(ctx, session.New(func() session.Round { return round{newGame()} }, session.WithBoard(m.board)))
}
