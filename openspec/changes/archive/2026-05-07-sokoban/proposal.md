# Proposal: Sokoban Game Implementation

## Intent

Add Sokoban as the second playable game in the CLI games collection, demonstrating an input-driven (turn-based) game model contrasting with Snake's tick-based approach. This validates the engine's ability to support different game loop patterns while reusing shared infrastructure.

## Scope

### In Scope
- Sokoban game module with `NewModule` + `Run` entry point
- Game state management (grid, player, boxes, targets)
- Push mechanics with wall/box collision detection
- Win condition detection (all boxes on targets)
- Undo stack with full state snapshots
- Embedded level data (1-3 classic micro levels)
- Rendering to `render.Screen`
- Undo action (`ActionUndo`) in engine/input
- Shell menu registration
- Unit tests for game logic

### Out of Scope
- Level editor or custom level loading
- Animation/visual effects
- Time-based mechanics (Sokoban is input-driven)
- Score tracking or leaderboards
- Mobile-responsive controls

## Capabilities

### New Capabilities
- `sokoban-game`: Full Sokoban game implementation with input-driven loop
- `input-undo`: Undo action support in the input action type system

### Modified Capabilities
- None — Snake's tick-based runtime remains unchanged

## Approach

Implement Sokoban with its own input-driven loop inside `Module.Run`, bypassing the tick-based `runtime.Runtime` used by Snake. This approach:
- Keeps the game loop simple (wait → process input → render)
- Avoids forcing a square peg (no-ticks) into a round hole (tick-based runtime)
- Allows full control over state transitions

Undo uses full state snapshots (deep copy of grid + entities) pushed onto a stack. Level data is embedded as Go string constants using classic Sokoban ASCII format.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/engine/input/input.go` | Modified | Add `ActionUndo` to Action type |
| `internal/games/sokoban/module.go` | New | Module entry point with Run loop |
| `internal/games/sokoban/game.go` | New | Game state, move logic, win check |
| `internal/games/sokoban/state.go` | New | Grid + entities encoding |
| `internal/games/sokoban/level.go` | New | Level data parsing |
| `internal/games/sokoban/undo.go` | New | Undo stack management |
| `internal/games/sokoban/render.go` | New | Screen rendering |
| `internal/shell/model.go` | Modified | Register Sokoban in menu |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Undo state memory bloat | Low | Limit undo stack to 50 moves |
| Level data format fragility | Low | Use well-tested ASCII format, add parse tests |
| Missing edge case in push logic | Med | Comprehensive unit tests for box pushing |

## Rollback Plan

1. Delete `internal/games/sokoban/` directory
2. Revert `internal/engine/input/input.go` (remove `ActionUndo`)
3. Remove Sokoban registration from `internal/shell/model.go`
4. Verify `go build` and tests pass

## Dependencies

- None — no external packages required for Sokoban logic

## Success Criteria

- [ ] Sokoban launches from shell menu and is playable
- [ ] Player can move and push boxes
- [ ] Win triggers when all boxes reach targets
- [ ] Undo reverts last move correctly (min 10 moves)
- [ ] All game logic has unit test coverage
- [ ] Code passes `go vet` and `gofmt`