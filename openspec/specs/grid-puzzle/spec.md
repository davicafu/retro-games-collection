# Sokoban Specifications

> Capability: sokoban-game | Domain: grid-puzzle, undo, levels | Status: draft

## Overview

Sokoban is a classic puzzle game where the player pushes boxes onto target locations. This spec defines the full implementation for the Sokoban game capability, organized into three domains: grid-puzzle (game state and mechanics), undo (history management), and levels (level data).

This is a FULL spec for a new game capability, not a delta from Snake.

---

## Domain: grid-puzzle

### Entities

| Entity | Symbol | Description |
|--------|--------|-------------|
| Floor | ` ` (space) | Empty walkable cell |
| Wall | `#` | Impassable barrier |
| Player | `@` | Controlled character |
| Box | `$` | Pushable crate |
| Target | `.` | Goal for boxes |
| Box-on-target | `*` | Box successfully placed |

### Game State

```go
type GameState struct {
    grid        [][]Cell      // 2D grid of cells
    player     Point        // Player position (X, Y)
    boxes      []Point      // Current box positions
    targets    []Point     // Target positions (immutable)
    width      int
    height    int
}
```

- **grid**: 2D slice where `grid[y][x]` holds the cell type
- **player**: Current position as `grid.Point`
- **boxes**: Mutable slice of positions, updated on push
- **targets**: Immutable, set at level load time

### Cell Types

```go
type Cell int

const (
    CellFloor      Cell = iota // 0: empty floor
    CellWall               // 1: wall
    CellTarget             // 2: target (floor underneath)
    CellBoxOnTarget       // 3: box on target
)
```

Note: Player position is stored separately, not in grid. Boxes on targets use both `boxes` slice and cell state to track covered targets.

### Input Actions

| Action | Trigger | Purpose |
|--------|---------|---------|
| `ActionUp` | Arrow Up / W | Move player north |
| `ActionDown` | Arrow Down / S | Move player south |
| `ActionLeft` | Arrow Left / A | Move player west |
| `ActionRight` | Arrow Right / D | Move player east |
| `ActionUndo` | U / Ctrl+Z | Undo last move |
| `ActionQuit` | Q / Escape | Return to menu |
| `ActionRestart` | R | Restart level |

### Movement Rules

1. **Simple Move**: Player moves to adjacent cell if floor or target (no box)
2. **Box Push**: Player moves + 1 adjacent box if:
   - Box's forward cell is floor or target (not wall, not another box)
3. **Invalid**: Wall, another box blocking push, or out of bounds

```go
func (g *GameState) Move(d Direction) (pushed bool, err error) {
    target := g.player + d.Offset()
    beyond := target + d.Offset()
    
    // Check bounds
    if !g.inBounds(target) {
        return false, ErrOutOfBounds
    }
    
    // Get cell at target
    cell := g.grid[target.Y][target.X]
    
    // Wall check
    if cell == CellWall {
        return false, ErrWall
    }
    
    // Box at target?
    boxIdx := g.boxAt(target)
    if boxIdx >= 0 {
        // Try to push box
        if !g.inBounds(beyond) {
            return false, ErrOutOfBounds
        }
        beyondCell := g.grid[beyond.Y][beyond.X]
        if beyondCell == CellWall || g.boxAt(beyond) >= 0 {
            return false, ErrCannotPush
        }
        // Push succeeds
        g.boxes[boxIdx] = beyond
        g.updateCellsForBoxes()
        return true, nil
    }
    
    // Simple move
    g.player = target
    return false, nil
}
```

### Win Condition

Win when all box positions overlap target positions.

```go
func (g *GameState) Won() bool {
    for _, box := range g.boxes {
        if !g.isTargetAt(box) {
            return false
        }
    }
    return len(g.boxes) > 0
}
```

### Turn-Based Loop

Unlike Snake's tick-based loop, Sokoban waits for each input:

```
┌─────────────────────────┐
│ 1. Wait for input       │◄──────────┐
│ 2. Parse action         │            │
│ 3. Execute move        │            │
│ 4. Check win condition  │            │
│ 5. Render frame         │────────────┘
│ 6. Repeat unless quit  │
└─────────────────────────┘
```

---

## Domain: undo

### Snapshot Structure

```go
type Snapshot struct {
    player Point
    boxes  []Point
    // Grid state derived from boxes positions
}
```

### Undo Stack

```go
type UndoStack struct {
    snapshots []Snapshot
    maxDepth  int // Default: 50
}
```

### Behavior

1. **Push on valid move**: Before executing any move, capture snapshot
2. **Pop on undo**: Restore to previous snapshot, discard current
3. **Clear on restart**: Clear stack when level restarts
4. **Limit**: Enforce `maxDepth` — drop oldest when full (FIFO)

### Implementation

```go
func (u *UndoStack) Push(state *GameState) {
    if len(u.snapshots) >= u.maxDepth {
        // Drop oldest
        copy(u.snapshots, u.snapshots[1:])
        u.snapshots = u.snapshots[:u.maxDepth-1]
    }
    u.snapshots = append(u.snapshots, Snapshot{
        player: state.player,
        boxes:  append([]Point{}, state.boxes...),
    })
}

func (u *UndoStack) Pop(state *GameState) bool {
    if len(u.snapshots) == 0 {
        return false
    }
    last := u.snapshots[len(u.snapshots)-1]
    u.snapshots = u.snapshots[:len(u.snapshots)-1]
    state.player = last.player
    state.boxes = append([]Point{}, last.boxes...)
    state.updateCellsForBoxes()
    return true
}
```

---

## Domain: levels

### Format

Classic Sokoban ASCII format:

```
#####
#@$.#
#####
```

| Char | Meaning |
|------|---------|
| `#` | Wall |
| `@` | Player |
| `$` | Box |
| `.` | Target |
| `*` | Box on target |
| ` ` | Floor |

### Level Data

Embedded as Go string constants:

```go
var levels = []string{
    // Level 1: Simple
    `
#####
#@$.#
#####
`,
    // Level 2: Two boxes
    `
  #####
###   #
#.@$  #
###   #
  #####
`,
    // Level 3: Classic micro
    `
  #####
  #   #
  #$  #
###  $##
#  $ $ #
# # .  #
#   .  #
########
`,
}
```

### Parsing

```go
func ParseLevel(data string) (*GameState, error) {
    lines := strings.Split(strings.TrimSpace(data), "\n")
    height := len(lines)
    width := max(len(lines[0]), len(lines[len(lines)-1]))
    
    grid := make([][]Cell, height)
    var boxes, targets []Point
    var player Point
    
    for y, line := range lines {
        grid[y] = make([]Cell, width)
        for x, ch := range line {
            switch ch {
            case '#':
                grid[y][x] = CellWall
            case '@':
                player = Point{X: x, Y: y}
            case '$':
                boxes = append(boxes, Point{X: x, Y: y})
            case '.':
                targets = append(targets, Point{X: x, Y: y})
                grid[y][x] = CellTarget
            case '*':
                boxes = append(boxes, Point{X: x, Y: y})
                targets = append(targets, Point{X: x, Y: y})
                grid[y][x] = CellBoxOnTarget
            default:
                grid[y][x] = CellFloor
            }
        }
    }
    
    return &GameState{
        grid:   grid,
        player: player,
        boxes:  boxes,
        targets: targets,
        width:  width,
        height: height,
    }, nil
}
```

### Level Selection

- Start at Level 1
- Advance to next level on win
- No level picker UI (out of scope)

---

## Integration

### Module Entry Point

```go
type Module struct {
    screen   render.Screen
    provider input.Provider
    levels   []string
}

func NewModule(screen render.Screen, provider input.Provider) Module {
    return Module{
        screen:   screen,
        provider: provider,
        levels:   levels,
    }
}

func (m Module) Run(ctx context.Context) error {
    levelIdx := 0
    for {
        state, err := ParseLevel(m.levels[levelIdx])
        if err != nil {
            return err
        }
        
        undo := NewUndoStack(50)
        
        if err := m.playLevel(ctx, state, &undo); err != nil {
            return err
        }
        
        // Check if won
        if state.Won() {
            levelIdx++
            if levelIdx >= len(m.levels) {
                // All levels complete
                return nil
            }
            // Continue to next level
            continue
        }
        
        // Return to menu on quit
        return nil
    }
}
```

### Rendering

```go
func Render(screen render.Screen, state *GameState) {
    screen.Clear()
    
    // Draw grid
    for y := 0; y < state.height; y++ {
        for x := 0; x < state.width; x++ {
            cell := state.grid[y][x]
            var r rune
            switch cell {
            case CellWall:
                r = '#'
            case CellTarget:
                r = '.'
            case CellBoxOnTarget:
                r = '*'
            default:
                r = ' '
            }
            screen.SetCell(x, y, render.Cell{Rune: r})
        }
    }
    
    // Draw boxes
    for _, box := range state.boxes {
        r := '$'
        if state.isTargetAt(box) {
            r = '*'
        }
        screen.SetCell(box.X, box.Y, render.Cell{Rune: r})
    }
    
    // Draw player
    screen.SetCell(state.player.X, state.player.Y, render.Cell{Rune: '@'})
    
    screen.Present()
}
```

---

## Testing Requirements

| Test | Coverage |
|------|----------|
| `Move` | Simple move, box push, wall collision, box-blocked push |
| `Won` | All boxes on targets, some boxes off targets |
| `UndoStack` | Push, pop, max depth, clear |
| `ParseLevel` | All entity types, dimensions |
| `Render` | All entity types rendered correctly |

---

## Acceptance Criteria

- [ ] Player can move in four directions with arrow keys or WASD
- [ ] Player can push single boxes
- [ ] Player cannot push boxes into walls or other boxes
- [ ] Win triggers when all boxes are on targets
- [ ] Undo (U or Ctrl+Z) reverts the last move
- [ ] Restart (R) resets current level and clears undo stack
- [ ] Quit (Q) returns to menu
- [ ] Level advances on win
- [ ] All game logic has unit test coverage
- [ ] Code passes `go vet` and `gofmt`

---

## File Structure

```
internal/games/sokoban/
├── module.go      # Module + Run entry point
├── game.go      # GameState + Move + Won
├── state.go     # Cell types + grid management
├── level.go    # ParseLevel + embedded levels
├── undo.go     # UndoStack + Snapshot
├── render.go   # Render to Screen
└── *_test.go   # Unit tests
```

---

## Dependencies

- `internal/engine/grid` — `grid.Point`
- `internal/engine/render` — `render.Screen`
- `internal/engine/input` — `input.Action`, `input.Provider`
- No external packages required