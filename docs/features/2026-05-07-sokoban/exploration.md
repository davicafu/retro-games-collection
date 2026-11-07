# Exploration: Sokoban Game Implementation

## Current State

The project has a working Snake game that serves as the reference implementation:
- **Module pattern**: `NewModule(clock, screen, provider, selector)` + `Run(ctx)` method
- **Game interface**: Uses `runtime.Game` with `Init`, `Update`, `Render`, `ShouldQuit`
- **Clock**: Tick-based game loop via `clock.Clock` interface (ticks channel)
- **Input**: Actions flow via `input.Provider` → `FrameInput{Actions []Action}`
- **Rendering**: `render.Screen` interface with `Clear`, `SetCell`, `Present`
- **Testing**: Uses injectable clock (`clock.Manual`), scripted input (`Static`), headless screen

The runtime loop in `internal/engine/runtime/runtime.go` drives game ticks via the clock, calling `Update` on each tick. Snake uses this tick-based loop.

## Affected Areas

| File | Role | Why Relevant |
|------|------|------------|
| `internal/games/snake/module.go` | Module pattern | Template to follow |
| `internal/games/snake/game.go` | Game logic | Turn-based state management |
| `internal/engine/input/input.go` | Actions | Need `ActionUndo` for Sokoban |
| `internal/engine/runtime/runtime.go` | Game loop | **MAJOR CHANGE**: Sokoban is NOT tick-based |
| `internal/shell/model.go` | Shell integration | Need to register Sokoban game entry |
| `internal/engine/grid/point.go` | Grid primitive | Can reuse `grid.Point` (alias for Position) |

## Key Differences: Snake vs Sokoban

| Aspect | Snake | Sokoban |
|--------|------|--------|
| Timing | Tick-based (120ms) | Input-driven (no clock ticks) |
| Loop | Auto-advances each tick | Waits for player input to move |
| Entities | Snake body + food | Player, boxes, targets, walls |
| Win | High score (endless) | All boxes on targets |
| Levels | Procedural | Defined data |
| Undo | No | Yes (full undo stack) |

## Approaches

### 1. Reuse Runtime with No-Op Clock (RECOMMENDED)
**Description**: Create a `clock.Fake` that never ticks, or a variant of `runtime.Runtime` that runs input-driven without a clock.

- Pros: Minimal code, follows existing patterns
- Cons: Might feel like forcing a square peg into a round hole
- Effort: **Low** — ~20 lines of new clock type

### 2. Separate Runtime for Input-Driven Games
**Description**: Create `internal/engine/runtime/inputdriven.go` that only advances on input, not ticks.

- Pros: Clean separation of concerns
- Cons: Duplicates interface with runtime.Runtime
- Effort: **Medium** — new file, similar methods

### 3. Sokoban's Own Loop in Module
**Description**: Module.Run implements its own loop, bypassing runtime.Runtime entirely.

- Pros: Full control, simplest integration
- Coupling: Game logic tied to module, less reusable runtime
- Effort: **Low** — similar to what shell does

## Engine Capabilities to Reuse

| Capability | File | Reuse? |
|-----------|------|--------|
| `grid.Point` | `internal/engine/grid/point.go` | YES — Position alias |
| `render.Screen` | `internal/engine/render/buffer.go` | YES |
| `input.Action` | `internal/engine/input/input.go` | Need `ActionUndo` |
| `Clock` | `internal/engine/clock/clock.go` | SKIP — not tick-based |

## What Needs to Be Built

### New Files (Sokoban Package)

| File | Purpose |
|------|---------|
| `internal/games/sokoban/module.go` | NewModule + Run entry point |
| `internal/games/sokoban/game.go` | Game state, move validation, push logic |
| `internal/games/sokoban/state.go` | Grid + entities encoding |
| `internal/games/sokoban/level.go` | Level loading from embedded data |
| `internal/games/sokoban/undo.go` | Undo stack implementation |
| `internal/games/sokoban/render.go` | Screen drawing |
| `internal/games/sokoban/{game,module,render}_test.go` | Unit tests |

### Level Data Options

1. **Embedded Go strings** — Simplest, no external files
2. **Separate .txt files** — Requires file loading
3. **Level selector** — First level or level picker

**Recommendation**: Embedded Go strings (Sokoban classic micro levels)

### Entities & Display

| Entity | Rune | Color |
|--------|------|------|
| Player | ☺ or @ | yellow |
| Box | ▣ or $ | brown/white |
| Target | × or . | green |
| Wall | █ or # | gray |
| Box on target | ■ or * | green |

## Risks

1. **Runtime mismatch**: Snake's runtime expects tick-based; Sokoban needs input-driven. Need new runtime variant or bypass.
2. **Undo state**: Must capture full game state (grid + player + boxes) for undo — needs serialization or deep copy.
3. **Win detection**: All boxes must be on targets — check after each move.
4. **Level data**: Need to define 1-3 test levels in a format that's easy to parse.

## Ready for Proposal

**YES** — The exploration is complete. Key findings:

- Sokoban is fundamentally different from Snake (input-driven vs tick-based)
- Recommend creating a "passthrough" clock that never ticks, OR Sokoban implements its own loop in Module.Run
- Minimal new engine work needed — mostly new Sokoban package
- Need to add `ActionUndo` to input actions
- Can reuse grid.Point, render.Screen, input.Provider

The orchestrator should now proceed to **sdd-propose** to create the change proposal.