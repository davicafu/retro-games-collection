# Sokoban Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stabilize Sokoban by making level outcomes explicit, tightening loop/timing behavior, hardening overlay flows, and protecting the corrected behavior with deterministic tests.

**Architecture:** Keep Sokoban on its current local runtime path, but replace ambiguous level-flow conventions with explicit result semantics. Harden the existing loop and overlay usage only where it serves the critical gameplay flows, then lock those contracts with focused tests.

**Tech Stack:** Go, Sokoban module in `internal/games/sokoban`, shared overlay package in `internal/engine/ui/overlay`, Go testing package.

---

## File map

| File | Responsibility |
|------|----------------|
| `internal/games/sokoban/module.go` | Level loop, level result semantics, overlay flow integration, top-level progression decisions |
| `internal/games/sokoban/module_test.go` | Contract tests for quit/progression/completion, pause/help/summary flows, and cancellation behavior |
| `internal/games/sokoban/game.go` | Only touch if a narrowly scoped game-state helper is needed by the level-result flow |

---

### Task 1: Replace ambiguous level exits with an explicit result contract

**Files:**
- Modify: `internal/games/sokoban/module.go`
- Modify: `internal/games/sokoban/module_test.go`

- [ ] **Step 1: Write the failing quit/completion contract tests**

Add to `internal/games/sokoban/module_test.go`:

```go
func TestRunLevelQuitReturnsQuitResult(t *testing.T) {
	module := Module{
		Screen:   render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionQuit}}}},
	}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	result, _, err := module.runLevel(context.Background(), game, 0, 0)
	if err != nil {
		t.Fatalf("expected quit path without runtime error, got %v", err)
	}
	if result.Kind != levelResultQuitToMenu {
		t.Fatalf("expected quit-to-menu result, got %v", result.Kind)
	}
}

func TestRunLevelWinReturnsCompletedResult(t *testing.T) {
	module := Module{Screen: render.NewBufferScreen(60, 24), Provider: &scriptedProvider{}}
	game, err := ParseLevelString("#####\n#@$.#\n#####")
	if err != nil {
		t.Fatalf("expected custom level parse to succeed, got %v", err)
	}

	result, elapsed, err := module.runLevel(context.Background(), game, 0, 0)
	if err != nil {
		t.Fatalf("expected completed path without runtime error, got %v", err)
	}
	if result.Kind != levelResultCompleted {
		t.Fatalf("expected completed result, got %v", result.Kind)
	}
	if elapsed < 0 {
		t.Fatalf("expected non-negative elapsed time, got %v", elapsed)
	}
}
```

- [ ] **Step 2: Run the contract tests to verify they fail**

Run: `go test ./internal/games/sokoban -run "TestRunLevelQuitReturnsQuitResult|TestRunLevelWinReturnsCompletedResult"`

Expected: FAIL because `runLevel` does not yet return an explicit level result type.

- [ ] **Step 3: Implement a minimal explicit level result contract**

In `internal/games/sokoban/module.go`, introduce an explicit result type and update `runLevel` + `Run()` to use it:

```go
type levelResultKind string

const (
	levelResultCompleted  levelResultKind = "completed"
	levelResultRestart    levelResultKind = "restart"
	levelResultNextLevel  levelResultKind = "next-level"
	levelResultQuitToMenu levelResultKind = "quit-to-menu"
)

type levelResult struct {
	Kind levelResultKind
}

func (m Module) runLevel(ctx context.Context, game *Game, levelIdx int, initialPenalty time.Duration) (levelResult, time.Duration, error) {
	// existing loop logic here
	// return levelResult{Kind: levelResultQuitToMenu}, 0, nil on quit
	// return levelResult{Kind: levelResultCompleted}, elapsed, nil on win
}
```

Then update `Run()` to decide progression by `result.Kind`, not by `nil/error` inference.

- [ ] **Step 4: Run the contract tests to verify they pass**

Run: `go test ./internal/games/sokoban -run "TestRunLevelQuitReturnsQuitResult|TestRunLevelWinReturnsCompletedResult"`

Expected: PASS

- [ ] **Step 5: Commit the result-contract fix**

```bash
git add internal/games/sokoban/module.go internal/games/sokoban/module_test.go
git commit -m "fix(sokoban): make level outcomes explicit"
```

---

### Task 2: Tighten progression, restart, and next-level timing semantics

**Files:**
- Modify: `internal/games/sokoban/module.go`
- Modify: `internal/games/sokoban/module_test.go`

- [ ] **Step 1: Write the failing progression/timing tests**

Add to `internal/games/sokoban/module_test.go`:

```go
func TestRunLevelPauseRestartDoesNotReportCompletion(t *testing.T) {
	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionPause}},
			{Actions: []input.Action{input.ActionDown, input.ActionSelect}},
		}},
	}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	result, _, err := module.runLevel(context.Background(), game, 0, 0)
	if err != nil {
		t.Fatalf("expected restart path without runtime error, got %v", err)
	}
	if result.Kind != levelResultRestart {
		t.Fatalf("expected restart result, got %v", result.Kind)
	}
}

func TestRunLevelPauseNextReturnsNextLevelResult(t *testing.T) {
	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionPause}},
			{Actions: []input.Action{input.ActionDown, input.ActionDown, input.ActionSelect}},
		}},
	}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	result, _, err := module.runLevel(context.Background(), game, 0, 0)
	if err != nil {
		t.Fatalf("expected next-level path without runtime error, got %v", err)
	}
	if result.Kind != levelResultNextLevel {
		t.Fatalf("expected next-level result, got %v", result.Kind)
	}
}
```

- [ ] **Step 2: Run the progression tests to verify they fail**

Run: `go test ./internal/games/sokoban -run "TestRunLevelPauseRestartDoesNotReportCompletion|TestRunLevelPauseNextReturnsNextLevelResult"`

Expected: FAIL until restart/next-level have explicit result semantics.

- [ ] **Step 3: Implement minimal restart/next-level flow separation**

In `internal/games/sokoban/module.go`, update the pause flow inside `runLevel`:

```go
switch result {
case pauseContinue:
	// redraw only
case pauseRestart:
	return levelResult{Kind: levelResultRestart}, 0, nil
case pauseNext:
	return levelResult{Kind: levelResultNextLevel}, time.Since(start) + initialPenalty, nil
case pauseQuit:
	return levelResult{Kind: levelResultQuitToMenu}, 0, nil
}
```

Then make `Run()` decide whether to:

- replay the same level on restart
- advance without completion overlay on next-level
- record time only for true completion or intentional skip, depending on the chosen rule

Document the chosen timing rule in code comments near the progression switch.

- [ ] **Step 4: Run the targeted Sokoban tests**

Run: `go test ./internal/games/sokoban -run "TestRunLevelPauseRestartDoesNotReportCompletion|TestRunLevelPauseNextReturnsNextLevelResult|TestRunLevelQuitReturnsQuitResult|TestRunLevelWinReturnsCompletedResult"`

Expected: PASS

- [ ] **Step 5: Commit the progression/timing separation**

```bash
git add internal/games/sokoban/module.go internal/games/sokoban/module_test.go
git commit -m "fix(sokoban): separate restart and next-level progression"
```

---

### Task 3: Harden pause/help/final-summary overlay runtime behavior

**Files:**
- Modify: `internal/games/sokoban/module.go`
- Modify: `internal/games/sokoban/module_test.go`

- [ ] **Step 1: Write the failing overlay flow tests**

Add to `internal/games/sokoban/module_test.go`:

```go
func TestShowPauseMenuReturnsContinueOnPauseClose(t *testing.T) {
	module := Module{
		Screen:   render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionPause}}}},
	}

	got, err := module.showPauseMenu(context.Background(), 0)
	if err != nil {
		t.Fatalf("expected pause close without error, got %v", err)
	}
	if got != pauseContinue {
		t.Fatalf("expected pauseContinue, got %v", got)
	}
}

func TestShowHelpOverlayReturnsContextErrorWhenCanceled(t *testing.T) {
	module := Module{Screen: render.NewBufferScreen(60, 24), Provider: &scriptedProvider{}}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := module.showHelpOverlay(ctx, game, 0, 0); err == nil {
		t.Fatal("expected context cancellation error")
	}
}
```

- [ ] **Step 2: Run the overlay flow tests to verify they fail if behavior regresses**

Run: `go test ./internal/games/sokoban -run "TestShowPauseMenuReturnsContinueOnPauseClose|TestShowHelpOverlayReturnsContextErrorWhenCanceled"`

Expected: PASS if already green; if not, treat as RED and fix before moving on.

- [ ] **Step 3: Normalize overlay runtime behavior in the module**

In `internal/games/sokoban/module.go`, keep or introduce these invariants:

- `showPauseMenu(...)` propagates `Present()` errors
- `showPauseMenu(...)` respects `ctx.Done()`
- `showHelpOverlay(...)` respects `ctx.Done()` and redraws gameplay cleanly on dismiss
- `showFinalSummary(...)` exits predictably on select/quit/pause and does not accidentally advance gameplay

If any overlay helper is growing too large, split one narrow helper inside `module.go` only if it reduces ambiguity.

- [ ] **Step 4: Run the overlay-focused test set**

Run: `go test ./internal/games/sokoban -run "TestShowPauseMenuReturnsContinueOnPauseClose|TestShowHelpOverlayDismissesAndRedrawsLevel|TestShowHelpOverlayReturnsContextErrorWhenCanceled|TestShowFinalSummaryExitsOnSelectQuitAndPause"`

Expected: PASS

- [ ] **Step 5: Commit overlay hardening**

```bash
git add internal/games/sokoban/module.go internal/games/sokoban/module_test.go
git commit -m "fix(sokoban): harden overlay flow behavior"
```

---

### Task 4: Lock the corrected Sokoban flows with a deterministic safety net

**Files:**
- Modify: `internal/games/sokoban/module_test.go`

- [ ] **Step 1: Add a regression-focused integration test set**

Add to `internal/games/sokoban/module_test.go`:

```go
func TestRunReturnsNilOnQuitToMenuWithoutLevelComplete(t *testing.T) {
	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionQuit}},
		}},
	}

	if err := module.Run(context.Background()); err != nil {
		t.Fatalf("expected Run to stop cleanly on quit-to-menu, got %v", err)
	}
}

func TestRunPauseQuitStopsWithoutProgression(t *testing.T) {
	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionPause}},
			{Actions: []input.Action{input.ActionQuit}},
		}},
	}

	if err := module.Run(context.Background()); err != nil {
		t.Fatalf("expected Run to stop cleanly on pause quit, got %v", err)
	}
}
```

- [ ] **Step 2: Run the new regression tests to verify their status**

Run: `go test ./internal/games/sokoban -run "TestRunReturnsNilOnQuitToMenuWithoutLevelComplete|TestRunPauseQuitStopsWithoutProgression"`

Expected: PASS if flows are already corrected; otherwise FAIL and fix before continuing.

- [ ] **Step 3: Strengthen any remaining deterministic seams**

If these tests reveal remaining ambiguity, make the minimal code change required in `module.go` to keep the flow deterministic. Do not broaden scope into new features or global runtime redesign.

- [ ] **Step 4: Run the full Sokoban package tests**

Run: `go test ./internal/games/sokoban`

Expected: PASS

- [ ] **Step 5: Commit the Sokoban safety net**

```bash
git add internal/games/sokoban/module_test.go internal/games/sokoban/module.go
git commit -m "test(sokoban): lock hardening flows with regression coverage"
```

---

### Task 5: Final verification for the hardening pass

**Files:**
- Modify: `internal/games/sokoban/module.go` only if verification reveals a real issue
- Modify: `internal/games/sokoban/module_test.go` only if verification reveals a missing regression test

- [ ] **Step 1: Run targeted verification**

Run: `go test ./internal/games/sokoban ./internal/engine/ui/overlay`

Expected: PASS

- [ ] **Step 2: Run full repository verification**

Run: `go test ./...`

Expected: PASS

- [ ] **Step 3: If verification reveals a real issue, fix it minimally**

Only if a failing test appears:

```go
// Make the smallest code or test change required to restore the intended contract.
```

- [ ] **Step 4: Re-run verification after any fix**

Run: `go test ./internal/games/sokoban ./internal/engine/ui/overlay && go test ./...`

Expected: PASS

- [ ] **Step 5: Commit the final hardening pass**

```bash
git add internal/games/sokoban/module.go internal/games/sokoban/module_test.go
git commit -m "fix(sokoban): harden level flow and overlay behavior"
```

---

## Spec coverage check

- quit / progression / completion: covered by Tasks 1-2
- loop/timing stabilization: covered by Task 2
- pause / help / overlays: covered by Task 3
- tests and determinism: covered by Task 4 and Task 5

## Final checklist

- [ ] Sokoban has an explicit level-flow contract.
- [ ] Quit is not treated as completion.
- [ ] Restart and next-level semantics are explicit.
- [ ] Overlay runtime behavior is hardened and context-aware.
- [ ] Sokoban regression tests protect the corrected flows.
- [ ] `go test ./...` passes.

## Next step

Execute the plan in a fresh Sokoban-hardening worktree, keeping each task tied to one flow boundary at a time.
