# Tasks: Sokoban Game Implementation

## Phase 1: Foundation & Input System

- [x] **1.1** Add `ActionUndo` constant to `internal/engine/input/input.go`
  - Add `ActionUndo Action = "undo"` to the Action type constants
  - Ensure existing `ActionUp`, `ActionDown`, `ActionLeft`, `ActionRight` remain unchanged

- [x] **1.2** Verify input action constants compile correctly
  - Run `go build ./...` to ensure no syntax errors

## Phase 2: Core Game Logic

- [x] **2.1** Create `internal/games/sokoban/game.go`
  - Define `Game` struct with `Width`, `Height`, `Player`, `Boxes`, `Walls`, `Targets`
  - Implement `Move(dir input.Action) bool` — process move, push boxes, return true if moved
  - Implement `Undo()` — pop from undo stack and restore state
  - Implement `Won() bool` — check all boxes on targets
  - Implement helper methods: `InBounds`, `IsWall`, `IsBox`, `BoxAt`, `TargetAt`

- [x] **2.2** Create `internal/games/sokoban/levels.go`
  - Define embedded level strings (3 classic micro levels in Sokoban ASCII format)
  - Implement `ParseLevel(lines []string) (*Game, error)` parser
  - Support chars: `#` wall, `@` player, `$` box, `.` target, `*` box-on-target, `+` player-on-target, ` ` floor
  - Return error for invalid/empty levels

- [x] **2.3** Create `internal/games/sokoban/undo.go`
  - Define `undoStack` type as `[]Game` (slice of deep-copied game states)
  - Implement `Push(g *Game)` — deep copy current state, append to stack
  - Implement `Pop() *Game` — return last state, remove from stack
  - Implement `Len() int` — return stack size
  - Cap stack at 50 entries

- [x] **2.4** Create `internal/games/sokoban/draw.go`
  - Implement `Draw(screen render.Screen, g *Game)` function
  - Render chars: `#` walls, `.` targets, `$` boxes, `@` player, `●` box-on-target
  - Use single-pass rendering (no Z-order layers)

- [x] **2.5** Create `internal/games/sokoban/module.go`
  - Define `Module` struct with `Screen render.Screen` and `Provider input.Provider`
  - Implement `NewModule(screen render.Screen, provider input.Provider) Module`
  - Implement `Run(ctx context.Context) error` — custom input-driven loop:
    ```
    loop:
      for {
        select {
        case <-ctx.Done():
          return ctx.Err()
        default:
          action := provider.Poll()
          switch action {
          case input.ActionUp/Down/Left/Right:
            game.Move(action)
            pushUndo()
          case input.ActionUndo:
            game.Undo()
          case input.ActionQuit, input.ActionEscape:
            break loop
          }
          if game.Won() { /* handle win, break */ }
          screen.Clear()
          draw.Draw(screen, &game)
          screen.Present()
      }
    ```

## Phase 3: Shell Integration

- [x] **3.1** Modify `internal/shell/model.go`
  - Add `"sokoban"` to the `games` slice in MainMenu
  - Add `runSokobanGame(model, msg)` handler case
  - Implement handler that creates `input.Static` provider, initializes module, runs game

## Phase 4: Testing

- [x] **4.1** Create `internal/games/sokoban/game_test.go`
  - Table-driven tests for: move into wall, push box, push box into wall, push box into box, win condition, undo stack
  - Use `input.Static` provider for deterministic input sequences
  - Test level parser with valid/invalid ASCII levels

- [x] **4.2** Create `internal/games/sokoban/levels_test.go`
  - Test `ParseLevel` with valid level strings
  - Test error cases (empty, malformed)

- [x] **4.3** Run all tests
  - Execute `go test ./internal/games/sokoban/...`
  - Ensure all tests pass

## Phase 5: Verification

- [x] **5.1** Run `go vet ./...`
- [x] **5.2** Run `gofmt -d ./internal/games/sokoban/`
- [ ] **5.3** Verify shell menu shows Sokoban entry
- [ ] **5.4** Play-test the game end-to-end (manual verification)

## Review Workload

```
sokoban$ git diff --stat
 internal/engine/input/input.go      |  1 +
 internal/games/sokoban/draw.go      | NEW FILE (~50 lines)
 internal/games/sokoban/game.go      | NEW FILE (~100 lines)
 internal/games/sokoban/levels.go    | NEW FILE (~80 lines)
 internal/games/sokoban/module.go    | NEW FILE (~60 lines)
 internal/games/sokoban/undo.go      | NEW FILE (~30 lines)
 internal/games/sokoban/*_test.go     | NEW FILES (~150 lines total)
 internal/shell/model.go             | ~10 lines changed
----------------------------------------------------------------
 8 files changed, ~8 insertions, ~470 lines
```

### Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Undo state memory bloat | Low | Stack capped at 50 |
| Box push edge cases (push into box) | Medium | Comprehensive unit tests |
| Level parsing edge cases | Low | Test valid/invalid formats |