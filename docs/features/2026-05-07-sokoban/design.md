# Design: Sokoban

## Technical Approach

Sokoban is a turn-based, input-driven game — no tick/clock needed. `Module.Run` implements its own loop: `provider.Poll()` → process move → check win → render → repeat. This contrasts with Snake's `runtime.Runtime` tick-based approach, validating that the engine supports heterogeneous game loops.

## Architecture Decisions

### Decision: Custom input-driven loop (no runtime.Runtime)

**Choice**: Sokoban's `Module.Run` manages its own loop directly.
**Alternatives considered**: Wrap Sokoban in `runtime.Runtime` and drive it with a dummy ticker.
**Rationale**: Sokoban moves only on input — forcing it into a tick model adds unnecessary complexity. Snake needs a clock for autonomous movement; Sokoban does not.

### Decision: Full-state snapshot undo (not command-based)

**Choice**: Undo stack stores deep copies of the full `GameState`.
**Alternatives considered**: Store only deltas (player delta + box deltas).
**Rationale**: Simpler to reason about, easier to test, no edge cases in reverse-application. Max 50 entries keeps memory bounded. Sokoban state is small enough that copying is negligible cost.

### Decision: Single game state struct (not separate entities)

**Choice**: All state (player, boxes, walls, targets, width, height) lives in one `Game` struct.
**Alternatives considered**: Separate Player, Box, Wall types with positions.
**Rationale**: Sokoban state is small and co-located. A flat struct with `[]grid.Point` slices is idiomatic for grid-based games in this codebase (see Snake's `Game`).

### Decision: Box-over-target shown as `●`, not separate layer

**Choice**: Render each cell as ONE character; if box occupies target cell, show `●`.
**Alternatives considered**: Render boxes then targets as two layers (■ on . = ●).
**Rationale**: Single-pass rendering, no Z-order complexity. The visual result is identical.

## Data Flow

```
Provider.Poll() → Game.Move(dir) → [undo push] → Game.Won()?
    → Screen.Clear() → Draw(walls,targets,boxes,player) → Screen.Present()
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/games/sokoban/module.go` | Create | `Module` struct + `NewModule` + `Run` loop |
| `internal/games/sokoban/game.go` | Create | `Game` state, `Move`, `Won`, `Undo` |
| `internal/games/sokoban/levels.go` | Create | Level strings + `ParseLevel` |
| `internal/games/sokoban/draw.go` | Create | `Draw(screen, game)` |
| `internal/games/sokoban/game_test.go` | Create | Unit tests for move/push/win/undo |
| `internal/engine/input/input.go` | Modify | Add `ActionUndo` constant |
| `internal/shell/model.go` | Modify | Register Sokoban entry in `games` slice + `runSokobanGame` |

## Interfaces / Contracts

```go
// internal/games/sokoban/game.go
type Game struct {
    Width   int
    Height  int
    Player  grid.Point
    Boxes   []grid.Point   // mutable — order-free set via position lookup
    Walls   []grid.Point   // static per level
    Targets []grid.Point   // static per level
}

func (g *Game) Move(dir input.Action) bool // returns true if move happened
func (g *Game) Undo()
func (g *Game) Won() bool
func (g *Game) InBounds(p grid.Point) bool
func (g *Game) IsWall(p grid.Point) bool
func (g *Game) IsBox(p grid.Point) bool
```

```go
// internal/games/sokoban/levels.go
func ParseLevel(lines []string) (*Game, error)

// Render chars: '#' wall, '@' player, '$' box, '.' target, '*' box-on-target, '+' player-on-target, ' ' floor
// Convert to internal representation on parse.
```

```go
// internal/games/sokoban/draw.go
func Draw(screen render.Screen, g *Game)
```

```go
// internal/games/sokoban/module.go
type Module struct {
    Screen   render.Screen
    Provider input.Provider
}

func (m Module) Run(ctx context.Context) error
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Move into wall, push box, push box into wall, push box into box, win condition, undo stack | Table-driven tests against `Static([]Action{...})` input provider |
| Unit | Level parser (valid levels, invalid levels) | Tests against known ASCII level strings |
| Unit | Render output (screen cell assertions via `BufferScreen.Rows()`) | Golden-ish assertions on rendered grid |

## Migration / Rollout

No migration required — new package only.

## Open Questions

None — all decisions resolved in proposal.