# In-Game Overlay Architecture Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce a shared in-game overlay capability and migrate Snake and Sokoban to use it with consistent structure and navigation.

**Architecture:** Add a reusable `internal/engine/ui/overlay` package that owns overlay state, navigation intents, and rendering. Keep all game-specific meanings in Snake and Sokoban by letting games declare option IDs and interpret overlay results themselves.

**Tech Stack:** Go, existing engine `render.Screen`, existing in-game modules for Snake and Sokoban, Go testing package.

---

## File map

| File | Responsibility |
|------|----------------|
| `internal/engine/ui/overlay/model.go` | Shared overlay model, section types, result types, navigation behavior |
| `internal/engine/ui/overlay/model_test.go` | Deterministic tests for selection, confirm, and dismiss behavior |
| `internal/engine/ui/overlay/render.go` | Shared in-game overlay renderer on top of `render.Screen` |
| `internal/engine/ui/overlay/render_test.go` | Headless rendering tests for frame, title, options, and info sections |
| `internal/games/snake/module.go` | Build shared overlays for pause and game-over flows; interpret returned option IDs |
| `internal/games/snake/render.go` | Replace custom overlay drawing with shared renderer |
| `internal/games/snake/module_test.go` | Verify Snake builds and resolves shared overlays correctly |
| `internal/games/snake/overlay.go` | Delete after migration |
| `internal/games/snake/overlay_test.go` | Delete after migration |
| `internal/games/sokoban/module.go` | Build shared overlays for pause, help, level-complete, and final summary flows |
| `internal/games/sokoban/module_test.go` | Add tests for overlay construction and result handling |
| `internal/games/sokoban/overlay.go` | Delete after migration |
| `internal/games/sokoban/overlay_draw.go` | Delete after migration |
| `internal/engine/input/input.go` | Remove overlay-specific semantic actions from global input |
| `internal/shell/model.go` | Keep mapping to generic gameplay actions only |

---

### Task 1: Create the shared overlay model

**Files:**
- Create: `internal/engine/ui/overlay/model.go`
- Create: `internal/engine/ui/overlay/model_test.go`

- [ ] **Step 1: Write the failing model tests**

Create `internal/engine/ui/overlay/model_test.go` with:

```go
package overlay

import "testing"

func TestActionOverlayMovesSelectionAndResolvesOptionID(t *testing.T) {
	model := NewAction(
		"PAUSED",
		[]Section{Text("Game paused")},
		[]Option{{ID: "resume", Label: "Continue"}, {ID: "restart-level", Label: "Restart"}},
	)

	model.Apply(IntentMoveDown)
	result := model.Apply(IntentConfirm)

	if result.Kind != ResultSelected {
		t.Fatalf("expected selected result, got %v", result.Kind)
	}
	if result.OptionID != "restart-level" {
		t.Fatalf("expected restart-level, got %q", result.OptionID)
	}
}

func TestInfoOverlayDismissesWithoutSelection(t *testing.T) {
	model := NewInfo("HELP", []Section{Text("Controls")}, DismissToPrevious)

	result := model.Apply(IntentClose)

	if result.Kind != ResultDismissed {
		t.Fatalf("expected dismissed result, got %v", result.Kind)
	}
	if result.Dismiss != DismissToPrevious {
		t.Fatalf("expected dismiss-to-previous, got %v", result.Dismiss)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/engine/ui/overlay`

Expected: FAIL with missing package symbols such as `NewAction`, `Section`, or `IntentConfirm`.

- [ ] **Step 3: Write the minimal shared model implementation**

Create `internal/engine/ui/overlay/model.go` with:

```go
package overlay

type Kind string

const (
	KindAction Kind = "action"
	KindInfo   Kind = "info"
)

type Intent string

const (
	IntentMoveUp   Intent = "move_up"
	IntentMoveDown Intent = "move_down"
	IntentConfirm  Intent = "confirm"
	IntentClose    Intent = "close"
)

type DismissPolicy string

const (
	DismissDefault  DismissPolicy = "default"
	DismissToGame   DismissPolicy = "to_game"
	DismissToMenu   DismissPolicy = "to_menu"
	DismissToPrevious DismissPolicy = "to_previous"
)

type ResultKind string

const (
	ResultNone      ResultKind = "none"
	ResultSelected  ResultKind = "selected"
	ResultDismissed ResultKind = "dismissed"
)

type Result struct {
	Kind     ResultKind
	OptionID string
	Dismiss  DismissPolicy
}

type Option struct {
	ID      string
	Label   string
	Enabled bool
}

type Section struct {
	Kind  string
	Lines []string
}

func Text(lines ...string) Section {
	return Section{Kind: "text", Lines: append([]string(nil), lines...)}
}

type Model struct {
	kind          Kind
	title         string
	sections      []Section
	options       []Option
	selectedIndex int
	dismissPolicy DismissPolicy
}

func NewAction(title string, sections []Section, options []Option) Model {
	cloned := append([]Option(nil), options...)
	for i := range cloned {
		if !cloned[i].Enabled {
			cloned[i].Enabled = true
		}
	}
	return Model{kind: KindAction, title: title, sections: append([]Section(nil), sections...), options: cloned, dismissPolicy: DismissDefault}
}

func NewInfo(title string, sections []Section, dismiss DismissPolicy) Model {
	return Model{kind: KindInfo, title: title, sections: append([]Section(nil), sections...), dismissPolicy: dismiss}
}

func (m *Model) Apply(intent Intent) Result {
	switch intent {
	case IntentMoveUp:
		if m.kind == KindAction && m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case IntentMoveDown:
		if m.kind == KindAction && m.selectedIndex < len(m.options)-1 {
			m.selectedIndex++
		}
	case IntentConfirm:
		if m.kind == KindInfo {
			return Result{Kind: ResultDismissed, Dismiss: m.dismissPolicy}
		}
		if len(m.options) > 0 {
			return Result{Kind: ResultSelected, OptionID: m.options[m.selectedIndex].ID}
		}
	case IntentClose:
		return Result{Kind: ResultDismissed, Dismiss: m.dismissPolicy}
	}
	return Result{Kind: ResultNone}
}

func (m Model) Title() string          { return m.title }
func (m Model) Kind() Kind             { return m.kind }
func (m Model) Sections() []Section    { return append([]Section(nil), m.sections...) }
func (m Model) Options() []Option      { return append([]Option(nil), m.options...) }
func (m Model) SelectedIndex() int     { return m.selectedIndex }
func (m Model) DismissPolicy() DismissPolicy { return m.dismissPolicy }
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/engine/ui/overlay`

Expected: PASS

- [ ] **Step 5: Commit the shared model**

```bash
git add internal/engine/ui/overlay/model.go internal/engine/ui/overlay/model_test.go
git commit -m "feat(ui): add shared overlay model"
```

---

### Task 2: Add the shared in-game renderer

**Files:**
- Create: `internal/engine/ui/overlay/render.go`
- Create: `internal/engine/ui/overlay/render_test.go`

- [ ] **Step 1: Write the failing renderer tests**

Create `internal/engine/ui/overlay/render_test.go` with:

```go
package overlay

import (
	"testing"

	enginerender "github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

func TestDrawActionOverlayShowsTitleAndSelection(t *testing.T) {
	screen := enginerender.NewBuffer(40, 20)
	model := NewAction("PAUSED", []Section{Text("Game paused")}, []Option{{ID: "resume", Label: "Continue"}, {ID: "menu", Label: "Menu"}})

	Draw(screen, model)

	if got := string(screen.Cell(14, 5).Rune); got != "P" {
		t.Fatalf("expected title start P, got %q", got)
	}
	if got := string(screen.Cell(8, 10).Rune); got != "▸" {
		t.Fatalf("expected selection marker, got %q", got)
	}
}
```

- [ ] **Step 2: Run the renderer tests to verify they fail**

Run: `go test ./internal/engine/ui/overlay -run TestDrawActionOverlayShowsTitleAndSelection`

Expected: FAIL with missing `Draw` symbol.

- [ ] **Step 3: Implement the shared renderer**

Create `internal/engine/ui/overlay/render.go` with:

```go
package overlay

import (
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

func Draw(screen render.Screen, model Model) {
	sw, sh := screen.Size()
	boxW := 36
	contentRows := 0
	for _, section := range model.Sections() {
		contentRows += len(section.Lines)
		if len(section.Lines) > 0 {
			contentRows++
		}
	}
	optionRows := len(model.Options())
	boxH := 4 + contentRows + optionRows
	boxX := (sw - boxW - 2) / 2
	boxY := (sh - boxH - 2) / 2

	render.FillRect(screen, boxX+1, boxY+1, boxW, boxH, ' ')
	render.DrawBox(screen, boxX, boxY, boxW, boxH)
	render.WriteText(screen, boxX+1+(boxW-len(model.Title()))/2, boxY+1, model.Title())

	row := boxY + 3
	for _, section := range model.Sections() {
		for _, line := range section.Lines {
			render.WriteText(screen, boxX+2, row, render.PadRight(line, boxW-2))
			row++
		}
		if len(section.Lines) > 0 {
			row++
		}
	}

	for i, option := range model.Options() {
		marker := "  "
		if i == model.SelectedIndex() {
			marker = "▸ "
		}
		render.WriteText(screen, boxX+2, row+i, render.PadRight(marker+option.Label, boxW-2))
	}
}
```

- [ ] **Step 4: Run the overlay package tests**

Run: `go test ./internal/engine/ui/overlay`

Expected: PASS

- [ ] **Step 5: Commit the renderer**

```bash
git add internal/engine/ui/overlay/render.go internal/engine/ui/overlay/render_test.go
git commit -m "feat(ui): add shared overlay renderer"
```

---

### Task 3: Migrate Snake to the shared overlay capability

**Files:**
- Modify: `internal/games/snake/module.go`
- Modify: `internal/games/snake/render.go`
- Modify: `internal/games/snake/module_test.go`
- Delete: `internal/games/snake/overlay.go`
- Delete: `internal/games/snake/overlay_test.go`

- [ ] **Step 1: Write a failing Snake overlay construction test**

Add to `internal/games/snake/module_test.go`:

```go
func TestPauseOverlayUsesSharedOverlayOptionIDs(t *testing.T) {
	g := NewGame(nil)
	g.showPauseOverlay()

	if got := g.overlay.Options()[0].ID; got != "resume" {
		t.Fatalf("expected resume option id, got %q", got)
	}
	if got := g.overlay.Options()[1].ID; got != "restart-run" {
		t.Fatalf("expected restart-run option id, got %q", got)
	}
}
```

- [ ] **Step 2: Run the Snake package tests to verify they fail**

Run: `go test ./internal/games/snake`

Expected: FAIL because `g.overlay` still uses the old local overlay type.

- [ ] **Step 3: Migrate Snake module and renderer to shared overlay**

Update `internal/games/snake/module.go` and `internal/games/snake/render.go` so they use the shared package:

```go
// module.go
import sharedoverlay "github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"

type Game struct {
	// ...
	overlay *sharedoverlay.Model
}

func (g *Game) showPauseOverlay() {
	m := sharedoverlay.NewAction("PAUSED", []sharedoverlay.Section{sharedoverlay.Text("Game paused")}, []sharedoverlay.Option{
		{ID: "resume", Label: "Continue", Enabled: true},
		{ID: "restart-run", Label: "Restart", Enabled: true},
		{ID: "back-to-menu", Label: "Menu", Enabled: true},
		{ID: "quit-game", Label: "Quit", Enabled: true},
	})
	g.overlay = &m
}

func (g *Game) handleOverlayConfirm() {
	if g.overlay == nil {
		return
	}
	result := g.overlay.Apply(sharedoverlay.IntentConfirm)
	switch result.OptionID {
	case "resume":
		g.overlay = nil
	case "restart-run":
		g.restart()
	case "back-to-menu":
		g.backToMenu = true
	case "quit-game":
		g.quit = true
	}
}

// render.go
import sharedoverlay "github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"

func DrawOverlay(screen render.Screen, overlay *sharedoverlay.Model) {
	sharedoverlay.Draw(screen, *overlay)
}
```

- [ ] **Step 4: Run the Snake tests**

Run: `go test ./internal/games/snake`

Expected: PASS

- [ ] **Step 5: Delete obsolete Snake overlay files and commit**

```bash
rm internal/games/snake/overlay.go internal/games/snake/overlay_test.go
git add internal/games/snake/module.go internal/games/snake/render.go internal/games/snake/module_test.go internal/games/snake/overlay.go internal/games/snake/overlay_test.go
git commit -m "refactor(snake): migrate overlays to shared ui package"
```

---

### Task 4: Migrate Sokoban to shared action and info overlays

**Files:**
- Modify: `internal/games/sokoban/module.go`
- Create: `internal/games/sokoban/module_test.go`
- Delete: `internal/games/sokoban/overlay.go`
- Delete: `internal/games/sokoban/overlay_draw.go`

- [ ] **Step 1: Write failing Sokoban overlay tests**

Create `internal/games/sokoban/module_test.go` with:

```go
package sokoban

import "testing"

func TestPauseOverlayOmitsNextLevelOnLastLevel(t *testing.T) {
	m := Module{}
	overlay := m.pauseOverlayForLevel(len(Levels) - 1)

	for _, option := range overlay.Options() {
		if option.ID == "next-level" {
			t.Fatalf("did not expect next-level on last level")
		}
	}
}

func TestHelpOverlayIsInformational(t *testing.T) {
	m := Module{}
	overlay := m.helpOverlay()

	if overlay.Kind() != "info" {
		t.Fatalf("expected info overlay, got %q", overlay.Kind())
	}
}
```

- [ ] **Step 2: Run Sokoban tests to verify they fail**

Run: `go test ./internal/games/sokoban`

Expected: FAIL because `pauseOverlayForLevel` and `helpOverlay` do not exist yet.

- [ ] **Step 3: Refactor Sokoban overlay construction to shared models**

Update `internal/games/sokoban/module.go` with helpers like:

```go
import sharedoverlay "github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"

func (m Module) pauseOverlayForLevel(levelIdx int) sharedoverlay.Model {
	options := []sharedoverlay.Option{
		{ID: "resume", Label: "Continuar", Enabled: true},
		{ID: "restart-level", Label: "Reiniciar (+10s)", Enabled: true},
	}
	if levelIdx < len(Levels)-1 {
		options = append(options, sharedoverlay.Option{ID: "next-level", Label: "Siguiente", Enabled: true})
	}
	options = append(options, sharedoverlay.Option{ID: "back-to-menu", Label: "Salir", Enabled: true})
	return sharedoverlay.NewAction("PAUSA", []sharedoverlay.Section{sharedoverlay.Text("Juego pausado")}, options)
}

func (m Module) helpOverlay() sharedoverlay.Model {
	return sharedoverlay.NewInfo(
		"AYUDA",
		[]sharedoverlay.Section{
			sharedoverlay.Text("Objetivo: llevar todas las cajas a los destinos."),
			sharedoverlay.Text("Controles: flechas/WASD mover, U deshacer, P pausa, ? ayuda."),
		},
		sharedoverlay.DismissToPrevious,
	)
}
```

- [ ] **Step 4: Route shared overlay results through Sokoban flow**

Still in `internal/games/sokoban/module.go`, update the pause/help/final-summary flow so it uses shared results:

```go
result := overlay.Apply(sharedoverlay.IntentConfirm)
switch result.OptionID {
case "resume":
	return pauseContinue
case "restart-level":
	return pauseRestart
case "next-level":
	return pauseNext
case "back-to-menu":
	return pauseQuit
}
```

Also update help/final summary drawing calls to use:

```go
sharedoverlay.Draw(m.Screen, overlay)
```

- [ ] **Step 5: Run the Sokoban tests and commit**

Run: `go test ./internal/games/sokoban`

Expected: PASS

```bash
rm internal/games/sokoban/overlay.go internal/games/sokoban/overlay_draw.go
git add internal/games/sokoban/module.go internal/games/sokoban/module_test.go internal/games/sokoban/overlay.go internal/games/sokoban/overlay_draw.go
git commit -m "refactor(sokoban): migrate overlays to shared ui package"
```

---

### Task 5: Remove overlay semantics from global input and run final verification

**Files:**
- Modify: `internal/engine/input/input.go`
- Modify: `internal/shell/model.go`

- [ ] **Step 1: Write a failing input test for the trimmed action set**

Add to `internal/engine/input/input_test.go`:

```go
func TestStaticProviderKeepsGameplayActionsOnly(t *testing.T) {
	provider := Static([]Action{ActionPause, ActionHelp, ActionUndo})
	frame := provider.Poll()

	if len(frame.Actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(frame.Actions))
	}
}
```

- [ ] **Step 2: Remove overlay-specific semantic actions**

Update `internal/engine/input/input.go` constants to:

```go
const (
	ActionUp     Action = "up"
	ActionDown   Action = "down"
	ActionLeft   Action = "left"
	ActionRight  Action = "right"
	ActionQuit   Action = "quit"
	ActionSelect Action = "select"
	ActionPause  Action = "pause"
	ActionUndo   Action = "undo"
	ActionHelp   Action = "help"
)
```

Then update `internal/shell/model.go` so in-game key handling only sends those generic gameplay actions and lets each game convert them to overlay intents locally.

- [ ] **Step 3: Run targeted verification**

Run: `go test ./internal/engine/input ./internal/engine/ui/overlay ./internal/games/snake ./internal/games/sokoban`

Expected: PASS

- [ ] **Step 4: Commit the cleanup**

```bash
git add internal/engine/input/input.go internal/engine/input/input_test.go internal/shell/model.go
git commit -m "refactor(input): remove overlay-specific global actions"
```

- [ ] **Step 5: Run repository verification for affected packages**

Run: `go test ./internal/engine/... ./internal/games/snake ./internal/games/sokoban`

Expected: PASS

---

## Spec coverage check

- Shared overlay capability: covered by Tasks 1-2
- Same visual structure: covered by Task 2
- Same base navigation: covered by Task 1 and integration in Tasks 3-4
- Action vs info overlays: covered by Tasks 1 and 4
- Launcher remains separate: enforced by file scope; launcher is not migrated into overlay package
- Remove domain logic from overlay system: covered by Task 1 contract and Tasks 3-4 result interpretation
- Reduce input leakage: covered by Task 5

## Final checklist

- [ ] Shared overlay package exists under `internal/engine/ui/overlay`
- [ ] Snake uses the shared overlay package
- [ ] Sokoban uses the shared overlay package
- [ ] Old per-game overlay implementations are removed
- [ ] Overlay-specific semantic actions are removed from global input
- [ ] Targeted Go tests pass for engine overlay, Snake, and Sokoban

## Next step

Execute this plan task-by-task, preferring subagent-driven execution so each migration step stays reviewable and easy to verify.
