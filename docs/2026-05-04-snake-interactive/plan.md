# Task 5 Interactive Snake Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the first real interactive `play snake` terminal session using a real backend adapter while preserving runtime and gameplay decoupling.

**Architecture:** Keep Bubble Tea in the launcher only, and introduce a minimal tcell-backed terminal adapter for visible gameplay. The runtime continues to depend on abstract screen/input contracts, while Snake remains unaware of tcell and only consumes abstract actions and rendering surfaces.

**Tech Stack:** Go, Cobra, Bubble Tea, Lip Gloss, tcell, existing runtime/input/render contracts

---

## Planned File Structure

- `go.mod` — add tcell dependency
- `internal/platform/terminal/tcellscreen/screen.go` — real screen adapter over tcell
- `internal/platform/terminal/tcellscreen/input.go` — tcell event translation into runtime actions
- `internal/platform/terminal/tcellscreen/screen_test.go` — key translation tests using pure functions
- `internal/games/snake/render.go` — Snake render helper over abstract screen contract
- `internal/games/snake/module.go` — runtime-facing Snake module with injected input provider
- `internal/games/snake/module_test.go` — module behavior tests for quit and direction preservation
- `internal/app/app.go` — app wiring for real interactive `play snake`
- `cmd/root.go` — keep minimal `Execute() error` wiring
- `main.go` — unchanged error printing contract unless required by compilation

## Task 1: Add real terminal input adapter

**Files:**
- Modify: `go.mod`
- Create: `internal/platform/terminal/tcellscreen/input.go`
- Test: `internal/platform/terminal/tcellscreen/screen_test.go`

- [ ] **Step 1: Write the failing key translation test**

```go
// internal/platform/terminal/tcellscreen/screen_test.go
package tcellscreen

import (
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/gdamore/tcell/v2"
)

func TestMapKeyEvent_MapsArrowAndQuitKeys(t *testing.T) {
	tests := []struct {
		name string
		evt  *tcell.EventKey
		want input.Action
	}{
		{name: "up", evt: tcell.NewEventKey(tcell.KeyUp, 0, 0), want: input.ActionUp},
		{name: "left", evt: tcell.NewEventKey(tcell.KeyLeft, 0, 0), want: input.ActionLeft},
		{name: "w", evt: tcell.NewEventKey(tcell.KeyRune, 'w', 0), want: input.ActionUp},
		{name: "quit", evt: tcell.NewEventKey(tcell.KeyRune, 'q', 0), want: input.ActionQuit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := mapKeyEvent(tt.evt)
			if !ok {
				t.Fatalf("expected key to map")
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/terminal/tcellscreen -run TestMapKeyEvent_MapsArrowAndQuitKeys -v`
Expected: FAIL with missing tcell dependency or undefined `mapKeyEvent`.

- [ ] **Step 3: Write minimal implementation**

```go
// go.mod
module github.com/davidbarreiro/cli-games-collection

go 1.24.2

require (
	github.com/charmbracelet/bubbletea v1.3.4
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/gdamore/tcell/v2 v2.8.1
	github.com/spf13/cobra v1.9.1
)
```

```go
// internal/platform/terminal/tcellscreen/input.go
package tcellscreen

import (
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/gdamore/tcell/v2"
)

func mapKeyEvent(evt *tcell.EventKey) (input.Action, bool) {
	switch evt.Key() {
	case tcell.KeyUp:
		return input.ActionUp, true
	case tcell.KeyDown:
		return input.ActionDown, true
	case tcell.KeyLeft:
		return input.ActionLeft, true
	case tcell.KeyRight:
		return input.ActionRight, true
	case tcell.KeyEscape:
		return input.ActionQuit, true
	}

	switch evt.Rune() {
	case 'w', 'W':
		return input.ActionUp, true
	case 's', 'S':
		return input.ActionDown, true
	case 'a', 'A':
		return input.ActionLeft, true
	case 'd', 'D':
		return input.ActionRight, true
	case 'q', 'Q':
		return input.ActionQuit, true
	default:
		return "", false
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/platform/terminal/tcellscreen -run TestMapKeyEvent_MapsArrowAndQuitKeys -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/platform/terminal/tcellscreen/input.go internal/platform/terminal/tcellscreen/screen_test.go
git commit -m "feat: add terminal input adapter mapping"
```

## Task 2: Add real terminal screen adapter

**Files:**
- Create: `internal/platform/terminal/tcellscreen/screen.go`
- Test: `internal/platform/terminal/tcellscreen/screen_test.go`

- [ ] **Step 1: Add the failing screen lifecycle test**

```go
func TestNewSimulationScreen_ImplementsRenderContract(t *testing.T) {
	s := newSimulationScreen(10, 4)
	s.SetCell(2, 1, runeCell('X'))

	if got := s.contents[1][2]; got != 'X' {
		t.Fatalf("expected simulation screen cell to be X, got %q", got)
	}

	s.Clear()
	if got := s.contents[1][2]; got != ' ' {
		t.Fatalf("expected clear to reset cell, got %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/platform/terminal/tcellscreen -run TestNewSimulationScreen_ImplementsRenderContract -v`
Expected: FAIL with undefined `newSimulationScreen` or `runeCell`.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/platform/terminal/tcellscreen/screen.go
package tcellscreen

import (
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/gdamore/tcell/v2"
)

type Screen struct{ screen tcell.Screen }

func New() (*Screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.Clear()
	return &Screen{screen: s}, nil
}

func (s *Screen) Close() { s.screen.Fini() }
func (s *Screen) Size() (int, int) { return s.screen.Size() }
func (s *Screen) Clear() { s.screen.Clear() }
func (s *Screen) SetCell(x, y int, cell render.Cell) { s.screen.SetContent(x, y, cell.Rune, nil, tcell.StyleDefault) }
func (s *Screen) Present() error { s.screen.Show(); return nil }

type simulationScreen struct {
	width    int
	height   int
	contents [][]rune
}

func newSimulationScreen(width, height int) *simulationScreen {
	rows := make([][]rune, height)
	for y := range rows {
		rows[y] = make([]rune, width)
		for x := range rows[y] {
			rows[y][x] = ' '
		}
	}
	return &simulationScreen{width: width, height: height, contents: rows}
}

func (s *simulationScreen) Size() (int, int) { return s.width, s.height }
func (s *simulationScreen) Clear() {
	for y := range s.contents {
		for x := range s.contents[y] {
			s.contents[y][x] = ' '
		}
	}
}
func (s *simulationScreen) SetCell(x, y int, cell render.Cell) { s.contents[y][x] = cell.Rune }
func (s *simulationScreen) Present() error { return nil }

func runeCell(r rune) render.Cell { return render.Cell{Rune: r} }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/platform/terminal/tcellscreen -run TestNewSimulationScreen_ImplementsRenderContract -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/platform/terminal/tcellscreen/screen.go internal/platform/terminal/tcellscreen/screen_test.go
git commit -m "feat: add terminal screen adapter"
```

## Task 3: Wire Snake module to live input and graceful quit

**Files:**
- Create: `internal/games/snake/render.go`
- Create: `internal/games/snake/module.go`
- Modify: `internal/games/snake/game.go`
- Test: `internal/games/snake/module_test.go`

- [ ] **Step 1: Write the failing module behavior test**

```go
// internal/games/snake/module_test.go
package snake

import (
	"context"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/tcellscreen"
)

type scriptedProvider struct{ frames []input.FrameInput; idx int }

func (p *scriptedProvider) Poll() input.FrameInput {
	if p.idx >= len(p.frames) {
		return input.FrameInput{}
	}
	frame := p.frames[p.idx]
	p.idx++
	return frame
}

func TestModule_RunPreservesDirectionAndQuitsGracefully(t *testing.T) {
	manual := clock.NewManual()
	screen := tcellscreen.NewSimulationScreenForTest(12, 12)
	provider := &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionDown}}, {}, {Actions: []input.Action{input.ActionQuit}}}}
	module := NewModule(manual, screen, provider)

	errCh := make(chan error, 1)
	go func() { errCh <- module.Run(context.Background()) }()
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)

	if err := <-errCh; err != nil {
		t.Fatalf("expected graceful quit, got %v", err)
	}
	if len(screen.Snapshot()) == 0 {
		t.Fatal("expected visible frame output")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestModule_RunPreservesDirectionAndQuitsGracefully -v`
Expected: FAIL with missing `NewModule` or `NewSimulationScreenForTest`.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/games/snake/render.go
package snake

import "github.com/davidbarreiro/cli-games-collection/internal/engine/render"

func Draw(screen render.Screen, body []Position, food []Position) {
	for _, segment := range body {
		screen.SetCell(segment.X, segment.Y, render.Cell{Rune: '█'})
	}
	for _, item := range food {
		screen.SetCell(item.X, item.Y, render.Cell{Rune: '•'})
	}
}
```

```go
// internal/games/snake/game.go
package snake

func (g *Game) Body() []Position { return append([]Position(nil), g.body...) }
func (g *Game) Food() []Position { return append([]Position(nil), g.food...) }
func (g *Game) Direction() Direction { return g.direction }
```

```go
// internal/platform/terminal/tcellscreen/screen.go (append)
func NewSimulationScreenForTest(width, height int) *simulationScreen { return newSimulationScreen(width, height) }
func (s *simulationScreen) Snapshot() [][]rune { return s.contents }
```

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

type runtimeGame struct{ *Game }

func (g *runtimeGame) Init(context.Context) error { return nil }
func (g *runtimeGame) Update(_ context.Context, frame input.FrameInput) error {
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
			g.dead = true
			return nil
		}
	}
	return g.Game.Update(direction)
}
func (g *runtimeGame) Render(_ context.Context, screen runtime.Screen) error { Draw(screen, g.Body(), g.Food()); return nil }
func (g *runtimeGame) ShouldQuit() bool { return g.Dead() || len(g.Food()) == 0 }

type Module struct {
	clock    clock.Clock
	screen   render.Screen
	provider input.Provider
}

func NewModule(clk clock.Clock, screen render.Screen, provider input.Provider) Module {
	return Module{clock: clk, screen: screen, provider: provider}
}

func (m Module) Run(ctx context.Context) error {
	rt := runtime.New(runtime.Config{Clock: m.clock, Input: m.provider, Screen: m.screen})
	game := &runtimeGame{Game: NewGame(Config{Width: 12, Height: 12, Start: Position{X: 1, Y: 1}, Food: []Position{{X: 4, Y: 1}}})}
	return rt.Run(ctx, game)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/games/snake -run TestModule_RunPreservesDirectionAndQuitsGracefully -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/render.go internal/games/snake/module.go internal/games/snake/module_test.go internal/games/snake/game.go internal/platform/terminal/tcellscreen/screen.go
git commit -m "feat: add interactive snake module wiring"
```

## Task 4: Wire app path to real terminal gameplay

**Files:**
- Modify: `internal/app/app.go`
- Modify: `cmd/root.go`
- Modify: `main.go`
- Test: `internal/app/app_test.go`

- [ ] **Step 1: Write the failing app wiring test**

```go
// internal/app/app_test.go
package app

import (
	"context"
	"testing"
)

func TestRunGameWith_DelegatesSnakeRunner(t *testing.T) {
	called := false
	err := RunGameWith(context.Background(), "snake", func(context.Context) error {
		called = true
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected snake runner to be called")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/app -run TestRunGameWith_DelegatesSnakeRunner -v`
Expected: FAIL with undefined `RunGameWith`.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/app/app.go
package app

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/games/snake"
	"github.com/davidbarreiro/cli-games-collection/internal/launcher"
	"github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/tcellscreen"
)

type Launcher func() error
type Runner func(game string) error

func RunLauncherSelection() (string, error) {
	finalModel, err := tea.NewProgram(launcher.New([]string{"snake", "sokoban"})).Run()
	if err != nil {
		return "", err
	}
	model, _ := finalModel.(launcher.Model)
	return model.SelectedGame, nil
}

func RunLauncher() error {
	_, err := RunLauncherSelection()
	return err
}

func RunGame(game string) error { return RunGameWith(context.Background(), game, defaultSnakeRunner) }

func RunGameWith(ctx context.Context, game string, runSnake func(context.Context) error) error {
	switch game {
	case "snake":
		return runSnake(ctx)
	default:
		return fmt.Errorf("unknown game %q", game)
	}
}

func defaultSnakeRunner(ctx context.Context) error {
	screen, err := tcellscreen.New()
	if err != nil {
		return err
	}
	defer screen.Close()
	provider := tcellscreen.NewProvider(screen.RawScreen())
	return snake.NewModule(clock.NewTicker(120*time.Millisecond), screen, provider).Run(ctx)
	}
```

```go
// internal/platform/terminal/tcellscreen/input.go (append)
type Provider struct{ screen tcell.Screen }

func NewProvider(screen tcell.Screen) *Provider { return &Provider{screen: screen} }
func (p *Provider) Poll() input.FrameInput {
	evt := p.screen.PollEvent()
	key, ok := evt.(*tcell.EventKey)
	if !ok {
		return input.FrameInput{}
	}
	action, ok := mapKeyEvent(key)
	if !ok {
		return input.FrameInput{}
	}
	return input.FrameInput{Actions: []input.Action{action}}
}
```

```go
// internal/platform/terminal/tcellscreen/screen.go (append)
func (s *Screen) RawScreen() tcell.Screen { return s.screen }
```

```go
// internal/engine/clock/clock.go (append)
type Ticker struct{ ticker *time.Ticker }

func NewTicker(interval time.Duration) *Ticker { return &Ticker{ticker: time.NewTicker(interval)} }
func (t *Ticker) Ticks() <-chan time.Time { return t.ticker.C }
```

```go
// cmd/root.go
package cmd

import (
	"github.com/davidbarreiro/cli-games-collection/internal/app"
	"github.com/spf13/cobra"
)

func Execute() error {
	return NewRootCommand(app.RunLauncher, app.RunGame).Execute()
}
```

```go
// main.go
package main

import (
	"fmt"
	"os"

	"github.com/davidbarreiro/cli-games-collection/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/app -run TestRunGameWith_DelegatesSnakeRunner -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/app/app.go internal/app/app_test.go internal/engine/clock/clock.go internal/platform/terminal/tcellscreen/input.go internal/platform/terminal/tcellscreen/screen.go cmd/root.go main.go
git commit -m "feat: wire play snake to real terminal adapter"
```

## Self-Review

- **Spec coverage:** This revised plan replaces the old headless-only Task 5 path with a visible, live-input terminal flow over a tcell adapter while preserving Snake and runtime decoupling.
- **Placeholder scan:** No unresolved placeholders remain.
- **Type consistency:** `tcellscreen.Provider` implements `input.Provider`, `tcellscreen.Screen` implements `render.Screen`, and `snake.Module` accepts abstract contracts rather than direct framework types.

## Recommended execution order

1. Task 1 — Add real terminal input adapter
2. Task 2 — Add real terminal screen adapter
3. Task 3 — Wire Snake module to live input and graceful quit
4. Task 4 — Wire app path to real terminal gameplay
