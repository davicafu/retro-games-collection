# CLI Games Collection — Design Spec

## Overview

CLI Games Collection is a progressive terminal game laboratory built in Go. Its goal is not just to ship a set of games, but to evolve a reusable runtime by validating engine capabilities through increasingly complex games.

The project will prioritize:

1. High-quality terminal UX from the beginning
2. A reusable runtime architecture with explicit boundaries
3. Deliberate, slower iteration over fast delivery

The project will use:

- Cobra for CLI command structure
- Bubble Tea + Lip Gloss for launcher and shell UX
- A decoupled runtime for game execution
- Optional tcell-backed terminal runtime where lower-level control is needed

## Product Vision

The system is an engine-lab expressed as a collection of games. Each game exists for two reasons:

1. It is a playable artifact with a polished terminal experience
2. It introduces, validates, or stress-tests one or more reusable engine capabilities

This means the roadmap is organized both by games and by runtime capabilities. The games are the evolutionary path; the engine capabilities are the architectural spine.

## Architectural Direction

### Entry Surfaces

The application will expose two primary entry surfaces:

- An interactive launcher for user-facing navigation
- CLI subcommands for direct execution, tooling, testing, debugging, and benchmarking

This dual-entry approach supports both polished UX and serious technical workflows.

### Runtime Boundary

The runtime must be independent from Cobra and as independent as practical from Bubble Tea.

The runtime owns:

- game lifecycle
- update/render loop
- timing and tick orchestration
- input abstraction
- scene/state transitions
- reusable collision and movement systems
- debug hooks
- testability via deterministic simulation

The shell layer owns:

- command parsing
- launcher UX
- settings menus
- help/about/roadmap screens
- orchestration into game execution

### Backend Boundary

Gameplay logic must not depend directly on Bubble Tea internals, ANSI output details, or raw terminal APIs. Terminal backends must be replaceable through platform adapters.

The preferred model is:

- Bubble Tea + Lip Gloss for app shell and launcher flows
- A backend abstraction for active gameplay
- tcell as an implementation option for lower-level runtime control

## Design Principles

1. Reusable systems live by capability, not by game
2. Domain rules must not depend on CLI or terminal framework details
3. Visual shell quality matters from day one
4. New abstractions must be justified by concrete gameplay pressure
5. Deterministic testing is a first-class design constraint

## Suggested Project Structure

```text
cmd/
internal/
  app/
  launcher/
  engine/
    runtime/
    scene/
    input/
    render/
    grid/
    physics/
    debug/
  platform/
    terminal/
  games/
    snake/
    sokoban/
    runner/
    flappy/
    breakout/
    spaceinvaders/
    pacman/
    metrosurvival/
    asteroids/
    platformer/
    tetris/
    hackthesystem/
docs/
```

The exact package names can evolve, but the architectural separation should remain stable.

## Capability Roadmap

### Layer 0 — App Shell

- Cobra command tree
- Bubble Tea launcher
- global settings
- game selection
- help and roadmap screens

### Layer 1 — Core Runtime

- init/update/render/teardown lifecycle
- configurable tick rate
- input provider abstraction
- frame writer/screen buffer abstraction
- pause/resume/quit
- scene switching
- injectable clock

### Layer 2 — Grid Engine

- discrete coordinates
- occupancy model
- tile-based collision
- step/turn movement
- level parsing basics

### Layer 3 — Scroll and Endless Motion

- advancing world state
- obstacle spawning
- progressive difficulty
- lane-based or constrained motion systems

### Layer 4 — Arcade Physics

- velocity
- acceleration
- gravity
- axis-aligned collision handling
- bounce/reflection rules

### Layer 5 — Multi-Actor Combat Loop

- projectiles
- multiple enemies
- wave progression
- cooldowns
- game-over and victory rules

### Layer 6 — Navigation and Light AI

- maze modeling
- pickups
- ghost/enemy movement policies
- simple chase/scatter behavior

### Layer 7 — Inertial Movement

- rotation
- thrust
- vector-based motion
- screen wrapping
- directional firing

### Layer 8 — Platform Structure

- platforms
- ladders
- gravity and falls
- floor and wall checks
- multi-level stage logic

### Layer 9 — Complex Spatial Manipulation

- tetromino/piece systems
- rotation rules
- line clear detection
- difficulty acceleration
- placement timing rules

### Layer 10 — Hybrid Meta-Systems

- multi-mode rulesets
- scripted phase transitions
- boss systems that alter game constraints
- mixed mechanic orchestration

## Game Roadmap by Era

### Era 1 — Grid Foundations

1. Snake
2. Sokoban

### Era 2 — Action Loops

3. Endless runner inspired by the dinosaur runner
4. Flappy Bird inspired game
5. Breakout / Arkanoid inspired game
6. Space Invaders inspired game

### Era 3 — Spatial Intelligence

7. Pac-Man inspired game
8. Metro Survival
9. Asteroids

### Era 4 — Complex Systems

10. Donkey Kong / Lode Runner inspired platformer
11. Tetris
12. Hack the System

## Game-to-Capability Mapping

### Snake

- core loop
- grid movement
- input handling
- discrete collision
- score tracking

### Sokoban

- rule-based pushing
- turn logic
- level representation
- puzzle-state validation

### Endless Runner

- scrolling world
- difficulty ramping
- obstacle generation
- timing pressure

### Flappy-Style Game

- gravity and impulse
- obstacle gaps
- forgiving but meaningful collision tuning

### Breakout

- reflection angles
- brick durability/destruction
- power-up support

### Space Invaders

- coordinated enemy waves
- projectile management
- simultaneous actor updates

### Pac-Man Inspired

- maze navigation
- pickups
- light enemy AI
- route constraints

### Metro Survival

- original rules built on rapid decision loops
- reusable obstacle and lane mechanics

### Asteroids

- inertial motion
- rotation and directional fire
- wraparound space

### Platformer Inspired by Donkey Kong / Lode Runner

- platform and ladder traversal
- gravity and enemy patrols
- stage-based level design

### Tetris

- piece representation
- rotation and collision rules
- line clears and speed progression

### Hack the System

Hack the System is the capstone. It should not be treated as an early feature spike. It exists to prove that the runtime can compose multiple mechanics and presentation modes into an original identity.

Potential phase types include:

- password cracking pattern game
- firewall bypass navigation
- decryption logic puzzle
- packet delivery obstacle run
- trace-escape pressure loop
- rule-changing boss encounter

It should reuse established systems where possible rather than inventing isolated one-off implementations.

## Initial Milestone Recommendation

The project should begin with a Foundation milestone and then immediately validate it with Snake.

### Milestone 0 — Foundation

- initialize repository structure
- establish Cobra root command
- build launcher skeleton with Bubble Tea
- define runtime contracts
- define terminal/input/render abstractions
- provide injectable clock and RNG
- provide headless simulation mode
- establish test harness shape

### Milestone 1 — Snake

- playable snake loop
- grid state and food spawning
- scoring
- collision handling
- deterministic simulation tests
- terminal rendering integration

Snake is the right first game because it validates most foundational runtime decisions while staying inside a constrained grid domain.

## Testing Strategy

### Primary Testing Targets

- pure gameplay logic
- collisions and movement rules
- spawn and progression logic
- level and board validation
- AI policies where applicable
- engine lifecycle contracts

### Deterministic Simulation Requirements

The architecture must support:

- injectable RNG
- injectable clock
- scripted input streams
- observable game state
- headless execution mode

This allows tests to validate game behavior over a known number of ticks without depending on real-time terminal rendering.

### Test Pyramid

- many unit tests for rule logic
- focused simulation tests for games and capabilities
- a small number of renderer/backend integration tests
- limited snapshots only where stable and useful

### Explicit Non-Goals for Testing

- frame-by-frame visual end-to-end coverage for all games
- reliance on sleeps or wall-clock timing
- coupling assertions to ANSI-specific output details

## Risks and Mitigations

### Risk: Overengineering the engine too early

Mitigation: only formalize reusable capabilities when repeated or strongly justified by the next game in the roadmap.

### Risk: Shell framework leaking into gameplay design

Mitigation: keep runtime contracts independent from Bubble Tea and Cobra.

### Risk: Beautiful shell, weak runtime core

Mitigation: pair UI work with deterministic runtime contracts from the first milestone.

### Risk: Too many isolated game implementations

Mitigation: require each new game to state which capability it introduces or validates.

## Decisions Captured

- Language: Go
- CLI framework: Cobra
- Shell/launcher UI: Bubble Tea + Lip Gloss
- Runtime interaction model: decoupled from shell framework
- Lower-level terminal control: tcell when useful
- Product structure: launcher + CLI subcommands
- Planning model: roadmap by games and by engine capabilities
- Delivery philosophy: quality and architecture before speed

## Open Questions for Later Planning

- whether replay support is part of the early runtime or deferred
- whether level authoring tools are text-only or need editor workflows
- whether sound simulation/feedback will exist at all or remain visual-only
- how much package-level abstraction is needed before the fourth game

## Recommended Next Step

Write an implementation plan for:

1. repository bootstrap
2. foundation runtime contracts
3. launcher shell
4. Snake as the first validating game
