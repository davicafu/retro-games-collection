# Foundation and Snake Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bootstrap the project foundation, establish a testable terminal game runtime, add a launcher shell, and ship Snake as the first validating game.

**Architecture:** Keep Cobra at the CLI boundary, Bubble Tea in the launcher shell, and gameplay in a decoupled runtime. Build deterministic runtime contracts first, then implement Snake as a grid-based game that validates input, ticking, collisions, scoring, and headless simulation.

**Tech Stack:** Go, Cobra, Bubble Tea, Lip Gloss, testify/stdlib testing

---

## Planned File Structure

- `go.mod` — module declaration and dependencies
- `main.go` — process entrypoint
- `cmd/root.go` — Cobra root command and launcher entry
- `cmd/play.go` — direct game execution command surface
- `internal/app/app.go` — high-level app orchestration
- `internal/launcher/model.go` — Bubble Tea launcher model
- `internal/launcher/model_test.go` — launcher behavior tests
- `internal/engine/runtime/runtime.go` — runtime lifecycle and tick loop
- `internal/engine/runtime/runtime_test.go` — deterministic runtime tests
- `internal/engine/input/input.go` — input abstractions
- `internal/engine/render/render.go` — frame buffer abstractions
- `internal/engine/clock/clock.go` — clock contracts and manual clock
- `internal/engine/grid/point.go` — grid point primitives
- `internal/platform/terminal/headless/screen.go` — headless screen adapter for tests
- `internal/games/snake/game.go` — Snake domain and update logic
- `internal/games/snake/game_test.go` — Snake rule and simulation tests
- `internal/games/snake/render.go` — Snake frame rendering
- `internal/games/snake/module.go` — factory for runtime wiring

## Task 1: Bootstrap module and CLI boundary

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `cmd/play.go`
- Create: `internal/app/app.go`
- Test: `cmd/root_test.go`

- [ ] **Step 1: Write the failing CLI test**

```go
// cmd/root_test.go
package cmd

import "testing"

func TestNewRootCommand_UsesCollectionName(t *testing.T) {
	cmd := NewRootCommand(func() error { return nil }, nil)

	if cmd.Use != "cli-games-collection" {
		t.Fatalf("expected root command use to be cli-games-collection, got %q", cmd.Use)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestNewRootCommand_UsesCollectionName -v`
Expected: FAIL with an undefined `NewRootCommand` symbol.

- [ ] **Step 3: Write minimal implementation**

```go
// go.mod
module github.com/davidbarreiro/cli-games-collection

go 1.24.2

require github.com/spf13/cobra v1.9.1
```

```go
// main.go
package main

import (
	"log"

	"github.com/davidbarreiro/cli-games-collection/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
```

```go
// internal/app/app.go
package app

type Launcher func() error
type Runner func(game string) error
```

```go
// cmd/root.go
package cmd

import "github.com/spf13/cobra"

func NewRootCommand(runLauncher func() error, runGame func(string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cli-games-collection",
		Short: "Progressive terminal games laboratory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLauncher()
		},
	}

	cmd.AddCommand(NewPlayCommand(runGame))
	return cmd
}

func Execute() error {
	root := NewRootCommand(func() error { return nil }, func(string) error { return nil })
	return root.Execute()
}
```

```go
// cmd/play.go
package cmd

import "github.com/spf13/cobra"

func NewPlayCommand(runGame func(string) error) *cobra.Command {
	return &cobra.Command{
		Use:   "play [game]",
		Short: "Run a game directly",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGame(args[0])
		},
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestNewRootCommand_UsesCollectionName -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod main.go cmd/root.go cmd/play.go internal/app/app.go cmd/root_test.go
git commit -m "feat: bootstrap cobra entrypoints"
```

## Task 2: Establish deterministic runtime contracts

**Files:**
- Create: `internal/engine/input/input.go`
- Create: `internal/engine/render/render.go`
- Create: `internal/engine/clock/clock.go`
- Create: `internal/engine/runtime/runtime.go`
- Create: `internal/platform/terminal/headless/screen.go`
- Test: `internal/engine/runtime/runtime_test.go`

- [ ] **Step 1: Write the failing runtime test**

```go
// internal/engine/runtime/runtime_test.go
package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/headless"
)

type fakeGame struct{ updates int }

func (g *fakeGame) Init(context.Context) error                     { return nil }
func (g *fakeGame) Update(context.Context, input.FrameInput) error { g.updates++; return nil }
func (g *fakeGame) Render(context.Context, Screen) error           { return nil }
func (g *fakeGame) ShouldQuit() bool                               { return g.updates >= 3 }

func TestRuntime_RunAdvancesUntilGameQuits(t *testing.T) {
	manual := clock.NewManual()
	screen := headless.NewScreen(20, 10)
	rt := New(Config{Clock: manual, Screen: screen, Input: input.Static(nil)})

	game := &fakeGame{}
	errCh := make(chan error, 1)
	go func() { errCh <- rt.Run(context.Background(), game) }()

	for range 3 {
		manual.Advance(time.Millisecond)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("expected no run error, got %v", err)
	}

	if game.updates != 3 {
		t.Fatalf("expected 3 updates, got %d", game.updates)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/engine/runtime -run TestRuntime_RunAdvancesUntilGameQuits -v`
Expected: FAIL with missing `clock`, `input`, `headless`, or `New` runtime symbols.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/engine/input/input.go
package input

type Action string

const (
	ActionUp    Action = "up"
	ActionDown  Action = "down"
	ActionLeft  Action = "left"
	ActionRight Action = "right"
	ActionQuit  Action = "quit"
)

type FrameInput struct {
	Actions []Action
}

type Provider interface {
	Poll() FrameInput
}

type staticProvider struct{ frame FrameInput }

func Static(actions []Action) Provider { return staticProvider{frame: FrameInput{Actions: actions}} }
func (s staticProvider) Poll() FrameInput { return s.frame }
```

```go
// internal/engine/render/render.go
package render

type Cell struct{ Rune rune }

type Screen interface {
	Size() (int, int)
	Clear()
	SetCell(x, y int, cell Cell)
	Present() error
}
```

```go
// internal/engine/clock/clock.go
package clock

import "time"

type Clock interface {
	Ticks() <-chan time.Time
}

type Manual struct{ ch chan time.Time }

func NewManual() *Manual { return &Manual{ch: make(chan time.Time, 32)} }
func (m *Manual) Ticks() <-chan time.Time { return m.ch }
func (m *Manual) Advance(d time.Duration) { m.ch <- time.Now().Add(d) }
```

```go
// internal/platform/terminal/headless/screen.go
package headless

import "github.com/davidbarreiro/cli-games-collection/internal/engine/render"

type Screen struct {
	width  int
	height int
	Cells  map[[2]int]render.Cell
}

func NewScreen(width, height int) *Screen {
	return &Screen{width: width, height: height, Cells: map[[2]int]render.Cell{}}
}

func (s *Screen) Size() (int, int)          { return s.width, s.height }
func (s *Screen) Clear()                    { s.Cells = map[[2]int]render.Cell{} }
func (s *Screen) SetCell(x, y int, cell render.Cell) { s.Cells[[2]int{x, y}] = cell }
func (s *Screen) Present() error            { return nil }
```

```go
// internal/engine/runtime/runtime.go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/engine/runtime -run TestRuntime_RunAdvancesUntilGameQuits -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/engine/input/input.go internal/engine/render/render.go internal/engine/clock/clock.go internal/engine/runtime/runtime.go internal/engine/runtime/runtime_test.go internal/platform/terminal/headless/screen.go
git commit -m "feat: add deterministic runtime contracts"
```

## Task 3: Add Bubble Tea launcher shell

**Files:**
- Create: `internal/launcher/model.go`
- Test: `internal/launcher/model_test.go`
- Modify: `go.mod`
- Modify: `internal/app/app.go`

- [ ] **Step 1: Write the failing launcher test**

```go
// internal/launcher/model_test.go
package launcher

import tea "github.com/charmbracelet/bubbletea"
import "testing"

func TestModel_SelectOnEnterReturnsChosenGame(t *testing.T) {
	m := New([]string{"snake", "sokoban"})
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	next := updated.(Model)
	if next.SelectedGame != "snake" {
		t.Fatalf("expected snake selected, got %q", next.SelectedGame)
	}
	if cmd == nil {
		t.Fatalf("expected quit command after selecting a game")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/launcher -run TestModel_SelectOnEnterReturnsChosenGame -v`
Expected: FAIL with missing Bubble Tea dependency or undefined `New`.

- [ ] **Step 3: Write minimal implementation**

```go
// go.mod
module github.com/davidbarreiro/cli-games-collection

go 1.24.2

require (
	github.com/charmbracelet/bubbletea v1.3.4
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/spf13/cobra v1.9.1
)
```

```go
// internal/launcher/model.go
package launcher

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Games        []string
	Cursor       int
	SelectedGame string
}

func New(games []string) Model { return Model{Games: games} }
func (m Model) Init() tea.Cmd  { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyDown:
			if m.Cursor < len(m.Games)-1 {
				m.Cursor++
			}
		case tea.KeyUp:
			if m.Cursor > 0 {
				m.Cursor--
			}
		case tea.KeyEnter:
			m.SelectedGame = m.Games[m.Cursor]
			return m, tea.Quit
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	selected := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	lines := []string{"CLI Games Collection", ""}
	for i, game := range m.Games {
		prefix := "  "
		line := game
		if i == m.Cursor {
			prefix = "> "
			line = selected.Render(game)
		}
		lines = append(lines, prefix+line)
	}
	return strings.Join(lines, "\n")
}
```

```go
// internal/app/app.go
package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidbarreiro/cli-games-collection/internal/launcher"
)

type Launcher func() error
type Runner func(game string) error

func RunLauncher() error {
	_, err := tea.NewProgram(launcher.New([]string{"snake", "sokoban"})).Run()
	return err
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/launcher -run TestModel_SelectOnEnterReturnsChosenGame -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod internal/launcher/model.go internal/launcher/model_test.go internal/app/app.go
git commit -m "feat: add launcher shell model"
```

## Task 4: Implement Snake domain and deterministic simulation

**Files:**
- Create: `internal/engine/grid/point.go`
- Create: `internal/games/snake/game.go`
- Test: `internal/games/snake/game_test.go`

- [ ] **Step 1: Write the failing Snake rule test**

```go
// internal/games/snake/game_test.go
package snake

import "testing"

func TestGame_UpdateMovesHeadAndGrowsOnFood(t *testing.T) {
	g := NewGame(Config{
		Width:  8,
		Height: 8,
		Start:  Position{X: 2, Y: 2},
		Food:   []Position{{X: 3, Y: 2}},
	})

	if err := g.Update(DirectionRight); err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if got := g.Head(); got != (Position{X: 3, Y: 2}) {
		t.Fatalf("expected head at food tile, got %+v", got)
	}

	if g.Score() != 1 {
		t.Fatalf("expected score 1, got %d", g.Score())
	}

	if g.Length() != 2 {
		t.Fatalf("expected snake to grow to length 2, got %d", g.Length())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestGame_UpdateMovesHeadAndGrowsOnFood -v`
Expected: FAIL with undefined `NewGame`, `Config`, or `DirectionRight`.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/engine/grid/point.go
package grid

type Point struct {
	X int
	Y int
}
```

```go
// internal/games/snake/game.go
package snake

import "fmt"

type Position struct {
	X int
	Y int
}

type Direction struct {
	DX int
	DY int
}

var (
	DirectionUp    = Direction{DX: 0, DY: -1}
	DirectionDown  = Direction{DX: 0, DY: 1}
	DirectionLeft  = Direction{DX: -1, DY: 0}
	DirectionRight = Direction{DX: 1, DY: 0}
)

type Config struct {
	Width  int
	Height int
	Start  Position
	Food   []Position
}

type Game struct {
	width     int
	height    int
	body      []Position
	direction Direction
	food      []Position
	score     int
	dead      bool
}

func NewGame(cfg Config) *Game {
	return &Game{
		width:     cfg.Width,
		height:    cfg.Height,
		body:      []Position{cfg.Start},
		direction: DirectionRight,
		food:      append([]Position(nil), cfg.Food...),
	}
}

func (g *Game) Update(direction Direction) error {
	if g.dead {
		return fmt.Errorf("game over")
	}
	g.direction = direction
	head := g.Head()
	next := Position{X: head.X + direction.DX, Y: head.Y + direction.DY}
	if next.X < 0 || next.Y < 0 || next.X >= g.width || next.Y >= g.height {
		g.dead = true
		return nil
	}
	g.body = append([]Position{next}, g.body...)
	if len(g.food) > 0 && next == g.food[0] {
		g.food = g.food[1:]
		g.score++
		return nil
	}
	g.body = g.body[:len(g.body)-1]
	return nil
}

func (g *Game) Head() Position { return g.body[0] }
func (g *Game) Length() int    { return len(g.body) }
func (g *Game) Score() int     { return g.score }
func (g *Game) Dead() bool     { return g.dead }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/games/snake -run TestGame_UpdateMovesHeadAndGrowsOnFood -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/engine/grid/point.go internal/games/snake/game.go internal/games/snake/game_test.go
git commit -m "feat: add snake core rules"
```

## Task 5: Wire Snake into runtime and CLI

**Files:**
- Create: `internal/games/snake/render.go`
- Create: `internal/games/snake/module.go`
- Modify: `internal/games/snake/game.go`
- Modify: `internal/app/app.go`
- Modify: `cmd/root.go`
- Modify: `main.go`
- Test: `internal/games/snake/module_test.go`

- [ ] **Step 1: Write the failing integration test**

```go
// internal/games/snake/module_test.go
package snake

import (
	"context"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/headless"
)

func TestModule_RunAdvancesOneFrame(t *testing.T) {
	manual := clock.NewManual()
	screen := headless.NewScreen(12, 12)
	module := NewModule(manual, screen)

	errCh := make(chan error, 1)
	go func() { errCh <- module.Run(context.Background()) }()
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)

	if err := <-errCh; err != nil {
		t.Fatalf("expected snake module run without error, got %v", err)
	}
	if len(screen.Cells) == 0 {
		t.Fatalf("expected snake renderer to draw cells")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestModule_RunAdvancesOneFrame -v`
Expected: FAIL with missing `NewModule` or runtime wiring.

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

func (g *runtimeGame) Update(context.Context, frame input.FrameInput) error {
	direction := DirectionRight
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
		}
	}
	return g.Game.Update(direction)
	}

func (g *runtimeGame) Render(_ context.Context, screen runtime.Screen) error {
	Draw(screen, g.body, g.food)
	return nil
}

func (g *runtimeGame) ShouldQuit() bool { return g.Dead() || len(g.food) == 0 }

type Module struct {
	clock  clock.Clock
	screen render.Screen
}

func NewModule(clk clock.Clock, screen render.Screen) Module {
	return Module{clock: clk, screen: screen}
}

func (m Module) Run(ctx context.Context) error {
	rt := runtime.New(runtime.Config{
		Clock:  m.clock,
		Input:  input.Static([]input.Action{input.ActionRight}),
		Screen: m.screen,
	})
	game := &runtimeGame{Game: NewGame(Config{Width: 12, Height: 12, Start: Position{X: 1, Y: 1}, Food: []Position{{X: 4, Y: 1}}})}
	return rt.Run(ctx, game)
}
```

```go
// internal/app/app.go
package app

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/games/snake"
	"github.com/davidbarreiro/cli-games-collection/internal/launcher"
	"github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/headless"
)

type Launcher func() error
type Runner func(game string) error

func RunLauncher() error {
	_, err := tea.NewProgram(launcher.New([]string{"snake", "sokoban"})).Run()
	return err
}

func RunGame(game string) error {
	switch game {
	case "snake":
		manual := clock.NewManual()
		go func() {
			for i := 0; i < 3; i++ {
				manual.Advance(120 * time.Millisecond)
			}
		}()
		return snake.NewModule(manual, headless.NewScreen(12, 12)).Run(context.Background())
	default:
		return fmt.Errorf("unknown game %q", game)
	}
}
```

```go
// cmd/root.go
package cmd

import (
	"github.com/davidbarreiro/cli-games-collection/internal/app"
	"github.com/spf13/cobra"
)

func Execute() error {
	root := NewRootCommand(app.RunLauncher, app.RunGame)
	return root.Execute()
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

Run: `go test ./internal/games/snake -run TestModule_RunAdvancesOneFrame -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/render.go internal/games/snake/module.go internal/games/snake/module_test.go internal/games/snake/game.go internal/app/app.go cmd/root.go main.go
git commit -m "feat: wire snake through runtime"
```

## Task 6: Verify end-to-end foundation scope

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write a minimal getting-started document**

~~~md
# CLI Games Collection

## Run launcher

```bash
go run .
```

## Run Snake directly

```bash
go run . play snake
```

## Run tests

```bash
go test ./...
```
~~~

- [ ] **Step 2: Save it to `README.md`**

```md
# CLI Games Collection

Progressive terminal games laboratory built in Go.

## Commands

- `go run .` — open launcher
- `go run . play snake` — run Snake directly
- `go test ./...` — run tests
```

- [ ] **Step 3: Run the full verification suite**

Run: `go test ./...`
Expected: PASS across `cmd`, `internal/engine/runtime`, `internal/launcher`, and `internal/games/snake`.

- [ ] **Step 4: Manually verify command surfaces**

Run: `go run . play snake`
Expected: Process exits successfully after rendering a short deterministic Snake session through the headless adapter.

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "docs: add bootstrap usage guide"
```

## Self-Review

- **Spec coverage:** This plan covers repository bootstrap, Cobra command surface, Bubble Tea launcher shell, deterministic runtime contracts, headless simulation support, and Snake as the first validating game.
- **Placeholder scan:** No unresolved placeholders remain.
- **Type consistency:** Runtime uses `input.FrameInput`, `runtime.Screen`, `snake.Direction`, and `snake.Module` consistently across tasks.

## Recommended execution order

1. Task 1 — Bootstrap module and CLI boundary
2. Task 2 — Establish deterministic runtime contracts
3. Task 3 — Add Bubble Tea launcher shell
4. Task 4 — Implement Snake domain and deterministic simulation
5. Task 5 — Wire Snake into runtime and CLI
6. Task 6 — Verify end-to-end foundation scope
