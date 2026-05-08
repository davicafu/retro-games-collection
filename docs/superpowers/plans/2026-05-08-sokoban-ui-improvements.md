# Sokoban UI Improvements Implementation Plan

> **For agentic workers:** Execute tasks in order. Steps use checkbox syntax.

**Goal:** Add pause menu, help menu, undo indicator, fix final table borders, and add restart penalty

**Tech Stack:** Go, existing Sokoban game engine

---

## Task 1: Add ActionHelp Input

**Files:**
- Modify: `internal/engine/input/input.go`
- Modify: `internal/shell/model.go`

- [ ] **Step 1: Add ActionHelp constant**

In `internal/engine/input/input.go`, add to the Action type and values:

```go
// In the Action type definition (around line 8-15):
const (
	ActionUp Action = "up"
	ActionDown Action = "down"
	ActionLeft Action = "left"
	ActionRight Action = "right"
	ActionQuit Action = "quit"
	ActionSelect Action = "select"
	ActionPause Action = "pause"
	ActionUndo Action = "undo"
	ActionHelp Action = "help"  // NEW
)
```

- [ ] **Step 2: Map ?/h keys in shell model**

In `internal/shell/model.go`, in `updatePlaying()` around line 174-191, add:

```go
case "?", "h", "H":
	m.chInput.Send(input.ActionHelp)
```

- [ ] **Step 3: Test**

Run: `go build ./...`

- [ ] **Step 4: Commit**

```bash
git add internal/engine/input/input.go internal/shell/model.go
git commit -m "feat(input): add ActionHelp for help menu"
```

---

## Task 2: Add UndoAvailable Method

**Files:**
- Modify: `internal/games/sokoban/game.go`

- [ ] **Step 1: Add UndoAvailable method**

Add after `ResetUndo()` method:

```go
func (g *Game) UndoAvailable() bool {
	return !g.undoUsed
}
```

- [ ] **Step 2: Test**

Run: `go build ./...`

- [ ] **Step 3: Commit**

```bash
git add internal/games/sokoban/game.go
git commit -m "feat(sokoban): add UndoAvailable method"
```

---

## Task 3: Undo Indicator in Header

**Files:**
- Modify: `internal/games/sokoban/module.go`

- [ ] **Step 1: Update drawLevel to show undo indicator**

In `module.go`, modify `drawLevel()` around line 107:

```go
func (m Module) drawLevel(game *Game, levelIdx int, elapsed time.Duration) {
	m.Screen.Clear()

	undoStr := "Undo: 1/1"
	if !game.UndoAvailable() {
		undoStr = "Undo: 0/1"  // Could use gray color
	}
	header := fmt.Sprintf("Level %d/%d   ⏱ %s   %s", levelIdx+1, len(Levels), fmtDuration(elapsed), undoStr)
	render.WriteText(m.Screen, 0, 0, header)

	// Draw the game board offset by 1 row down
	DrawAt(m.Screen, game, 0, 1)
}
```

- [ ] **Step 2: Test**

Run: `go build ./...`

- [ ] **Step 3: Commit**

```bash
git add internal/games/sokoban/module.go
git commit -m "feat(sokoban): add undo indicator to header"
```

---

## Task 4: Help Menu

**Files:**
- Modify: `internal/games/sokoban/module.go`

- [ ] **Step 1: Add showHelp method**

Add after `drawLevel()` method (around line 112):

```go
func (m Module) showHelp() error {
	m.Screen.Clear()

	helpText := `╔══════════════════════════════════════════╗
║               AYUDA                    ║
╠════════════════════════════════════════╣
║  Objetivo:                              ║
║  Lleva todas las cajas a los           ║
║  puntos objetivos (.)                  ║
║                                         ║
║  Controles:                            ║
║  ←↑↓→ / WASD  Mover                    ║
║  U              Deshacer (1 por nivel) ║
║  P / ESC       Menú de pausa          ║
║  ? / h         Esta ayuda             ║
║  Q              Salir al menú          ║
╚════════════════════════════════════════╝`

	sw, sh := m.Screen.Size()
	lines := strings.Split(helpText, "\n")
	boxH := len(lines)
	boxW := 44
	boxY := (sh - boxH) / 2
	boxX := (sw - boxW) / 2

	for i, line := range lines {
		render.WriteText(m.Screen, boxX, boxY+i, line)
	}

	if err := m.Screen.Present(); err != nil {
		return err
	}

	// Wait for any key to dismiss
	for {
		frame := m.Provider.Poll()
		if len(frame.Actions) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}
```

- [ ] **Step 2: Handle ActionHelp in runLevel**

In `runLevel()`, add case for `ActionHelp`:

```go
case input.ActionHelp:
	m.showHelp()
	m.drawLevel(game, levelIdx, time.Since(start))
```

- [ ] **Step 3: Test**

Run: `go build ./...`

- [ ] **Step 4: Commit**

```bash
git add internal/games/sokoban/module.go
git commit -m "feat(sokoban): add help menu (?/h keys)"
```

---

## Task 5: Pause Menu with Restart Penalty

**Files:**
- Modify: `internal/games/sokoban/module.go`

- [ ] **Step 1: Add showPauseMenu method**

Add after `showHelp()` method:

```go
type pauseResult int

const (
	pauseContinue pauseResult = iota
	pauseRestart
	pauseNext
	pauseQuit
)

func (m Module) showPauseMenu() pauseResult {
	m.Screen.Clear()

	boxW := 32
	boxH := 8
	sw, sh := m.Screen.Size()
	boxX := (sw - boxW - 2) / 2
	boxY := (sh - boxH - 2) / 2

	render.DrawBox(m.Screen, boxX, boxY, boxW, boxH)

	title := "PAUSA"
	render.WriteText(m.Screen, boxX+1+(boxW-len(title))/2, boxY+1, title)

	options := []string{"C - Continuar", "R - Reiniciar (+10s)", "N - Siguiente nivel", "Q - Salir"}
	for i, opt := range options {
		render.WriteText(m.Screen, boxX+2, boxY+3+i, opt)
	}

	if err := m.Screen.Present(); err != nil {
		return pauseQuit
	}

	// Wait for key
	for {
		frame := m.Provider.Poll()
		for _, action := range frame.Actions {
			switch action {
			case input.ActionUp:
				// Could implement cursor navigation, but for now just handle keys
			case input.ActionDown:
			case input.ActionSelect, input.ActionUp, input.ActionDown:
				// Show cursor effect - simplified for now
			}
			// Check for specific keys
			for _, r := range frame.Runes {
				switch string(r) {
				case "c", "C":
					return pauseContinue
				case "r", "R":
					return pauseRestart
				case "n", "N":
					return pauseNext
				case "q", "Q":
					return pauseQuit
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}
```

Note: You'll also need to check if current level is last to hide "Next Level" option.

- [ ] **Step 2: Modify runLevel to handle pause and restart**

Update `runLevel()` signature to accept initial penalty:

```go
func (m Module) runLevel(ctx context.Context, game *Game, levelIdx int, initialPenalty time.Duration) (time.Duration, error) {
	start := time.Now().Add(-initialPenalty)  // Apply penalty
```

- [ ] **Step 3: Handle pause in runLevel loop**

Replace `case input.ActionQuit, input.ActionPause:` with pause menu handling:

```go
case input.ActionPause:
	result := m.showPauseMenu()
	switch result {
	case pauseContinue:
		m.drawLevel(game, levelIdx, time.Since(start))
	case pauseRestart:
		game.ResetUndo()
		start = time.Now().Add(-10 * time.Second)  // Reset with 10s penalty
		m.drawLevel(game, levelIdx, time.Since(start))
	case pauseNext:
		return time.Since(start), nil  // Return early, not a "win"
	case pauseQuit:
		return 0, nil
	}
```

- [ ] **Step 4: Update Run() to pass penalty parameter**

In `Run()`, update the `runLevel()` call:

```go
elapsed, err := m.runLevel(ctx, game, levelIdx, 0)
```

- [ ] **Step 5: Test**

Run: `go build ./...`

- [ ] **Step 6: Commit**

```bash
git add internal/games/sokoban/module.go
git commit -m "feat(sokoban): add pause menu with restart penalty"
```

---

## Task 6: Fix Final Table Borders

**Files:**
- Modify: `internal/games/sokoban/module.go`

- [ ] **Step 1: Fix showFinalSummary border alignment**

Replace the current `showFinalSummary()` with corrected version:

```go
func (m Module) showFinalSummary(times []time.Duration) error {
	m.Screen.Clear()

	var b strings.Builder
	b.WriteString("╔═══════════════════════════════════╗\n")
	b.WriteString("║      🎉 FIN DEL JUEGO! 🎉          ║\n")
	b.WriteString("╠═══════════════════════════════════╣\n")

	for i, t := range times {
		line := fmt.Sprintf("║  Nivel %2d:    %s", i+1, fmtDuration(t))
		padding := 37 - len(line)
		line += strings.Repeat(" ", padding) + "║\n"
		b.WriteString(line)
	}

	b.WriteString("╠═══════════════════════════════════╣\n")

	var total time.Duration
	for _, t := range times {
		total += t
	}
	totalLine := "║  TOTAL:      " + fmtDuration(total)
	padding := 37 - len(totalLine)
	totalLine += strings.Repeat(" ", padding) + "║\n"
	b.WriteString(totalLine)

	b.WriteString("╚═══════════════════════════════════╝")

	// Center the box on screen
	sw, sh := m.Screen.Size()
	lines := strings.Split(b.String(), "\n")
	boxH := len(lines)
	boxW := 41
	boxY := (sh - boxH) / 2
	boxX := (sw - boxW) / 2

	for i, line := range lines {
		render.WriteText(m.Screen, boxX, boxY+i, line)
	}

	if err := m.Screen.Present(); err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	return nil
}
```

- [ ] **Step 2: Test**

Run: `go build ./...`

- [ ] **Step 3: Commit**

```bash
git add internal/games/sokoban/module.go
git commit -m "fix(sokoban): align final table borders"
```

---

## Completion Checklist

- [ ] Task 1: ActionHelp input works
- [ ] Task 2: UndoAvailable method added
- [ ] Task 3: Undo indicator shows in header
- [ ] Task 4: Help menu displays correctly
- [ ] Task 5: Pause menu works with restart penalty
- [ ] Task 6: Final table borders align properly
- [ ] All tests pass
- [ ] Manual playtest confirms all features