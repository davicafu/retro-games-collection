# Retro Games Collection

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

The final game is a retro-futuristic fake hacking terminal, inspired by 80s
hacker movies. It is not real hacking: it mixes a narrative puzzle, arcade
mini-games and a pressure system. The game text is in Spanish; commands and
common IT terms stay in English.

You are a ghost at a terminal. Each level is a network: start at its
gateway, read the files, work out passwords and keys from the clues, and
reach the secret the client asked for. Five levels, each a company, each
one tying the story closer to the next.

## Commands

| Command                    | What it does                                   |
| -------------------------- | ---------------------------------------------- |
| `help`                     | List the commands                              |
| `ls`, `cat <file>`         | List and read the machine's files              |
| `scan`                     | Machines this one reaches, and their state     |
| `connect <machine>`        | Go to a machine; a protected one asks for a breach, a password or a key |
| `decrypt <file> <key>`     | Decrypt a file                                 |
| `map`                      | Network map: arrows to pick, enter to connect, backspace to go back |
| `hint`                     | A vague hint, then an explicit one; costs points |
| `clear`                    | Clear the screen                               |

`↑↓` walk the command history. `ctrl+c` quits, since `q` is a letter here.

## Breaches

A protected machine opens with a short arcade breach:

- **ICE**: reach the port `[]` through a maze before a sentry catches you.
- **SYNC**: lock a moving signal inside a window, several times, with few misses.
- **GATES**: set the input switches so every logic gate outputs what the core asks for.

Every breach has a time limit, and each level makes them harder.

## Trace and score

Every mistake raises the trace: an unknown command, a wrong password, a lost
breach. A honeypot raises it hard. At 100% the run is over. Each level
starts at 1000 points and loses 100 per hint and 5 per point of trace; the
final score adds up the levels.

## Levels

1. **Meridian Grid**: the tutorial. Follow the clues to a password and a key.
2. **Helix Biomedical**: a firewall guards a machine until it is yours, and the final file is encrypted.
3. **Kestrel Analytics**: a password split across machines, a honeypot that looks open, and a key to recover before the core.
4. **Cormoran Shipping**: an IDS watches some machines (`!` on the map, "vigilada" in `scan`). On them the trace rises on its own until you own the IDS machine.
5. **Halcyon Group**: the company behind the others. Watched machines again, and the key machine locks after two wrong passwords; then only an ICE breach opens it.

Clearing a level saves the run (see [High scores and saved runs](#high-scores-and-saved-runs)).

---

# How to Run

## Requirements

- Go 1.24+
- A modern terminal with ANSI support

---

## Clone the repository

```bash
git clone https://github.com/davicafu/retro-games-collection.git

cd retro-games-collection
```

---

## Install dependencies

```bash
go mod tidy
```

---

## Quickstart

Build once, then run the binary. It opens the launcher at once; `go run .`
also works, but compiles the whole project before every start.

```bash
go build -o retro-games .
./retro-games
```

Rebuild after changing the code. Example output of `./retro-games --help`:

```text
Progressive terminal games laboratory

Usage:
  retro-games [flags]
  retro-games [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  play        Run a game directly

Flags:
  -h, --help   help for retro-games

Use "retro-games [command] --help" for more information about a command.
```

---

## Run a specific game

```bash
./retro-games play snake
```

Or without building first:

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
go run . play defrag
go run . play hack
```

Game names are case-insensitive. An unknown name opens the launcher menu.

---

## High scores and saved runs

Every game keeps its five best results. When a finished game enters its
table you type a name (up to 12 characters; empty is saved as `ANON`).
Sokoban, which has no score, ranks finished runs by levels completed and
then by less total time. In the launcher, press `r` to read the tables,
`←→` to switch game and `esc` to go back.

Hack saves your run each time you clear a level. Choosing Hack with a saved
run asks whether to continue it or start a new game; a new game, a win or a
run ended by the trace deletes the save.

Both live in the user's config directory, under `retro-games`
(`records.json` and `saves.json`; on macOS
`~/Library/Application Support/retro-games`). An unreadable file is renamed
to `<name>.bad` and a fresh one starts; if the directory cannot be used,
records last only for the session.

---
