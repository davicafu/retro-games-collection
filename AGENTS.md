# AGENTS.md

## Project Purpose

This project is a progressive collection of terminal games built as a technical laboratory for architecture, algorithms, game loops, input handling, collisions, and reusable engine patterns in Go.

The goal is not just to build several games. The goal is to evolve a reusable runtime through increasingly demanding game designs.

## Product Direction

- Prioritize terminal UI quality from day one
- Be consistent with a reusable runtime architecture
- Move deliberately; delivery speed is less important than doing things well
- Treat each game as both a playable artifact and a validation point for engine capabilities

## Core Stack

- **Language:** Go
- **CLI:** Cobra
- **Shell UI / launcher:** Bubble Tea + Lip Gloss
- **Gameplay terminal backend:** abstracted; may use tcell when lower-level control is needed

## Architectural Rules

1. **Cobra is only for the CLI surface**
   - command parsing
   - configuration entry
   - dispatch into app/use cases
   - never game logic

2. **Bubble Tea is not the game engine**
   - use it for launcher flows, menus, overlays, shell UX
   - keep the gameplay runtime outside of Bubble Tea control

3. **Reusable code lives by capability, not by game**
   - put shared systems in engine/runtime packages
   - do not create ad-hoc cross-game helpers unless they represent a real capability

4. **Gameplay logic must not depend directly on terminal framework details**
   - isolate screen, input, timing, and backend behavior behind adapters/contracts

5. **Design for deterministic testing**
   - injectable clock
   - injectable RNG
   - scripted input support
   - headless simulation capability

## Roadmap Structure

Plan work in two dimensions:

1. **By game**
   - Snake
   - Sokoban
   - Runner
   - Flappy-inspired
   - Breakout
   - Space Invaders
   - Pac-Man-inspired
   - Metro Survival
   - Asteroids
   - Platformer inspired by Donkey Kong / Lode Runner
   - Tetris
   - Hack the System

2. **By engine capability**
   - app shell
   - core runtime
   - grid systems
   - scrolling systems
   - arcade physics
   - multi-actor combat loops
   - navigation and light AI
   - inertial motion
   - platform structure
   - complex spatial manipulation
   - hybrid meta-systems

## Implementation Heuristics

- Prefer small, explicit packages with one clear responsibility
- Avoid speculative abstraction unless the next game clearly justifies it
- Promote repeated patterns into engine capabilities once they appear more than once or are clearly foundational
- Keep domain models expressed in gameplay terms, not framework terms

## Testing Expectations

Favor:
- unit tests for rules and state transitions
- deterministic simulation tests
- focused contract tests for runtime behavior
- minimal, stable snapshots where useful

Avoid:
- sleep-based timing tests
- heavy ANSI-coupled assertions
- broad visual E2E coverage for every frame

## First Milestones

1. Foundation
   - repository structure
   - Cobra root command
   - Bubble Tea launcher shell
   - runtime interfaces/contracts
   - headless simulation support

2. Snake
   - first fully playable validating game

## Feature Docs

- All specs, designs, plans, and evidence live in `docs/features/<YYYY-MM-DD>-<feature>/`; see `docs/features/README.md`
- This overrides any workflow default location (`odd/tasks/`, `openspec/`, `docs/superpowers/`)

## Working Style

- Go slowly if needed, but make decisions intentionally
- Explain architectural tradeoffs clearly
- Prefer strong foundations over quick hacks
- When proposing new structure, map it to the next concrete game need
