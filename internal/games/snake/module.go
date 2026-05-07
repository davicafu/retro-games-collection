package snake

import (
	"context"
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/runtime"
)

type runtimeGame struct {
	*Game
	overlay   *OverlayModel
	forceQuit bool
	paused    bool
	selector  FoodSelector
}

func (g *runtimeGame) Init(context.Context) error { return nil }

func (g *runtimeGame) Update(_ context.Context, frame input.FrameInput) error {
	if g.overlay != nil {
		return g.updateOverlay(frame)
	}
	return g.updatePlay(frame)
}

func (g *runtimeGame) updateOverlay(frame input.FrameInput) error {
	for _, action := range frame.Actions {
		switch action {
		case input.ActionUp:
			g.overlay.MoveUp()
		case input.ActionDown:
			g.overlay.MoveDown()
		case input.ActionSelect:
			return g.resolveOverlay()
		case input.ActionPause:
			if g.paused {
				g.overlay = nil
				g.paused = false
				return nil
			}
		case input.ActionQuit:
			if g.paused {
				g.overlay = nil
				g.paused = false
				return nil
			}
			return g.resolveOverlay()
		}
	}
	return nil
}

func (g *runtimeGame) resolveOverlay() error {
	switch g.overlay.Resolve() {
	case OverlayActionContinue:
		g.overlay = nil
		g.paused = false
	case OverlayActionRestart:
		g.Game = NewGame(Config{
			Width: 12, Height: 12, Start: Position{X: 1, Y: 1},
			Food:     []Position{{X: 4, Y: 1}},
			Selector: g.selector,
		})
		g.overlay = nil
		g.paused = false
	case OverlayActionBackToMenu:
		g.forceQuit = true
		g.overlay = nil
		g.paused = false
		return ErrBackToMenu
	case OverlayActionQuit:
		g.forceQuit = true
		g.overlay = nil
		g.paused = false
	}
	return nil
}

func (g *runtimeGame) updatePlay(frame input.FrameInput) error {
	for _, action := range frame.Actions {
		switch action {
		case input.ActionUp:
			g.direction = DirectionUp
		case input.ActionDown:
			g.direction = DirectionDown
		case input.ActionLeft:
			g.direction = DirectionLeft
		case input.ActionRight:
			g.direction = DirectionRight
		case input.ActionQuit:
			g.forceQuit = true
			return nil
		case input.ActionPause:
			return g.activatePause()
		}
	}

	if err := g.Game.Update(g.direction); err != nil {
		return err
	}

	if g.Dead() {
		g.activateGameOverOverlay()
	}

	return nil
}

func (g *runtimeGame) activatePause() error {
	g.paused = true
	m := NewOverlayModel("PAUSED", []OverlayOption{
		{Label: "Continue", Action: OverlayActionContinue},
		{Label: "Restart", Action: OverlayActionRestart},
		{Label: "Menu", Action: OverlayActionBackToMenu},
		{Label: "Quit", Action: OverlayActionQuit},
	})
	g.overlay = &m
	return nil
}

func (g *runtimeGame) activateGameOverOverlay() {
	m := NewOverlayModel(
		fmt.Sprintf("Game Over\nScore: %d", g.Score()),
		[]OverlayOption{
			{Label: "Restart", Action: OverlayActionRestart},
			{Label: "Menu", Action: OverlayActionBackToMenu},
			{Label: "Quit", Action: OverlayActionQuit},
		},
	)
	g.overlay = &m
}

func (g *runtimeGame) Render(_ context.Context, screen runtime.Screen) error {
	layout := defaultPlayfieldLayout(12, 12)
	if shouldUseProjectedLayout(screen, layout) {
		scoreStr := fmt.Sprintf("Score: %d", g.Score())
		writeStr(screen, 1, 0, padRight(scoreStr, layout.totalWidth()-1))
	}

	Draw(screen, g.Body(), g.Food())
	if g.overlay != nil {
		DrawOverlay(screen, g.overlay)
	}
	return nil
}

func (g *runtimeGame) ShouldQuit() bool {
	if g.forceQuit {
		return true
	}
	if g.overlay != nil {
		return false
	}
	return g.Dead()
}

type Module struct {
	clock    clock.Clock
	screen   render.Screen
	provider input.Provider
	selector FoodSelector
}

var ErrBackToMenu = fmt.Errorf("back to menu")

func NewModule(clk clock.Clock, screen render.Screen, provider input.Provider, selector FoodSelector) Module {
	return Module{clock: clk, screen: screen, provider: provider, selector: selector}
}

func (m Module) Run(ctx context.Context) error {
	rt := runtime.New(runtime.Config{Clock: m.clock, Input: m.provider, Screen: m.screen})
	game := &runtimeGame{
		Game:     NewGame(Config{Width: 12, Height: 12, Start: Position{X: 1, Y: 1}, Food: []Position{{X: 4, Y: 1}}, Selector: m.selector}),
		selector: m.selector,
	}
	return rt.Run(ctx, game)
}
