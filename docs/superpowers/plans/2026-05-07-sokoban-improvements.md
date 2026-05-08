# Sokoban Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add final summary screen, harder 10-level progression, and single undo per level to Sokoban

**Architecture:** Three independent changes: (1) collect elapsed times per level and display summary after level 10, (2) replace current levels with harder versions, (3) add undoUsed flag to game struct and reset per level.

**Tech Stack:** Go, existing Sokoban game engine

---

## Task 1: Single Undo Per Level

**Files:**
- Modify: `internal/games/sokoban/game.go`
- Modify: `internal/games/sokoban/module.go`

- [ ] **Step 1: Add undoUsed field to Game struct**

In `internal/games/sokoban/game.go`, add `undoUsed bool` field to the `Game` struct around line 8-16:

```go
type Game struct {
	Width     int
	Height    int
	Player    grid.Point
	Walls     []grid.Point
	Boxes     []grid.Point
	Targets   []grid.Point
	undoStack *undoStack
	undoUsed  bool  // NEW: tracks if undo was used this level
}
```

- [ ] **Step 2: Add ResetUndo method**

Add this method to `game.go` after the `Undo()` method (around line 94):

```go
func (g *Game) ResetUndo() {
	g.undoUsed = false
}
```

- [ ] **Step 3: Modify Undo method to check undoUsed**

Replace the current `Undo()` method (lines 86-94) with:

```go
func (g *Game) Undo() {
	if g.undoUsed {
		return // Already used undo this level
	}
	g.initUndoStack()
	entry := g.undoStack.Pop()
	if entry == nil {
		return
	}
	g.Player = entry.player
	g.Boxes = entry.boxes
	g.undoUsed = true
}
```

- [ ] **Step 4: Reset undo at start of each level**

In `internal/games/sokoban/module.go`, in the `Run()` method around line 29-44, add `game.ResetUndo()` after creating the game:

```go
for levelIdx := 0; levelIdx < len(Levels); levelIdx++ {
	game, err := ParseLevelString(Levels[levelIdx])
	if err != nil {
		return fmt.Errorf("failed to load level %d: %w", levelIdx, err)
	}

	game.ResetUndo()  // NEW: reset undo for new level

	elapsed, err := m.runLevel(ctx, game, levelIdx)
	// ... rest of code
}
```

- [ ] **Step 5: Test the change**

Run: `go test ./internal/games/sokoban/...`
Expected: All tests pass

- [ ] **Step 6: Commit**

```bash
git add internal/games/sokoban/game.go internal/games/sokoban/module.go
git commit -m "feat(sokoban): add single undo per level"
```

---

## Task 2: Track and Display Per-Level Times

**Files:**
- Modify: `internal/games/sokoban/module.go`

- [ ] **Step 1: Add times slice to store elapsed durations**

In `module.go`, in the `Run()` method around line 28-47, declare a slice to store times before the loop:

```go
func (m Module) Run(ctx context.Context) error {
	times := make([]time.Duration, 0, len(Levels))  // NEW

	for levelIdx := 0; levelIdx < len(Levels); levelIdx++ {
		// ... existing code ...

		elapsed, err := m.runLevel(ctx, game, levelIdx)
		if err != nil {
			return err
		}

		times = append(times, elapsed)  // NEW: store elapsed time

		if err := m.showLevelComplete(levelIdx, elapsed); err != nil {
			return err
		}
	}

	return m.showFinalSummary(times)  // NEW: show final summary instead of nil
}
```

- [ ] **Step 2: Add showFinalSummary method**

Add this new method after `showLevelComplete()` in `module.go` (around line 139):

```go
func (m Module) showFinalSummary(times []time.Duration) error {
	m.Screen.Clear()

	var b strings.Builder
	b.WriteString("╔══════════════════════════════════════╗\n")
	b.WriteString("║      🎉 FIN DEL JUEGO! 🎉             ║\n")
	b.WriteString("╠══════════════════════════════════════╣\n")

	for i, t := range times {
		line := fmt.Sprintf("║  Nivel %2d:    %s               ║", i+1, fmtDuration(t))
		b.WriteString(line + "\n")
	}

	b.WriteString("╠══════════════════════════════════════╣\n")

	var total time.Duration
	for _, t := range times {
		total += t
	}
	totalLine := fmt.Sprintf("║  TOTAL:      %s               ║", fmtDuration(total))
	b.WriteString(totalLine + "\n")

	b.WriteString("╚══════════════════════════════════════╝")

	// Center the box on screen
	sw, sh := m.Screen.Size()
	lines := strings.Split(b.String(), "\n")
	boxH := len(lines)
	boxW := 42
	boxY := (sh - boxH) / 2
	boxX := (sw - boxW) / 2

	for i, line := range lines {
		render.WriteText(m.Screen, boxX, boxY+i, line)
	}

	if err := m.Screen.Present(); err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	return nil
}
```

Note: This method requires adding `"strings"` to the imports if not already present.

- [ ] **Step 3: Add strings import if needed**

Check the imports in `module.go` and add `"strings"` if missing:

```go
import (
	"context"
	"fmt"
	"strings"  // ADD THIS if not present
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)
```

- [ ] **Step 4: Test the change**

Run: `go build ./...`
Expected: Builds successfully

- [ ] **Step 5: Commit**

```bash
git add internal/games/sokoban/module.go
git commit -m "feat(sokoban): add per-level time tracking and final summary"
```

---

## Task 3: Create 10 Harder Levels

**Files:**
- Modify: `internal/games/sokoban/levels.go`

- [ ] **Step 1: Replace current levels with harder versions**

Replace the `Levels` variable in `levels.go` (lines 12-97) with these 10 harder levels:

```go
var Levels = []string{
	// Level 1: Simple push, but tighter space (1 box, 1 target)
	`#####
#@$.#
#####`,

	// Level 2: Two boxes, corner trick (2 boxes, 2 targets)
	`######
#  @ #
# $  #
#  . #
# $ .#
######`,

	// Level 3: Three boxes, need planning (3 boxes, 3 targets)
	`#######
#  @  #
# $.$.#
#  $  #
#  .  #
#######`,

	// Level 4: Narrow corridor challenge (3 boxes, 3 targets)
	`########
#      #
# .$@$ #
#      #
# . $ .#
########`,

	// Level 5: Four boxes, requires strategy (4 boxes, 4 targets)
	`#########
#   @    #
#  $ $   #
# . $ .  #
# .   .  #
#########`,

	// Level 6: Dead-end traps (3 boxes, 3 targets)
	`########
#   @  #
# $    #
# . $ .#
#    $ #
#   .  #
########`,

	// Level 7: Multiple solution paths, easy to deadlock (4 boxes, 4 targets)
	`##########
# @     #
# $ $   #
#. $.$. #
#  $ $  #
# .   . #
##########`,

	// Level 8: Tight squeeze (4 boxes, 4 targets)
	`##########
#   @    #
#  $ $   #
# . $ .  #
#  $ $   #
# .   .  #
##########`,

	// Level 9: The gauntlet (4 boxes, 4 targets)
	`###########
#  @     #
#  $     #
#  . $ . #
#    $   #
# . $. . #
#    .   #
###########`,

	// Level 10: Final challenge (5 boxes, 5 targets)
	`############
#   @      #
#   $      #
#  . $ .   #
#   $ $    #
#  . $ .   #
#   .      #
############`,
}
```

- [ ] **Step 2: Test the change**

Run: `go test ./internal/games/sokoban/...`
Expected: All tests pass (test `ParseLevelString` should work with new levels)

- [ ] **Step 3: Play test (manual verification)**

Run the game and verify:
- All 10 levels are solvable
- Each level is harder than the previous
- Timer works correctly
- Undo works once per level
- Final summary shows all times after level 10

- [ ] **Step 4: Commit**

```bash
git add internal/games/sokoban/levels.go
git commit -m "feat(sokoban): replace 10 levels with harder progression"
```

---

## Completion Checklist

- [ ] Task 1: Single undo per level works
- [ ] Task 2: Per-level times tracked and final summary displays
- [ ] Task 3: 10 harder levels implemented
- [ ] All tests pass
- [ ] Manual playtest confirms all features work