# Snake Playfield Chrome and Game Over Overlay Adjustment

## Context

Snake is now much closer to a real playable loop, but its gameplay presentation still feels too raw for a user-facing terminal game.

Two problems are currently visible:

1. The playfield edges are not visually obvious, so collision with the boundary feels abrupt
2. Losing the game exits too abruptly, without a user-facing recovery flow

This makes Snake feel like a technical prototype rather than a polished milestone.

## Updated Goal

Snake should gain:

- visible playfield boundaries
- an always-visible score display
- a reusable game-over overlay with action selection

The implementation should improve Snake immediately while also establishing a reusable pattern for future games.

## Product Behavior

### During gameplay

- the playfield boundary is visible at all times
- the score is visible at all times
- collision with the border still kills the snake, but now the player sees the limit clearly before impact

### On game over

the game should not terminate abruptly. Instead, it should display an overlay with:

- title: `Game Over`
- final score
- selectable actions:
  - `Play Again`
  - `Back to Menu`
  - `Quit`

## Design Direction

The selected approach is a **hybrid reusable overlay**:

- implement a simple overlay now
- design it so the option-selection and rendering pattern can be reused later
- keep Snake-specific messaging and action mapping outside the reusable overlay core

This avoids both duplication and premature infrastructure.

## Architectural Boundaries

### Snake Domain

The domain remains responsible for:

- movement
- collisions
- score
- death state

The domain must not know about terminal overlay rendering.

### Playfield Chrome / HUD Presentation

The visible border and score belong to presentation, not game rules.

They should render around the playfield using existing runtime/screen abstractions.

### Overlay Component

The reusable part should be responsible for:

- displaying a title/message block
- displaying a list of options
- tracking currently selected option
- returning a selected intention

It must not know what Snake is or what a menu choice means semantically.

### Action Handling

The selected overlay option should resolve to an intention such as:

- restart
- back_to_menu
- quit

The game/app orchestration layer is responsible for reacting to that intention.

## Scope Included

- visible playfield border for Snake
- visible score display for Snake
- game-over overlay with keyboard selection
- reusable option-selection pattern suitable for future games
- actions:
  - play again
  - back to menu
  - quit

## Scope Excluded

- generalized pause system
- advanced victory overlays
- animated transitions
- theme/skin system
- complex HUD layout engine
- reusable overlay framework for every future screen state

This is a focused capability slice, not a complete terminal UI framework.

## Acceptance Criteria

This adjustment is done only if all of the following are true:

### Functional

- the Snake playfield border is visible during gameplay
- the score is visible during gameplay
- losing the game shows an overlay instead of exiting immediately
- the overlay supports selecting:
  - play again
  - back to menu
  - quit

### Reusability

- overlay option rendering/selection logic is not hard-coded as Snake-only logic
- Snake-specific content and actions are cleanly separated from the reusable selection mechanism

### Product

- the game feels understandable at the edge of the playfield
- death no longer feels abrupt or broken
- the player has an immediate next-step choice after losing

## Consequence for Planning

The next implementation plan revision should include:

- playfield chrome rendering
- score rendering
- reusable option-selection overlay component
- Snake game-over integration
- restart / menu / quit action handling
