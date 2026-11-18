# Task 5 Adjustment — Interactive Snake in Real Terminal

## Context

The original Task 5 wording allowed a technically wired but non-interactive execution path for `play snake`, using headless adapters and scripted input. That is no longer acceptable as the milestone definition for this project.

The project goal for this stage is to validate the first **real playable terminal game loop**, not just runtime connectivity.

## Updated Goal

Task 5 is redefined from:

> Wire Snake through runtime and CLI

to:

> Deliver the first real interactive Snake loop in a terminal backend adapter, while preserving runtime decoupling from UI frameworks.

## Product Meaning of “play snake”

For this milestone, `play snake` must mean:

- a real terminal session opens
- the game renders visibly in that terminal
- keyboard input is live
- the snake can be controlled by the player
- the session exits cleanly

It must not mean:

- an invisible headless harness
- scripted input pretending to be gameplay
- a test-only execution path exposed as if it were product behavior

## Preferred Technical Direction

The selected approach is:

- a **minimal adapter-owned runtime path** for real gameplay
- **tcell** as the first real gameplay backend
- a **custom adapter boundary** above tcell so gameplay and runtime do not depend directly on it

Bubble Tea remains the launcher/shell technology. It must not become the gameplay engine.

## Scope Included in Adjusted Task 5

### Included

- real terminal screen adapter backed by tcell
- real input adapter backed by tcell
- visible Snake rendering in terminal
- live keyboard control for Snake
- clean startup and shutdown path
- `app -> runtime -> snake -> terminal adapter` integration

### Explicitly Excluded

- advanced HUD or overlays
- audio
- in-game menu system
- pause system beyond simple exit if needed
- polished resizing behavior
- generalized backend support for all future games
- final production-grade engine backend abstraction for every scenario

This task should build the **minimum honest gameplay backend** required to validate Snake as a playable artifact.

## Architectural Boundaries

### App Layer

Responsibilities:

- choose the game to run
- assemble the appropriate runtime dependencies
- invoke the Snake module with a real terminal adapter path

The app layer must no longer route `play snake` through headless adapters.

### Runtime Layer

Responsibilities:

- lifecycle
- tick/update/render orchestration
- input consumption via abstraction
- output via abstract screen contract

The runtime should keep depending on contracts, not on tcell directly.

### Platform Terminal Layer

This is the new real gameplay boundary introduced by the adjusted task.

Responsibilities:

- tcell screen lifecycle
- keyboard event translation into runtime input actions
- frame presentation through the existing screen abstraction or a minimally refined equivalent

### Snake Module

Snake must continue to express domain and render logic in gameplay terms, not in terminal-framework terms.

Snake should know nothing about tcell.

## Acceptance Criteria

Task 5 is done only if all of the following are true:

### Functional

- `play snake` opens a visible terminal game session
- the snake responds to real keyboard input
- the frame output is visible in terminal
- quitting or game termination exits cleanly

### Architectural

- headless adapters are used only for tests
- runtime remains decoupled from Bubble Tea
- Snake remains decoupled from tcell
- the real gameplay path goes through a platform adapter layer

### Quality

- existing Tasks 1–4 remain green
- Snake domain tests remain intact
- new adapter/runtime behavior is covered by focused tests where practical
- limitations are documented explicitly instead of hidden

## Minimal Success Standard

The goal is not a polished final backend.

The goal is to prove, honestly and visibly, that the project can:

1. run a real game loop in a terminal
2. read live input
3. render visible state
4. preserve architectural boundaries while doing so

## Consequence for Planning

The previous Task 5 plan should be replaced or amended so it no longer permits a headless-only production path.

The next implementation plan revision should:

- introduce a real terminal adapter task
- define the minimum tcell-backed screen/input responsibilities
- update Task 5 success criteria to visible/interactable gameplay
- keep the scope tightly limited to Snake
