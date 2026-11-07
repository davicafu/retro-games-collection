# Sokoban Run Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn Sokoban into a strict session-only run mode with linear progression, a global pool of 4 undos, and run-aware overlays and summaries.

**Architecture:** Add a run/session layer above Sokoban’s per-level `Game` so cross-level concerns live in one place. Keep level rules inside `Game`, keep UI rendering in the shared overlay package, and let the Sokoban module own run progression, undo economy, and final outcomes.

**Tech Stack:** Go, `internal/games/sokoban`, shared overlay package in `internal/engine/ui/overlay`, Go testing package.

---

## File map

| File | Responsibility |
|------|----------------|
| `internal/games/sokoban/module.go` | Run session state, progression, global undo pool, pause/help/final overlays, summaries |
| `internal/games/sokoban/module_test.go` | Deterministic tests for run session rules, undo pool, quit behavior, and summaries |
| `internal/games/sokoban/game.go` | Level-scoped board rules only; only touch if global undo migration needs a small helper |
| `internal/games/sokoban/game_test.go` | Update only if a small board-level helper changes |

---

### Task 1: Introduce explicit run session state

**Files:**
- Modify: `internal/games/sokoban/module.go`
- Modify: `internal/games/sokoban/module_test.go`

- [ ] **Step 1: Write the failing run-session tests**

Add to `internal/games/sokoban/module_test.go`:

```go
func TestNewRunSessionStartsAtLevelOneWithFourUndos(t *testing.T) {
	session := newRunSession(len(Levels))

	if session.currentLevel != 0 {
		t.Fatalf("expected run to start at level index 0, got %d", session.currentLevel)
	}
	if session.undosRemaining != 4 {
		t.Fatalf("expected run to start with 4 undos, got %d", session.undosRemaining)
	}
}

func TestRunSessionAdvanceMarksLevelAsCleared(t *testing.T) {
	session := newRunSession(10)

	session.completeLevel(12 * time.Second)

	if session.currentLevel != 1 {
		t.Fatalf("expected session to advance to level 1, got %d", session.currentLevel)
	}
	if session.completedLevels != 1 {
		t.Fatalf("expected one completed level, got %d", session.completedLevels)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/games/sokoban -run "TestNewRunSessionStartsAtLevelOneWithFourUndos|TestRunSessionAdvanceMarksLevelAsCleared"`

Expected: FAIL because `newRunSession` and its state do not exist yet.

- [ ] **Step 3: Implement the minimal run-session structure**

In `internal/games/sokoban/module.go`, add a local run state such as:

```go
type runResult string

const (
	runResultActive    runResult = "active"
	runResultAbandoned runResult = "abandoned"
	runResultCompleted runResult = "completed"
)

type runSession struct {
	currentLevel    int
	completedLevels int
	undosRemaining  int
	totalTime       time.Duration
	result          runResult
	levelCount      int
}

func newRunSession(levelCount int) runSession {
	return runSession{currentLevel: 0, undosRemaining: 4, result: runResultActive, levelCount: levelCount}
}

func (s *runSession) completeLevel(elapsed time.Duration) {
	s.totalTime += elapsed
	s.completedLevels++
	s.currentLevel++
	if s.completedLevels == s.levelCount {
		s.result = runResultCompleted
	}
}
```

- [ ] **Step 4: Re-run the tests to verify they pass**

Run: `go test ./internal/games/sokoban -run "TestNewRunSessionStartsAtLevelOneWithFourUndos|TestRunSessionAdvanceMarksLevelAsCleared"`

Expected: PASS

- [ ] **Step 5: Commit the run-session state**

```bash
git add internal/games/sokoban/module.go internal/games/sokoban/module_test.go
git commit -m "feat(sokoban): add run session state"
```

---

### Task 2: Move undo pressure from per-level to global run pool

**Files:**
- Modify: `internal/games/sokoban/module.go`
- Modify: `internal/games/sokoban/module_test.go`
- Modify: `internal/games/sokoban/game.go` only if a tiny helper is necessary

- [ ] **Step 1: Write the failing global-undo tests**

Add to `internal/games/sokoban/module_test.go`:

```go
func TestRunSessionUndoConsumesGlobalPool(t *testing.T) {
	session := newRunSession(10)

	session.consumeUndo()

	if session.undosRemaining != 3 {
		t.Fatalf("expected 3 undos remaining, got %d", session.undosRemaining)
	}
}

func TestRunSessionUndoPoolPersistsAcrossLevels(t *testing.T) {
	session := newRunSession(10)
	session.consumeUndo()
	session.completeLevel(5 * time.Second)

	if session.undosRemaining != 3 {
		t.Fatalf("expected undo usage to persist across levels, got %d", session.undosRemaining)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/games/sokoban -run "TestRunSessionUndoConsumesGlobalPool|TestRunSessionUndoPoolPersistsAcrossLevels"`

Expected: FAIL because global undo accounting does not exist yet.

- [ ] **Step 3: Implement the minimal global undo pool**

In `internal/games/sokoban/module.go`, add and use:

```go
func (s *runSession) consumeUndo() bool {
	if s.undosRemaining <= 0 {
		return false
	}
	s.undosRemaining--
	return true
}
```

Then route undo input through the session in gameplay flow so undo only happens when the run has remaining undos. Do not add any second pressure system.

- [ ] **Step 4: Re-run the tests to verify they pass**

Run: `go test ./internal/games/sokoban -run "TestRunSessionUndoConsumesGlobalPool|TestRunSessionUndoPoolPersistsAcrossLevels"`

Expected: PASS

- [ ] **Step 5: Commit the global undo pool**

```bash
git add internal/games/sokoban/module.go internal/games/sokoban/module_test.go internal/games/sokoban/game.go internal/games/sokoban/game_test.go
git commit -m "feat(sokoban): use a global undo pool for runs"
```

---

### Task 3: Reshape progression and pause around strict run rules

**Files:**
- Modify: `internal/games/sokoban/module.go`
- Modify: `internal/games/sokoban/module_test.go`

- [ ] **Step 1: Write the failing run-rule tests**

Add to `internal/games/sokoban/module_test.go`:

```go
func TestPauseOverlayDoesNotOfferRestartOrNextLevelInRunMode(t *testing.T) {
	module := Module{}
	model := module.buildPauseOverlay(0)

	ids := optionIDs(model.Options())
	if contains(ids, pauseOptionIDRestart) || contains(ids, pauseOptionIDNextLevel) {
		t.Fatalf("expected run-mode pause overlay without restart/next-level, got %v", ids)
	}
}

func TestRunAbandonStopsSessionWithoutProgression(t *testing.T) {
	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionPause}},
			{Actions: []input.Action{input.ActionDown}},
			{Actions: []input.Action{input.ActionSelect}},
		}},
	}

	if err := module.Run(context.Background()); err != nil {
		t.Fatalf("expected clean abandon exit, got %v", err)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/games/sokoban -run "TestPauseOverlayDoesNotOfferRestartOrNextLevelInRunMode|TestRunAbandonStopsSessionWithoutProgression"`

Expected: FAIL until pause semantics are simplified for run mode.

- [ ] **Step 3: Implement strict run-mode pause behavior**

In `internal/games/sokoban/module.go`:

- build pause overlay with only:
  - continue
  - abandon run
- remove current-level restart and next-level from run-mode pause behavior
- treat abandoning as run end, not level restart

Keep overlay rendering in the shared package; only change Sokoban’s declared options and interpretations.

- [ ] **Step 4: Re-run the tests to verify they pass**

Run: `go test ./internal/games/sokoban -run "TestPauseOverlayDoesNotOfferRestartOrNextLevelInRunMode|TestRunAbandonStopsSessionWithoutProgression"`

Expected: PASS

- [ ] **Step 5: Commit the run-mode progression rules**

```bash
git add internal/games/sokoban/module.go internal/games/sokoban/module_test.go
git commit -m "feat(sokoban): enforce strict run progression rules"
```

---

### Task 4: Add run-aware HUD and final summaries

**Files:**
- Modify: `internal/games/sokoban/module.go`
- Modify: `internal/games/sokoban/module_test.go`

- [ ] **Step 1: Write the failing summary/HUD tests**

Add to `internal/games/sokoban/module_test.go`:

```go
func TestBuildHelpOverlayExplainsGlobalUndoPool(t *testing.T) {
	module := Module{}
	model := module.buildHelpOverlay()

	joined := strings.Join(sectionLines(model.Sections()), "\n")
	if !strings.Contains(joined, "4 undos") {
		t.Fatalf("expected help overlay to explain the 4-undo run rule, got %q", joined)
	}
}

func TestBuildFinalSummaryOverlayReportsRunMetrics(t *testing.T) {
	module := Module{}
	session := newRunSession(len(Levels))
	session.completedLevels = 3
	session.totalTime = 45 * time.Second
	session.undosRemaining = 1
	session.result = runResultAbandoned

	model := module.buildFinalSummaryOverlay(session)
	joined := strings.Join(sectionLines(model.Sections()), "\n")
	if !strings.Contains(joined, "45") || !strings.Contains(joined, "1") {
		t.Fatalf("expected run summary metrics, got %q", joined)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/games/sokoban -run "TestBuildHelpOverlayExplainsGlobalUndoPool|TestBuildFinalSummaryOverlayReportsRunMetrics"`

Expected: FAIL until overlays are updated for run semantics.

- [ ] **Step 3: Implement run-aware help and final summaries**

Update `internal/games/sokoban/module.go` so:

- help overlay explains the 4-undo run rule and that quitting ends the run
- final summary builder accepts run-session metrics instead of only raw times
- success and failure summaries both report:
  - total time
  - levels completed / reached
  - undos used / remaining
  - final result

- [ ] **Step 4: Re-run the tests to verify they pass**

Run: `go test ./internal/games/sokoban -run "TestBuildHelpOverlayExplainsGlobalUndoPool|TestBuildFinalSummaryOverlayReportsRunMetrics"`

Expected: PASS

- [ ] **Step 5: Commit the run-aware UI metrics**

```bash
git add internal/games/sokoban/module.go internal/games/sokoban/module_test.go
git commit -m "feat(sokoban): add run-aware hud and summaries"
```

---

### Task 5: Final verification for Sokoban Run Mode

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

Only if needed:

```go
// Apply the smallest corrective change needed to restore the intended run-mode contract.
```

- [ ] **Step 4: Re-run verification after any fix**

Run: `go test ./internal/games/sokoban ./internal/engine/ui/overlay && go test ./...`

Expected: PASS

- [ ] **Step 5: Commit the verified run mode**

```bash
git add internal/games/sokoban/module.go internal/games/sokoban/module_test.go internal/games/sokoban/game.go internal/games/sokoban/game_test.go
git commit -m "feat(sokoban): introduce run mode"
```

---

## Spec coverage check

- strict linear progression: covered by Tasks 1 and 3
- 4 total undos for full run: covered by Task 2
- no selector / no replay: covered by Task 3
- run-aware HUD/help/summaries: covered by Task 4
- deterministic run tests: covered by Tasks 1-5

## Final checklist

- [ ] Sokoban run starts at level 1 with 4 total undos.
- [ ] Undo usage is global to the run.
- [ ] Pause overlay only exposes run-valid actions.
- [ ] Quitting ends the run and resets session progress.
- [ ] Final summaries report run metrics clearly.
- [ ] `go test ./...` passes.

## Next step

Execute this plan in a fresh Sokoban run-mode worktree and keep the changes local to Sokoban until the mode is stable and test-backed.
