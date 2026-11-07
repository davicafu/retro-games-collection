# Snake Playfield and Game Over Overlay Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add visible playfield boundaries, a live score display, and a reusable game-over overlay flow so Snake feels like a polished game and establishes reusable gameplay UX patterns.

**Architecture:** Keep game rules in the Snake domain, and implement playfield chrome plus overlay behavior in presentation/runtime-facing layers. The reusable part is the option-selection model and rendering pattern, while Snake owns when game over occurs and how restart/menu/quit intents are handled.

**Tech Stack:** Go, existing runtime/screen abstractions, tcell-backed gameplay path, Snake module/runtime wiring

---

## Planned File Structure

- `internal/games/snake/render.go` — render playfield border, score, snake, food, and overlay framing
- `internal/games/snake/module.go` — runtime-facing gameplay state, restart/menu/quit handling
- `internal/games/snake/module_test.go` — integration-style tests for chrome and overlay behavior
- `internal/games/snake/overlay.go` — reusable option-selection model for game-over overlay
- `internal/games/snake/overlay_test.go` — focused tests for selection and intention mapping
- `internal/app/app.go` — react to game-over overlay intents (`restart`, `back_to_menu`, `quit`)

## Task 1: Add reusable overlay selection model

**Files:**
- Create: `internal/games/snake/overlay.go`
- Test: `internal/games/snake/overlay_test.go`

- [ ] **Step 1: Write the failing overlay navigation test**

```go
// internal/games/snake/overlay_test.go
package snake

import "testing"

func TestOverlayModel_MoveSelectionAndResolveAction(t *testing.T) {
	overlay := NewOverlayModel("Game Over", []OverlayOption{
		{Label: "Play Again", Action: OverlayActionRestart},
		{Label: "Back to Menu", Action: OverlayActionBackToMenu},
		{Label: "Quit", Action: OverlayActionQuit},
	})

	overlay.MoveDown()
	if overlay.SelectedIndex() != 1 {
		t.Fatalf("expected selected index 1, got %d", overlay.SelectedIndex())
	}

	if got := overlay.Resolve(); got != OverlayActionBackToMenu {
		t.Fatalf("expected back-to-menu action, got %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestOverlayModel_MoveSelectionAndResolveAction -v`
Expected: FAIL with undefined `NewOverlayModel` or `OverlayActionBackToMenu`.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/games/snake/overlay.go
package snake

type OverlayAction string

const (
	OverlayActionRestart    OverlayAction = "restart"
	OverlayActionBackToMenu OverlayAction = "back_to_menu"
	OverlayActionQuit       OverlayAction = "quit"
)

type OverlayOption struct {
	Label  string
	Action OverlayAction
}

type OverlayModel struct {
	title   string
	options []OverlayOption
	index   int
}

func NewOverlayModel(title string, options []OverlayOption) OverlayModel {
	return OverlayModel{title: title, options: append([]OverlayOption(nil), options...)}
}

func (o *OverlayModel) MoveDown() {
	if o.index < len(o.options)-1 {
		o.index++
	}
}

func (o *OverlayModel) MoveUp() {
	if o.index > 0 {
		o.index--
	}
}

func (o OverlayModel) Resolve() OverlayAction {
	if len(o.options) == 0 {
		return ""
	}
	return o.options[o.index].Action
}

func (o OverlayModel) SelectedIndex() int { return o.index }
func (o OverlayModel) Title() string      { return o.title }
func (o OverlayModel) Options() []OverlayOption {
	return append([]OverlayOption(nil), o.options...)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/games/snake -run TestOverlayModel_MoveSelectionAndResolveAction -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/overlay.go internal/games/snake/overlay_test.go
git commit -m "feat: add reusable snake overlay selection model"
```

## Task 2: Render playfield border and score chrome

**Files:**
- Modify: `internal/games/snake/render.go`
- Modify: `internal/games/snake/module_test.go`

- [ ] **Step 1: Write the failing chrome render test**

```go
func TestDraw_RendersBorderAndScore(t *testing.T) {
	screen := tcellscreen.NewSimulationScreenForTest(14, 8)
	Draw(screen, []Position{{X: 1, Y: 1}}, []Position{{X: 3, Y: 1}}, 2)

	snapshot := screen.Snapshot()
	if snapshot[0][0] != '+' {
		t.Fatalf("expected top-left border corner, got %q", snapshot[0][0])
	}
	if snapshot[0][2] != 'S' {
		t.Fatalf("expected score header on top row, got %q", snapshot[0][2])
	}
	if snapshot[2][2] != '█' {
		t.Fatalf("expected snake body rendered inside playfield, got %q", snapshot[2][2])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestDraw_RendersBorderAndScore -v`
Expected: FAIL because `Draw` does not yet render chrome or accept score.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/games/snake/render.go
package snake

import (
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

func Draw(screen render.Screen, body []Position, food []Position, score int) {
	drawBorder(screen)
	drawText(screen, 2, 0, fmt.Sprintf("Score: %d", score))
	for _, segment := range body {
		screen.SetCell(segment.X+1, segment.Y+1, render.Cell{Rune: '█'})
	}
	for _, item := range food {
		screen.SetCell(item.X+1, item.Y+1, render.Cell{Rune: '•'})
	}
}

func drawBorder(screen render.Screen) {
	width, height := screen.Size()
	for x := 0; x < width; x++ {
		top := '-'
		bottom := '-'
		if x == 0 || x == width-1 {
			top = '+'
			bottom = '+'
		}
		screen.SetCell(x, 0, render.Cell{Rune: top})
		screen.SetCell(x, height-1, render.Cell{Rune: bottom})
	}
	for y := 1; y < height-1; y++ {
		screen.SetCell(0, y, render.Cell{Rune: '|'})
		screen.SetCell(width-1, y, render.Cell{Rune: '|'})
	}
}

func drawText(screen render.Screen, x, y int, text string) {
	for i, r := range text {
		screen.SetCell(x+i, y, render.Cell{Rune: r})
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/games/snake -run TestDraw_RendersBorderAndScore -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/render.go internal/games/snake/module_test.go
git commit -m "feat: render snake border and score chrome"
```

## Task 3: Add game-over overlay behavior to Snake module

**Files:**
- Modify: `internal/games/snake/module.go`
- Modify: `internal/games/snake/module_test.go`

- [ ] **Step 1: Write the failing game-over overlay test**

```go
func TestModule_RunShowsGameOverOverlayOnDeath(t *testing.T) {
	manual := clock.NewManual()
	screen := tcellscreen.NewSimulationScreenForTest(14, 8)
	provider := &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionRight}}, {Actions: []input.Action{input.ActionRight}}, {Actions: []input.Action{input.ActionRight}}, {Actions: []input.Action{input.ActionRight}}, {Actions: []input.Action{input.ActionQuit}}}}
	module := NewModule(manual, screen, provider)

	errCh := make(chan error, 1)
	go func() { errCh <- module.Run(context.Background()) }()
	for i := 0; i < 5; i++ {
		manual.Advance(time.Millisecond)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("expected graceful exit from overlay flow, got %v", err)
	}

	snapshot := screen.Snapshot()
	if snapshot[3][4] != 'G' {
		t.Fatalf("expected game over overlay title, got %q", snapshot[3][4])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestModule_RunShowsGameOverOverlayOnDeath -v`
Expected: FAIL because module exits immediately on death.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/games/snake/module.go
package snake

import (
	"context"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/runtime"
)

type runtimeGame struct {
	*Game
	overlayActive bool
	overlay       OverlayModel
	selected      OverlayAction
}

func (g *runtimeGame) Init(context.Context) error { return nil }

func (g *runtimeGame) Update(_ context.Context, frame input.FrameInput) error {
	if g.overlayActive {
		for _, action := range frame.Actions {
			switch action {
			case input.ActionUp:
				g.overlay.MoveUp()
			case input.ActionDown:
				g.overlay.MoveDown()
			case input.ActionQuit, input.ActionRight:
				g.selected = g.overlay.Resolve()
				g.dead = true
				return nil
			}
		}
		return nil
	}

	direction := g.Direction()
	for _, action := range frame.Actions {
		switch action {
		case input.ActionUp:
			direction = DirectionUp
		case input.ActionDown:
			direction = DirectionDown
		case input.ActionLeft:
			direction = DirectionLeft
		case input.ActionRight:
			direction = DirectionRight
		case input.ActionQuit:
			g.selected = OverlayActionQuit
			g.dead = true
			return nil
		}
	}

	if err := g.Game.Update(direction); err != nil {
		return err
	}
	if g.Game.Dead() {
		g.overlayActive = true
		g.overlay = NewOverlayModel("Game Over", []OverlayOption{
			{Label: "Play Again", Action: OverlayActionRestart},
			{Label: "Back to Menu", Action: OverlayActionBackToMenu},
			{Label: "Quit", Action: OverlayActionQuit},
		})
		g.dead = false
	}
	return nil
}

func (g *runtimeGame) Render(_ context.Context, screen runtime.Screen) error {
	Draw(screen, g.Body(), g.Food(), g.Score())
	if g.overlayActive {
		drawOverlay(screen, g.overlay, g.Score())
	}
	return nil
}

func (g *runtimeGame) ShouldQuit() bool { return g.dead }

func drawOverlay(screen render.Screen, overlay OverlayModel, score int) {
	drawText(screen, 4, 3, overlay.Title())
	drawText(screen, 4, 4, fmt.Sprintf("Final Score: %d", score))
	for i, option := range overlay.Options() {
		prefix := ' '
		if i == overlay.SelectedIndex() {
			prefix = '>'
		}
		screen.SetCell(4, 6+i, render.Cell{Rune: prefix})
		drawText(screen, 6, 6+i, option.Label)
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/games/snake -run TestModule_RunShowsGameOverOverlayOnDeath -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/module.go internal/games/snake/module_test.go
git commit -m "feat: add snake game over overlay flow"
```

## Task 4: Route restart / menu / quit overlay intentions

**Files:**
- Modify: `internal/games/snake/module.go`
- Modify: `internal/app/app.go`
- Test: `internal/app/app_test.go`

- [ ] **Step 1: Write the failing restart routing test**

```go
// internal/app/app_test.go
func TestRunGameWith_RestartsWhenSnakeRequestsRestart(t *testing.T) {
	count := 0
	err := RunGameWith(context.Background(), "snake", func(context.Context) error {
		count++
		if count == 1 {
			return snake.ErrRestartRequested
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected restart flow to finish cleanly, got %v", err)
	}
	if count != 2 {
		t.Fatalf("expected snake runner to execute twice, got %d", count)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/app -run TestRunGameWith_RestartsWhenSnakeRequestsRestart -v`
Expected: FAIL because restart/menu/quit overlay outcomes are not wired yet.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/games/snake/module.go (append)
import "errors"

var (
	ErrRestartRequested = errors.New("snake restart requested")
	ErrBackToMenu       = errors.New("snake back to menu requested")
)

func (g *runtimeGame) overlayResult() error {
	switch g.selected {
	case OverlayActionRestart:
		return ErrRestartRequested
	case OverlayActionBackToMenu:
		return ErrBackToMenu
	default:
		return nil
	}
}

func (m Module) Run(ctx context.Context) error {
	rt := runtime.New(runtime.Config{Clock: m.clock, Input: m.provider, Screen: m.screen})
	game := &runtimeGame{Game: NewGame(Config{Width: 12, Height: 12, Start: Position{X: 1, Y: 1}, Food: []Position{{X: 4, Y: 1}}})}
	if err := rt.Run(ctx, game); err != nil {
		return err
	}
	return game.overlayResult()
}
```

```go
// internal/app/app.go
func RunGameWith(ctx context.Context, game string, runSnake func(context.Context) error) error {
	switch game {
	case "snake":
		for {
			err := runSnake(ctx)
			switch err {
			case nil:
				return nil
			case snake.ErrRestartRequested:
				continue
			case snake.ErrBackToMenu:
				return nil
			default:
				return err
			}
		}
	default:
		return fmt.Errorf("game %q is not implemented", game)
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/app -run TestRunGameWith_RestartsWhenSnakeRequestsRestart -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/module.go internal/app/app.go internal/app/app_test.go
git commit -m "feat: route snake overlay actions through app flow"
```

## Self-Review

- **Spec coverage:** This plan covers visible borders, score, reusable overlay selection, game-over presentation, and action routing for restart/menu/quit.
- **Placeholder scan:** No unresolved placeholders remain.
- **Type consistency:** `OverlayAction`, `OverlayOption`, `OverlayModel`, and the game-over action routing contracts are consistent across overlay, module, and app tasks.

## Recommended execution order

1. Task 1 — Add reusable overlay selection model
2. Task 2 — Render playfield border and score chrome
3. Task 3 — Add game-over overlay behavior to Snake module
4. Task 4 — Route restart / menu / quit overlay intentions
