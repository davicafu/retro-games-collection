# Snake Continuous Food Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Snake continuously playable by spawning new food after each pickup, while keeping runtime behavior real and tests deterministic.

**Architecture:** Keep food spawning as a domain-level responsibility with an injectable selection strategy. Real gameplay uses a pseudo-random selector chosen by the app/module path, while tests use deterministic selectors to prove repeated pickups, valid spawning, and full-board termination.

**Tech Stack:** Go, existing Snake domain/runtime, injectable spawn strategy, deterministic tests

---

## Planned File Structure

- `internal/games/snake/game.go` — Snake domain state, update rules, spawn trigger, full-board handling
- `internal/games/snake/game_test.go` — deterministic rule tests for repeated pickups and terminal conditions
- `internal/games/snake/spawn.go` — spawn strategy contracts and default implementations
- `internal/games/snake/spawn_test.go` — selector-focused tests for valid candidate choice
- `internal/games/snake/module.go` — module wiring for runtime path using pseudo-random strategy

## Task 1: Add injectable food selection strategy

**Files:**
- Create: `internal/games/snake/spawn.go`
- Test: `internal/games/snake/spawn_test.go`

- [ ] **Step 1: Write the failing selector test**

```go
// internal/games/snake/spawn_test.go
package snake

import "testing"

func TestSequenceSelector_SelectsNextCandidateDeterministically(t *testing.T) {
	selector := NewSequenceSelector([]int{1})
	candidates := []Position{{X: 0, Y: 0}, {X: 2, Y: 2}, {X: 4, Y: 4}}

	got, ok := selector.Select(candidates)
	if !ok {
		t.Fatal("expected selector to choose a candidate")
	}
	if got != (Position{X: 2, Y: 2}) {
		t.Fatalf("expected second candidate, got %+v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestSequenceSelector_SelectsNextCandidateDeterministically -v`
Expected: FAIL with undefined `NewSequenceSelector`.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/games/snake/spawn.go
package snake

import "math/rand"

type FoodSelector interface {
	Select(candidates []Position) (Position, bool)
}

type SequenceSelector struct {
	indices []int
	idx     int
}

func NewSequenceSelector(indices []int) *SequenceSelector {
	return &SequenceSelector{indices: append([]int(nil), indices...)}
}

func (s *SequenceSelector) Select(candidates []Position) (Position, bool) {
	if len(candidates) == 0 {
		return Position{}, false
	}
	if len(s.indices) == 0 {
		return candidates[0], true
	}
	choice := s.indices[s.idx%len(s.indices)]
	s.idx++
	return candidates[choice%len(candidates)], true
}

type RandomSelector struct{ rng *rand.Rand }

func NewRandomSelector(rng *rand.Rand) *RandomSelector { return &RandomSelector{rng: rng} }

func (s *RandomSelector) Select(candidates []Position) (Position, bool) {
	if len(candidates) == 0 {
		return Position{}, false
	}
	return candidates[s.rng.Intn(len(candidates))], true
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/games/snake -run TestSequenceSelector_SelectsNextCandidateDeterministically -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/spawn.go internal/games/snake/spawn_test.go
git commit -m "feat: add snake food selection strategies"
```

## Task 2: Make Snake respawn food after eating

**Files:**
- Modify: `internal/games/snake/game.go`
- Modify: `internal/games/snake/game_test.go`

- [ ] **Step 1: Write the failing repeated-pickup test**

```go
// internal/games/snake/game_test.go
func TestGame_UpdateRespawnsFoodAfterPickup(t *testing.T) {
	g := NewGame(Config{
		Width:    6,
		Height:   6,
		Start:    Position{X: 2, Y: 2},
		Food:     []Position{{X: 3, Y: 2}},
		Selector: NewSequenceSelector([]int{0}),
	})

	if err := g.Update(DirectionRight); err != nil {
		t.Fatalf("expected pickup update to succeed, got %v", err)
	}

	if g.Score() != 1 {
		t.Fatalf("expected score 1, got %d", g.Score())
	}

	food := g.Food()
	if len(food) != 1 {
		t.Fatalf("expected one respawned food item, got %d", len(food))
	}
	if food[0] == (Position{X: 3, Y: 2}) {
		t.Fatal("expected new food to move to a new valid cell")
	}
	for _, segment := range g.Body() {
		if food[0] == segment {
			t.Fatalf("expected food not to overlap snake body, got %+v", food[0])
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestGame_UpdateRespawnsFoodAfterPickup -v`
Expected: FAIL because `Config` lacks `Selector` or food does not respawn.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/games/snake/game.go
package snake

import (
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

type Position = grid.Point

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
	Width    int
	Height   int
	Start    Position
	Food     []Position
	Selector FoodSelector
}

type Game struct {
	width     int
	height    int
	body      []Position
	direction Direction
	food      []Position
	selector  FoodSelector
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
		selector:  cfg.Selector,
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
		g.respawnFood()
		return nil
	}
	g.body = g.body[:len(g.body)-1]
	return nil
}

func (g *Game) respawnFood() {
	if g.selector == nil {
		return
	}
	candidates := g.availableFoodCells()
	next, ok := g.selector.Select(candidates)
	if !ok {
		g.dead = true
		return
	}
	g.food = []Position{next}
}

func (g *Game) availableFoodCells() []Position {
	occupied := map[Position]struct{}{}
	for _, segment := range g.body {
		occupied[segment] = struct{}{}
	}
	for _, food := range g.food {
		occupied[food] = struct{}{}
	}
	var cells []Position
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			candidate := Position{X: x, Y: y}
			if _, blocked := occupied[candidate]; blocked {
				continue
			}
			cells = append(cells, candidate)
		}
	}
	return cells
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/games/snake -run TestGame_UpdateRespawnsFoodAfterPickup -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/game.go internal/games/snake/game_test.go
git commit -m "feat: respawn snake food after pickups"
```

## Task 3: Handle full-board termination deterministically

**Files:**
- Modify: `internal/games/snake/game_test.go`

- [ ] **Step 1: Write the failing full-board test**

```go
func TestGame_UpdateEndsWhenNoFoodCellRemains(t *testing.T) {
	g := NewGame(Config{
		Width:    2,
		Height:   1,
		Start:    Position{X: 0, Y: 0},
		Food:     []Position{{X: 1, Y: 0}},
		Selector: NewSequenceSelector([]int{0}),
	})

	if err := g.Update(DirectionRight); err != nil {
		t.Fatalf("expected final pickup update to succeed, got %v", err)
	}

	if !g.Dead() {
		t.Fatal("expected game to end when no valid food cell remains")
	}
	if got := g.Food(); len(got) != 0 {
		t.Fatalf("expected no food remaining, got %+v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestGame_UpdateEndsWhenNoFoodCellRemains -v`
Expected: FAIL because full-board termination is not yet explicit.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/games/snake/game.go (adjust respawnFood)
func (g *Game) respawnFood() {
	if g.selector == nil {
		return
	}
	candidates := g.availableFoodCells()
	if len(candidates) == 0 {
		g.dead = true
		g.food = nil
		return
	}
	next, ok := g.selector.Select(candidates)
	if !ok {
		g.dead = true
		g.food = nil
		return
	}
	g.food = []Position{next}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/games/snake -run TestGame_UpdateEndsWhenNoFoodCellRemains -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/game.go internal/games/snake/game_test.go
git commit -m "fix: end snake when board has no food space"
```

## Task 4: Use pseudo-random selector in runtime path

**Files:**
- Modify: `internal/games/snake/module.go`
- Test: `internal/games/snake/module_test.go`

- [ ] **Step 1: Write the failing runtime continuity test**

```go
func TestModule_RunContinuesAfterFirstFoodPickup(t *testing.T) {
	manual := clock.NewManual()
	screen := tcellscreen.NewSimulationScreenForTest(12, 12)
	provider := &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionRight}}, {}, {Actions: []input.Action{input.ActionQuit}}}}
	module := NewModule(manual, screen, provider)

	errCh := make(chan error, 1)
	go func() { errCh <- module.Run(context.Background()) }()
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)

	if err := <-errCh; err != nil {
		t.Fatalf("expected graceful exit after continued play, got %v", err)
	}

	snapshot := screen.Snapshot()
	if snapshot[1][4] != '█' {
		t.Fatalf("expected snake to continue past first pickup, got %q", snapshot[1][4])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestModule_RunContinuesAfterFirstFoodPickup -v`
Expected: FAIL because module still ends when food list is empty.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/games/snake/module.go
package snake

import (
	"context"
	"math/rand"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/runtime"
)

// keep existing runtimeGame methods, but update ShouldQuit
func (g *runtimeGame) ShouldQuit() bool { return g.Dead() }

func NewModule(clk clock.Clock, screen render.Screen, provider input.Provider) Module {
	return Module{clock: clk, screen: screen, provider: provider}
}

func (m Module) Run(ctx context.Context) error {
	rt := runtime.New(runtime.Config{Clock: m.clock, Input: m.provider, Screen: m.screen})
	selector := NewRandomSelector(rand.New(rand.NewSource(time.Now().UnixNano())))
	game := &runtimeGame{Game: NewGame(Config{
		Width:    12,
		Height:   12,
		Start:    Position{X: 1, Y: 1},
		Food:     []Position{{X: 4, Y: 1}},
		Selector: selector,
	})}
	return rt.Run(ctx, game)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/games/snake -run TestModule_RunContinuesAfterFirstFoodPickup -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/module.go internal/games/snake/module_test.go
git commit -m "feat: keep snake running after food pickups"
```

## Self-Review

- **Spec coverage:** This plan covers continuous respawn, non-overlapping food, deterministic selectors for tests, runtime pseudo-random selection, and explicit full-board termination.
- **Placeholder scan:** No unresolved placeholders remain.
- **Type consistency:** `FoodSelector`, `SequenceSelector`, `RandomSelector`, and `Config.Selector` are used consistently across domain and runtime tasks.

## Recommended execution order

1. Task 1 — Add injectable food selection strategy
2. Task 2 — Make Snake respawn food after eating
3. Task 3 — Handle full-board termination deterministically
4. Task 4 — Use pseudo-random selector in runtime path
