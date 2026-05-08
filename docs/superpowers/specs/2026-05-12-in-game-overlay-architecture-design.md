# In-Game Overlay Architecture Design

This design defines a reusable in-game overlay system for terminal games in this repository. It standardizes visual structure and base navigation across pause menus, help modals, level-complete screens, and final summary overlays without pushing game logic into the UI layer.

## Quick path

1. Add a shared in-game overlay package under the engine/UI layer.
2. Move Snake and Sokoban to declare overlay content through a common model.
3. Keep overlay rendering and navigation shared, while each game interprets selected option IDs itself.

## Outcome

The project gets a reusable in-game capability instead of per-game overlay implementations. Overlays will look and behave consistently across games, but each game will keep ownership of its own domain actions, state transitions, and scoring rules.

## Scope

This design applies only to **in-game overlays**.

It does **not** apply to the launcher shell. The launcher is a separate surface with different UX constraints and should remain special.

## Problem

The current project reuses low-level drawing primitives, but it does not yet reuse an overlay system.

Today:

- Snake owns its own overlay model and rendering path.
- Sokoban owns a separate overlay model and several modal implementations.
- Overlay navigation and overlay semantics are partly duplicated.
- Some UI flow concepts leaked into generic input handling.

This makes it harder to keep behavior consistent and increases the risk of each new game inventing its own modal system.

## Decision summary

| Topic | Decision |
|-------|----------|
| Reuse target | Reuse overlays only inside games |
| Visual consistency | All in-game overlays share the same base structure |
| Navigation consistency | All in-game overlays share the same base navigation model |
| Domain ownership | Games keep responsibility for interpreting overlay results |
| Overlay capability shape | Use one composable overlay model with sections and declarative options |
| Overlay categories | Support both action overlays and informational overlays |
| Launcher | Remains outside this abstraction |

## Architectural direction

The reusable capability should live in a shared in-game UI layer, conceptually under something like:

```text
internal/engine/ui/overlay/
```

That shared capability should provide:

- a reusable overlay view-model
- a reusable navigation model
- a reusable renderer for in-game screens
- reusable rules for selection, confirm, and close behavior

Games should provide:

- the overlay title
- the content sections to render
- the set of available options for that situation
- the interpretation of the selected option ID or dismiss result

This preserves the project rule that reusable systems live by capability, not by game.

## Responsibilities

### The shared overlay system is responsible for

- holding overlay metadata and content
- rendering a consistent frame and layout
- managing current selection for selectable overlays
- handling base navigation intents
- returning neutral UI results such as selected option IDs or dismiss events

### The shared overlay system is not responsible for

- restarting a level
- advancing to the next level
- resuming gameplay
- calculating time penalties or scores
- mutating the game domain directly
- deciding what a given option ID means inside a game

This boundary is mandatory. If the overlay system starts executing game actions directly, it stops being reusable infrastructure and becomes disguised game logic.

## Overlay model

The reusable overlay should be modeled explicitly as UI state, not as pre-rendered text.

### Core overlay fields

The shared model should include concepts equivalent to:

- `kind` — whether the overlay is actionable or informational
- `title`
- `sections[]`
- `options[]`
- `selectedIndex`
- `dismissPolicy`
- `footerHints`

This is a view-model, not a domain object.

## Content composition

The body of the overlay should be composed from reusable section types rather than hardcoded per-screen templates.

Suggested section categories:

- text section
- key/value section
- bullet or list section
- spacer section

This lets the same shared structure support:

- pause menus
- help modals
- level-complete overlays
- score summaries
- final run summaries

The important point is that the renderer gets structured content, not domain-specific screen types.

## Option model

Options should be declarative.

Each option should carry concepts equivalent to:

- `id`
- `label`
- `enabled` (optional)

The option ID is the important contract. The overlay system does not need to understand `next-level`, `restart-level`, `retry-run`, or `back-to-menu`. It only needs to surface whichever option ID the game declared.

This keeps the capability game-agnostic.

## Overlay kinds

### 1. Action overlay

Used for menus that expose selectable actions.

Examples:

- pause menu
- level complete menu
- game-over menu with retry/quit options

Behavior:

- supports selection
- supports confirm
- supports close/back according to dismiss policy

### 2. Info overlay

Used for informational modals without a selectable options list.

Examples:

- help screen
- instructional modal
- informational summary that only needs dismiss

Behavior:

- no selection state is required
- confirm or close simply dismisses the modal
- the game decides where the dismiss returns control

This distinction is important because not every overlay should pretend to be a menu.

## Navigation model

Navigation should be based on a small set of UI intents, not on domain actions.

Suggested shared intents:

- move up
- move down
- confirm
- close/back

The system should **not** expose game-specific actions such as:

- restart
- next level
- resume
- quit to menu

Those remain domain outcomes chosen by the game through option IDs.

## Dismiss behavior

Dismiss behavior should be explicit in the model, but still neutral.

Examples of intent-level semantics:

- dismiss to gameplay
- dismiss to prior flow
- dismiss default

The overlay layer may report that kind of dismiss intent, but it should not execute the transition itself. The game module remains responsible for turning the dismiss result into a real state transition.

## Interaction flow

The intended flow is:

1. The game decides that an overlay must be shown.
2. The game builds the shared overlay model.
3. While the overlay is active, gameplay routes relevant input into overlay navigation intents.
4. The overlay system updates selection or produces a UI result.
5. The game module interprets that result and performs the real domain action.

This preserves clean ownership.

## Rendering rules

The project should have a single reusable in-game overlay renderer that applies a common visual grammar.

That renderer should keep consistent:

- frame shape
- title placement
- section spacing
- option list placement
- selected item highlighting
- padding and alignment
- footer or hint placement

Games may vary:

- title text
- body content
- number of sections
- number of options
- option labels
- presence of stats or summary lines

Games should not vary the base interaction pattern or overall overlay structure.

## Integration with current games

### Snake

The shared overlay system should replace the current per-game overlay path for:

- pause overlay
- game-over or end summary overlays
- any future in-game help or informational modal

### Sokoban

The shared overlay system should replace the current mix of ad hoc modal implementations for:

- pause overlay
- help overlay
- level-complete overlay
- final summary overlay

This is a refactor toward capability reuse, not just a UI cleanup.

## Input boundary

The overlay system should not force more game-specific commands into the global engine input package.

Instead, the recommended approach is:

- the game enters an overlay mode
- raw key input is translated into overlay-local UI intents
- the overlay processes those intents
- the game exits overlay mode when a result is produced

This avoids turning the generic input layer into a bag of game/menu semantics.

## Testing strategy

This capability should be designed for deterministic tests.

### Unit tests for the overlay model

Verify:

- selection movement
- confirm returns the correct option ID
- informational overlays dismiss correctly
- disabled options behave correctly if supported
- dismiss behavior reports the expected neutral result

### Construction tests in games

Verify:

- Snake builds valid overlays for the intended moments
- Sokoban builds valid overlays for the intended moments
- each game exposes only the options that make sense in that context

### Renderer tests

Verify through headless screen assertions:

- frame placement
- title placement
- selection highlight placement
- section rendering order
- footer or hint rendering

Tests should avoid sleep-based timing and fragile ANSI assertions.

## Constraints

- The launcher remains separate.
- Overlay code must stay independent from game domain logic.
- Rendering must remain reusable and deterministic.
- The shared capability must be broad enough for current overlays, but not so abstract that it becomes speculative framework code.

## Tradeoffs

### Benefits

- consistent in-game UX
- less duplicated overlay code
- cleaner boundaries between UI and game logic
- easier extension to future games
- easier testing of overlay behavior independent from full gameplay loops

### Costs

- requires refactoring both Snake and Sokoban
- needs a carefully designed model to avoid underfitting or overfitting
- introduces an engine/UI capability that must stay intentionally small and concrete

## Non-goals

- unifying the launcher with in-game overlays
- creating a universal UI framework for the whole application
- encoding game-specific outcomes directly into the shared overlay package

## Adoption plan

1. Define the shared overlay model and navigation behavior.
2. Implement the shared in-game renderer.
3. Migrate Snake overlays to the shared system.
4. Migrate Sokoban overlays to the shared system.
5. Remove duplicated overlay code and game-specific rendering helpers that become obsolete.
6. Revisit generic input handling to remove overlay-specific semantic leakage.

## Review checklist

- [ ] The design applies only to in-game overlays, not the launcher.
- [ ] The shared system owns navigation and rendering, not game actions.
- [ ] Games provide option IDs and interpret results themselves.
- [ ] Informational overlays and actionable overlays are both supported.
- [ ] The design reduces duplication without introducing speculative framework complexity.

## Next step

After this design is approved, write an implementation plan that breaks the work into a shared overlay capability first, then migrations for Snake and Sokoban.
