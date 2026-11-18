# Snake Playfield Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Correct Snake’s layout model so the logical board remains pure while score and borders render outside the playable area.

**Architecture:** Introduce an explicit playfield layout/projection model in the renderer instead of using scattered magic offsets. The Snake domain remains unchanged; rendering owns the board origin, board dimensions, chrome rows, and logical-to-screen projection.

**Tech Stack:** Go, existing Snake runtime, simulation screen tests, explicit render projection helpers

---

## Planned File Structure

- `internal/games/snake/render.go` — layout struct, projection helper, border + score drawing, board rendering
- `internal/games/snake/render_test.go` — focused layout/projection tests
- `internal/games/snake/module_test.go` — integration assertions updated to real rendered coordinates only where needed

## Task 1: Add explicit playfield layout projection

**Files:**
- Create: `internal/games/snake/render_test.go`
- Modify: `internal/games/snake/render.go`

- [ ] **Step 1: Write the failing projection test**

```go
// internal/games/snake/render_test.go
package snake

import "testing"

func TestPlayfieldLayout_ProjectsLogicalCellsIntoInterior(t *testing.T) {
	layout := defaultPlayfieldLayout(12, 12)

	topLeft := layout.toScreen(Position{X: 0, Y: 0})
	if topLeft != (Position{X: 1, Y: 2}) {
		t.Fatalf("expected logical (0,0) -> screen (1,2), got %+v", topLeft)
	}

	bottomRight := layout.toScreen(Position{X: 11, Y: 11})
	if bottomRight != (Position{X: 12, Y: 13}) {
		t.Fatalf("expected logical (11,11) -> screen (12,13), got %+v", bottomRight)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestPlayfieldLayout_ProjectsLogicalCellsIntoInterior -v`
Expected: FAIL with undefined `defaultPlayfieldLayout` or `toScreen`.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/games/snake/render.go
package snake

import (
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

type playfieldLayout struct {
	boardWidth   int
	boardHeight  int
	boardOriginX int
	boardOriginY int
}

func defaultPlayfieldLayout(boardWidth, boardHeight int) playfieldLayout {
	return playfieldLayout{
		boardWidth:   boardWidth,
		boardHeight:  boardHeight,
		boardOriginX: 1,
		boardOriginY: 2,
	}
}

func (l playfieldLayout) toScreen(p Position) Position {
	return Position{X: l.boardOriginX + p.X, Y: l.boardOriginY + p.Y}
}

func (l playfieldLayout) totalWidth() int  { return l.boardWidth + 2 }
func (l playfieldLayout) totalHeight() int { return l.boardHeight + 3 }

func Draw(screen render.Screen, body []Position, food []Position, score int) {
	layout := defaultPlayfieldLayout(12, 12)
	drawBorder(screen, layout)
	drawText(screen, 2, 0, fmt.Sprintf("Score: %d", score))
	for _, segment := range body {
		point := layout.toScreen(segment)
		screen.SetCell(point.X, point.Y, render.Cell{Rune: '█'})
	}
	for _, item := range food {
		point := layout.toScreen(item)
		screen.SetCell(point.X, point.Y, render.Cell{Rune: '•'})
	}
}

func drawBorder(screen render.Screen, layout playfieldLayout) {
	width, height := layout.totalWidth(), layout.totalHeight()
	for x := 0; x < width; x++ {
		top := '-'
		bottom := '-'
		if x == 0 || x == width-1 {
			top = '+'
			bottom = '+'
		}
		screen.SetCell(x, 1, render.Cell{Rune: top})
		screen.SetCell(x, height-1, render.Cell{Rune: bottom})
	}
	for y := 2; y < height-1; y++ {
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

Run: `go test ./internal/games/snake -run TestPlayfieldLayout_ProjectsLogicalCellsIntoInterior -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/render.go internal/games/snake/render_test.go
git commit -m "feat: add explicit snake playfield layout projection"
```

## Task 2: Prove chrome does not invade board edges

**Files:**
- Modify: `internal/games/snake/render_test.go`

- [ ] **Step 1: Write the failing edge-cell render test**

```go
func TestDraw_RendersBottomRightCellInsidePlayfield(t *testing.T) {
	screen := tcellscreen.NewSimulationScreenForTest(14, 15)
	Draw(screen, []Position{{X: 11, Y: 11}}, nil, 7)

	snapshot := screen.Snapshot()
	if snapshot[13][12] != '█' {
		t.Fatalf("expected bottom-right logical cell inside playfield at (12,13), got %q", snapshot[13][12])
	}
	if snapshot[13][13] != '|' {
		t.Fatalf("expected right border outside playfield, got %q", snapshot[13][13])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestDraw_RendersBottomRightCellInsidePlayfield -v`
Expected: FAIL if projection/layout still overlaps chrome.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/games/snake/render_test.go (append import)
import "github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/tcellscreen"
```

If the test fails after Task 1 implementation, adjust only the layout constants and border loops in `render.go` until the bottom-right logical cell renders inside the board and the border remains outside it.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/games/snake -run TestDraw_RendersBottomRightCellInsidePlayfield -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/render.go internal/games/snake/render_test.go
git commit -m "fix: keep snake board cells inside chrome bounds"
```

## Task 3: Update integration assertions to real projected coordinates

**Files:**
- Modify: `internal/games/snake/module_test.go`

- [ ] **Step 1: Write the failing runtime render assertion update**

```go
func TestModule_RunRendersAndQuitsGracefully(t *testing.T) {
	manual := clock.NewManual()
	screen := tcellscreen.NewSimulationScreenForTest(14, 15)
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

	snapshot := screen.Snapshot()
	if got := snapshot[4][2]; got != '█' {
		t.Fatalf("expected snake head rendered at projected playfield cell (2,4), got %q", got)
	}
	if got := snapshot[2][5]; got != '•' {
		t.Fatalf("expected food rendered at projected playfield cell (5,2), got %q", got)
	}
	if got := snapshot[1][0]; got != '+' {
		t.Fatalf("expected top-left border corner, got %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/games/snake -run TestModule_RunRendersAndQuitsGracefully -v`
Expected: FAIL until the simulation screen size and coordinates match the new layout.

- [ ] **Step 3: Write minimal implementation**

Update only the test expectations and screen dimensions in `module_test.go` so they match the explicit layout model introduced in Tasks 1–2. Do not change domain logic.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/games/snake -run TestModule_RunRendersAndQuitsGracefully -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/games/snake/module_test.go
git commit -m "test: align snake runtime render checks with playfield layout"
```

## Self-Review

- **Spec coverage:** This revised plan covers the real problem: explicit logical-to-render projection, score/border outside the board, and tests that prove edge cells stay inside the playable region.
- **Placeholder scan:** No unresolved placeholders remain.
- **Type consistency:** `playfieldLayout`, `defaultPlayfieldLayout`, and `toScreen` are consistent across renderer and tests.

## Recommended execution order

1. Task 1 — Add explicit playfield layout projection
2. Task 2 — Prove chrome does not invade board edges
3. Task 3 — Update integration assertions to real projected coordinates
