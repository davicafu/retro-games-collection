# CLI Games Collection — Spec Summary

## One-paragraph summary

CLI Games Collection is a Go-based terminal game laboratory designed to evolve a reusable runtime through a progressive sequence of increasingly complex games. Cobra will own the CLI surface, Bubble Tea + Lip Gloss will own launcher and shell UX, and gameplay will run on a decoupled runtime with terminal backends abstracted behind adapters, optionally using tcell where lower-level control is needed.

## Core decisions

- Prioritize terminal UI quality from day one
- Build reusable engine capabilities deliberately, not accidentally
- Prefer architectural clarity over delivery speed
- Expose both a visual launcher and direct CLI subcommands
- Organize planning by both games and engine capabilities
- Design for deterministic testing from the start

## Architectural boundaries

- **Cobra** handles commands, config, and dispatch only
- **Bubble Tea** handles shell UX, not the gameplay engine
- **Gameplay runtime** owns loop, timing, input abstraction, scenes, collisions, and debug hooks
- **Terminal backends** remain replaceable through adapters
- **Reusable code** is grouped by capability, not by individual game

## Capability progression

The engine grows in layers:

1. app shell
2. core runtime
3. grid systems
4. scrolling systems
5. arcade physics
6. combat/multi-actor loops
7. navigation and simple AI
8. inertial motion
9. platform structures
10. complex spatial manipulation
11. hybrid meta-systems

## Game roadmap by era

### Era 1 — Grid Foundations
- Snake
- Sokoban

### Era 2 — Action Loops
- Runner
- Flappy-inspired
- Breakout
- Space Invaders

### Era 3 — Spatial Intelligence
- Pac-Man-inspired
- Metro Survival
- Asteroids

### Era 4 — Complex Systems
- Platformer inspired by Donkey Kong / Lode Runner
- Tetris
- Hack the System

## Why Snake goes first

Snake is the first validation game because it exercises the core loop, grid movement, input handling, collision basics, scoring, and rendering while keeping the problem space constrained.

## Role of Hack the System

Hack the System is the capstone. It should prove that the project can compose multiple reusable mechanics and presentation modes into one original experience rather than acting as an early experimental spike.

## Testing strategy

- Favor unit tests for rules and state transitions
- Add deterministic simulation tests with injected clock, RNG, and scripted input
- Keep renderer/backend integration tests focused and limited
- Avoid sleep-based timing tests and brittle ANSI-heavy assertions

## Immediate next milestone

- Foundation milestone: repository structure, Cobra root command, Bubble Tea launcher shell, runtime contracts, abstractions, and headless simulation support
- First validating game: Snake
