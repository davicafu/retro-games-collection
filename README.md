# Regro Collection Games

A progressive collection of terminal games built in Go as a technical laboratory for architecture, algorithms, game loops, input handling, collisions, reusable runtime patterns, and deterministic simulations.

The goal is not simply to build multiple games.

The goal is to evolve a reusable gameplay runtime through increasingly demanding game designs.

---

# Vision

This repository explores how far terminal-based games can go while maintaining:

- clean architecture
- deterministic gameplay systems
- reusable engine capabilities
- strong separation of concerns
- high-quality terminal UX

---

# Tech Stack

- **Language:** Go
- **CLI Framework:** Cobra
- **Launcher / Shell UI:** Bubble Tea + Lip Gloss
- **Gameplay Backend:** Abstracted runtime (with optional `tcell` integration for lower-level terminal control)

---

# Architectural Principles

## Cobra is only the CLI surface

Cobra is responsible for:

- command parsing
- configuration entry
- dispatching into application flows

Cobra must never contain gameplay logic.

---

## Bubble Tea is not the game engine

Bubble Tea is used for:

- menus
- launcher flows
- overlays
- shell UX

Actual gameplay runtime and loops must remain framework-independent.

---

# Roadmap

| Order | Game                               | Main Learning Goals                            |
| ----: | ---------------------------------- | ---------------------------------------------- |
|     1 | Snake                              | Grid systems, collisions, scoring              |
|     2 | Sokoban                            | Turn-based logic, levels, deterministic rules  |
|     3 | Endless Runner                     | Scrolling systems, speed scaling               |
|     4 | Flappy-inspired                    | Gravity, impulse physics, obstacle generation  |
|     5 | Breakout / Arkanoid                | Bounce physics, angles, power-ups              |
|     6 | Space Invaders                     | Shooting systems, enemy waves                  |
|     7 | Pac-Man-inspired                   | Maze navigation, simple enemy AI               |
|     8 | Metro Survival                     | Lane systems, rapid decision gameplay          |
|     9 | Asteroids                          | Inertia, rotation, screen wrapping             |
|    10 | Donkey Kong / Lode Runner-inspired | Platforms, ladders, gravity                    |
|    11 | Tetris                             | Spatial manipulation, rotations, line clearing |
|    12 | Hack the System                    | Hybrid puzzle/arcade/narrative systems         |

---

# Hack the System

The final project is envisioned as a retro-futuristic fake hacking terminal game.

It is not real hacking.

Instead, it combines arcade mechanics, puzzles, pressure systems, and narrative presentation inspired by 80s hacker movies.

Possible gameplay modules:

- **Crack Password**  
  Identify patterns before time runs out.

- **Bypass Firewall**  
  Navigate a cursor through hostile network paths.

- **Decrypt Data**  
  Logic-combination puzzles inspired by Mastermind.

- **Packet Runner**  
  Deliver packets between nodes while avoiding antivirus systems.

- **Trace Escape**  
  Every action increases detection level; complete the objective before getting traced.

- **Boss System**  
  Enemy AI dynamically changes board rules and constraints.

---

# How to Run

## Requirements

- Go 1.24+
- A modern terminal with ANSI support

---

## Clone the repository

```bash
git clone https://github.com/your-user/retro-collection-games.git

cd retro-collection-games
```

---

## Install dependencies

```bash
go mod tidy
```

---

## Run the launcher

```bash
go run .
```

---

## Run a specific game

```bash
go run . play snake
go run . play sokoban
go run . play runner
go run . play drone
go run . play breach
go run . play botnet
go run . play leak
go run . play packet
go run . play debris
go run . play rack
```

Game names are case-insensitive. An unknown name opens the launcher menu.

---
