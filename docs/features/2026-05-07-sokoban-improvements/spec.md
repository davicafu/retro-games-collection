# Sokoban Improvements - Spec

## Overview

Three enhancements to Sokoban: final summary with per-level times, harder 10-level progression, and single undo per level.

## Changes

### 1. Final Summary Screen

**When:** After completing level 10

**What:** Display a table showing:
- Time for each level (format: MM:SS.mmm)
- Total time sum

**Implementation:**
- In `module.go`, store `elapsed` times in a `[]time.Duration` slice
- After level 10 completes, call `showFinalSummary(times)` instead of `showLevelComplete`
- New method `showFinalSummary` renders centered box with all times

### 2. 10 Harder Levels

**When:** Initial implementation

**What:** Replace current 10 levels with harder versions:
- More boxes/targets (3-5 per level)
- Complex layouts requiring planning
- Narrow corridors, dead ends
- Multiple solutions but many lead to deadlocks

**Implementation:**
- Update `levels.go` with new level strings
- Keep the same ASCII format (@, $, ., #)
- Ensure each level is solvable

### 3. Single Undo Per Level

**When:** During gameplay

**What:** Player can only undo once per level. After using it, pressing 'U' does nothing.

**Implementation:**
- Add `undoUsed bool` field to `Game` struct in `game.go`
- Add `ResetUndo()` method to reset the flag
- Modify `Undo()` method: if `undoUsed == true`, return early without doing anything
- In `module.go`, call `game.ResetUndo()` at start of each level

## Files Affected

| File | Changes |
|------|---------|
| `internal/games/sokoban/levels.go` | Replace 10 levels with harder versions |
| `internal/games/sokoban/game.go` | Add `undoUsed` field and check in `Undo()` |
| `internal/games/sokoban/module.go` | Store times, call reset, show final summary |

## Success Criteria

- [ ] After level 10, summary shows all 10 level times + total
- [ ] All 10 levels are solvable and harder than current ones
- [ ] Undo works once per level; second attempt does nothing
- [ ] Undo resets when starting a new level