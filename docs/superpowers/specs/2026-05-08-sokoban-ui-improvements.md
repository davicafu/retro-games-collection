# Sokoban UI Improvements - Spec

## Overview

Five improvements to Sokoban UI/UX: pause menu, help menu, undo indicator, fixed final table, and restart penalty.

## Changes

### 1. Pause Menu (P/ESC)

**When:** Player presses P or ESC during gameplay

**What:** Show pause menu overlay with options:
- Continue (C) - return to game
- Restart (R) - restart current level with 10s penalty
- Next Level (N) - skip to next level (not shown on last level)
- Quit (Q) - return to main menu

**Implementation:**
- Add `showPauseMenu()` method in `module.go`
- In `runLevel()`, handle `ActionPause` to show menu instead of returning
- Return result from menu: "continue", "restart", "next", "quit"
- If "restart", add 10s to elapsed time
- If "next", return early with current elapsed (not win)

### 2. Help Menu (?/h)

**When:** Player presses ? or h during gameplay

**What:** Show help overlay:

```
╔═══════════════════════════════════╗
║           AYUDA                   ║
╠═══════════════════════════════════╣
║  Objetivo:                        ║
║  Lleva todas las cajas a los      ║
║  puntos objetivos (.)             ║
║                                   ║
║  Controles:                       ║
║  ←↑↓→ / WASD  Mover               ║
║  U              Deshacer          ║
║  P / ESC       Menú de pausa     ║
║  ? / h         Ayuda             ║
║  Q              Salir             ║
╚═══════════════════════════════════╝
```

**Implementation:**
- Add `showHelp()` method in `module.go`
- In `runLevel()`, handle new `ActionHelp` input
- Wait for key press to dismiss

### 3. Undo Indicator

**When:** During gameplay

**What:** Update header to show: `Level X/10   ⏱ MM:SS.mmm   Undo: 1/1`

**Implementation:**
- Add `undoAvailable` field to `Game` struct (or method)
- In `drawLevel()`, check if undo is available and format indicator
- When undo is used, show `Undo: 0/1` (grayed out)

### 4. Final Table Borders

**When:** After completing all levels

**What:** Fix right border alignment. Currently:

```
║  Nivel 1:    00:05.234               ║
```

Should be:

```
║  Nivel 1:    00:05.234              ║
```

**Implementation:**
- Adjust padding in `showFinalSummary()` to make right border align
- Use dynamic width calculation: `boxW - len(line) - 1` for right padding

### 5. Restart Penalty

**When:** Player chooses "Restart" in pause menu

**What:** Add 10 seconds to level time

**Implementation:**
- Add parameter to `runLevel()`: `initialPenalty time.Duration`
- When restart is called, call `runLevel()` again with `initialPenalty = 10s`
- Accumulate penalty in elapsed time

## Input Actions Required

Add to `internal/engine/input/input.go`:
- `ActionHelp` - for ? and h keys

## Files Affected

| File | Changes |
|------|---------|
| `internal/engine/input/input.go` | Add ActionHelp constant |
| `internal/shell/model.go` | Map ?/h to ActionHelp |
| `internal/games/sokoban/game.go` | Add `UndoAvailable()` method |
| `internal/games/sokoban/module.go` | Add pause menu, help menu, undo indicator, fix borders, restart penalty |

## Success Criteria

- [ ] P/ESC shows pause menu with Continue/Restart/Next/Quit
- [ ] ?/h shows help screen with objective and controls
- [ ] Header shows "Undo: 1/1" or "Undo: 0/1"
- [ ] Final table borders align properly
- [ ] Restart adds 10s penalty to level time