# Debris (game 9)

Objective: ninth game of the roadmap, Asteroids-inspired: inertia, rotation and screen wrapping on the fixed-tick runtime.
Branch: `feat/debris` (from `feat/packet`) · Delivery: `ask-on-risk`, commits only when the user asks · Test runner: `go test ./...`

## Specs

Source requirements: README roadmap row `| 9 | Asteroids | Inertia, rotation, screen wrapping |` and the user's theme choice "Space debris cleanup (Recommended)" (L2), whose description was: "Eres un dron de mantenimiento en órbita que desintegra chatarra espacial: satélites muertos que se rompen en trozos cada vez más pequeños. Rompe un poco con la línea tech/hacker y aprovecha el espacio con wrap." Everything marked *(assumption)* is a proposed design default, not a user requirement.

- **S1 — Theme.** "Eres un dron de mantenimiento en órbita que desintegra chatarra espacial: satélites muertos que se rompen en trozos cada vez más pequeños." Original game: no names, ship, rock shapes, saucers or sounds from Asteroids.
- **S2 — Orbit (wrapping).** "aprovecha el espacio con wrap". Everything that leaves one edge of the playfield comes back at the opposite one, and distances and collisions measure across the edges too.
- **S3 — Drone motion.** *(assumption)* Left and right rotate the drone one of 32 headings per press; up fires its thruster for a short burst per press (terminals send no key release; repeats while holding join into continuous thrust). The drone keeps drifting with inertia, slowed only by a light drag, up to a top speed. Motion looks the same in every direction although terminal cells are twice as tall as wide.
- **S4 — Shots.** *(assumption)* Space fires a shot along the heading, adding the drone's velocity; at most 4 shots fly at once, each for a limited range, with a short cooldown between shots.
- **S5 — Debris.** "satélites muertos que se rompen en trozos cada vez más pequeños". *(assumption, how)* Three sizes: a hit large piece breaks into 2 medium ones, a medium into 2 small ones, a small one is destroyed. Pieces drift in straight lines; smaller ones are faster. Points: large 20, medium 50, small 100.
- **S6 — Lives.** *(assumption)* 3 lives. Touching debris costs a life: the drone respawns at the centre, at rest, and cannot be hit for 2 s (it blinks). The round ends when lives run out.
- **S7 — Waves.** *(assumption)* Wave n starts with 3+n large pieces (up to 8), placed away from the drone and moving a little faster each wave; clearing a wave starts the next after a short pause. The HUD shows wave, score and lives.
- **S8 — Viability.** User on Botnet (2026-11-11): "tienen que ser viables". *(assumption: applied to Debris too)* An autopilot that acts only every few ticks clears the first 3 waves with its 3 lives in at least 27 of 30 seeded games, pinned by a test.
- **S9 — Controls and flow.** *(assumption)* ←→/AD rotate, ↑/W thrust, space fire. Pause, Game Over (score and wave, Restart/Menu) and Q come from `internal/engine/session`. End title "HULL BREACHED".
- **S10 — Launcher.** *(assumption)* "Debris" appears in the launcher with its controls hint, and `play debris` starts it.
- **S11 — Engine.** Per AGENTS.md roadmap capability "inertial motion": integer vectors, 32 headings and a wrapping space live in a new `internal/engine/motion` package.
- **S12 — Determinism and timing.** Per AGENTS.md: injectable clock and RNG, scripted input, headless simulation tests; no sleep-based tests; integer physics only. 60 ticks per second, with real-time speeds pinned by a test.

## Tasks

| ID | Specs | Route | Status | Evidence |
| --- | --- | --- | --- | --- |
| T1 | S2, S11 | inline | done (uncommitted) | RED: `Vec` undefined → GREEN `engine/motion/motion_test.go` |
| T2 | S3, S4, S5, S6, S7, S12 | inline | done (uncommitted) | RED: `Piece` undefined → GREEN `game_test.go`; mutants killed: drag, top speed, shield, cooldown, shots in flight, wave pause, safe spawn distance, split spread, start without shield, piece cap, shot wrap (survived first: hits measure across edges anyway, so a test pins wrapped shot positions) |
| T3 | S8 | inline | done (uncommitted) | 29 of 30 games clear 3 waves (L4) |
| T4 | S1, rendering, HUD | inline | done (uncommitted) | `render_test.go` (written after the renderer, no RED observed); 80x24 preview checked |
| T5 | S9 | inline | done (uncommitted) | `module_test.go`: keys, end title, summary, run and quit |
| T6 | S10 | inline | done (uncommitted) | RED: 9 games, `play debris`, controls hint → GREEN; README run command and features index row |

Routing: each unit builds on the previous one; no independent parallel units, so no writer delegation. Forecast: about 1100 authored changed lines; commits wait for the user's explicit request (user rule).

## Log

- **L1** (2026-11-13, user): "si" (to: "¿Sigo con el juego 9 (inspirado en Asteroids) mientras pruebas Packet?")
- **L2** (2026-11-13, user answer): theme = "Space debris cleanup (Recommended)".
- **L3**: Design defaults (all S#-marked assumptions): physics in integer units where a cell is 1000 wide and 2000 tall, so motion is isotropic on screen; 16 headings from a fixed integer table (no floating point); 3 lives, 2 s of invulnerability after a respawn; end title "HULL BREACHED".
- **L4**: Viability calibration, measured over 30 seeds (games clearing 3 waves): first version 14. The autopilot shooting threats when that needed fewer turns than escaping: 19 (deaths were nearly all small pieces it saw coming). 32 headings instead of 16, since a heading step of 22.5 degrees cannot aim at a small piece more than a few cells away, for a person either: 22. Closing in on pieces out of range, slowly (thrusting at full speed made it crash): 23. Small pieces at 10.2 cells/s instead of 12: 29. Drone 18 cells/s, shots 36 cells/s for 0.75 s, medium pieces 8.4 and large 5.4 cells/s at wave 1, 8% faster each wave up to 60%.
- **L5**: Checks: `go vet ./...` clean, `go test -race ./...` fully green. Not committed: commits wait for the user's request. Not play-tested in a real terminal.
- **L6** (2026-11-13, user): "en debris hay unos que atravieso, no se si estan bien, los que son grande, ademas creo que no funcionan bien los controles la anve apenas se distingue, los objetivos deberia ser algo rojo, para que se distinguiera bie, y si lo otro no se te hace daño asi si puede distinguir bien". Cause of passing through large pieces: they were drawn from the cell holding their centre, up to a whole row above where hits measure them; shapes are now centred on the true position (`firstCell`, `TestDrawPiecesWhereTheyAreHit`). Debris is red, the drone cyan and drawn as a solid triangle (`▲◥▶◢▼◣◀◤`), shots yellow. Controls: turning uses the engine's key-repeat glide (a tap turns exactly one heading, holding turns steadily, a full turn in about 1.3 s); a thrust press lasts 0.4 s instead of 0.17 s, so thrust keeps going while the other hand taps a turn (a terminal stops repeating a held key when another is pressed), with a softer acceleration. Viability 28 of 30 (a larger drone hitbox was measured, 26-27, and left out).
- **L7** (2026-11-13, user, on the launcher names): "los nombres si me valen asi" (to the format "Leak · Pac-Man-like"). The launcher shows each game's inspiration next to its name: Runner · Chrome Dino-like, Drone · Flappy Bird-like, Breach · Breakout-like, Botnet · Space Invaders-like, Leak · Pac-Man-like, Packet · Subway Surfers-like, Debris · Asteroids-like; Snake and Sokoban already name their genre. `play <name>` keeps the short names.
