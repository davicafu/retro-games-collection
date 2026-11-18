# Breach (game 5, Breakout-inspired)

Objective: fifth game of the roadmap; introduces bounce physics with angles and power-ups. The player is the attacker: a virus bouncing off an injector to break through a firewall.
Branch: `feat/breach` (from `feat/drone`) · Delivery: `ask-on-risk` · Test runner: `go test ./...`

## Specs

Source requirements: README roadmap row `| 5 | Breakout / Arkanoid | Bounce physics, angles, power-ups |` and the user's theme choice (L1): "El Firewall aunque el nombre sería mejor otro estilo un virus intentando penetrsr, no siempre tenernos que ser los buenos". Everything marked *(assumption)* is a proposed design default, not a user requirement.

- **S1 — Theme.** The ball is a virus, the paddle an injector at the bottom, the bricks firewall blocks. Breaking every block breaches the firewall. Original game: no names, art or levels from Breakout or Arkanoid.
- **S2 — Physics.** *(assumption)* Integer fixed-point, 200 sub-cells per cell, 60 ticks per second (L8: "un poquito mas de fluidez estaria bien"; L10: "me gustaria mas a 60"; was 100 sub-cells and 50 ms). Real-time speeds are kept (diagonal 9 cells/s, drops 6 cells/s, Wide 15 s); Steep is rounded to (17,37). The ball moves one axis at a time and each velocity component stays below one cell per tick, so it never tunnels through a block.
- **S3 — Bounces.** *(assumption)* Walls and ceiling reflect the matching velocity component. A block hit reflects the component of the axis that hit it.
- **S4 — Angles.** *(assumption)* Where the virus lands on the injector sets the exit angle: outer, middle or inner thirds of each half give shallow, diagonal or steep angles, mirrored left and right. The exact centre keeps the current horizontal direction.
- **S5 — Launch and lives.** *(assumption)* The virus rides on the injector until Up/Space/Enter launches it. Missing it costs a life; 2 lives (L8: "tener máximo 2 vidas"; was 3). Losing the last one ends the round.
- **S6 — Blocks and score.** *(assumption)* 5 rows of 4-cell blocks; the top row is hardened (2 hits). Each hit that damages a block scores 10.
- **S7 — Power-ups.** *(assumption)* A destroyed block may drop a power-up (chance from an injected RNG). It falls; catching it with the injector applies it: **Wide** (wider injector for a while) or **Life** (restores a lost life, never above 2; was +1 up to 5).
- **S8 — Win.** *(assumption)* Destroying every block of the last level (S12) ends the round with the title "FIREWALL BREACHED"; losing all lives shows "Game Over". Requires the session's Game Over title to come from the round when the round provides one.
- **S9 — Controls and flow.** *(assumption)* Left/Right (arrows, A/D) move the injector: a press moves it one cell on that same tick, then it glides at 40 cells/s for 6 ticks (100 ms, longer than a typical key-repeat interval), and key repeats keep it gliding while held (L8, L10: "me gustaria que fuera mas natural el movimiento"; was 2 cells per press). Only a virus entering the injector row from above bounces; one that slipped past the edge keeps falling; Up/W/Space/Enter launch. Pause, Restart and Menu come from `internal/engine/session`.
- **S10 — Launcher.** *(assumption)* "Breach" appears in the launcher with its controls hint, and `play breach` starts it.
- **S12 — Levels.** User (L4): "mete más niveles m, al menos 10 con dificultad progresiva y evidente". *(assumption, how)* 10 levels played in a row in one round; score and lives carry over. Each level is harder on every lever: total block strength strictly rises, virus speed strictly rises (100% → 154%, every component still below one cell per tick), power-ups get rarer, 3-hit blocks appear, and from level 6 indestructible blocks (they bounce the virus and do not count for clearing the level). Clearing a level loads the next with the virus back on the injector; clearing level 10 breaches the firewall (S8). HUD shows `Level N/10` and a `LEVEL N` banner before each launch. Restart starts again from level 1.
- **S11 — Determinism.** Per AGENTS.md: injectable clock and RNG, scripted input, headless simulation tests; no sleep-based tests.

## Tasks

| ID | Specs | Route | Status | Evidence |
| --- | --- | --- | --- | --- |
| T1 | S8 | inline | done (uncommitted) | RED: `title = "Game Over", want the round's title` → GREEN with optional `session.OverTitler`; other games unchanged |
| T2 | S2–S6, S11 | inline | done (uncommitted) | RED: build failed → GREEN `game_test.go`; mutation check: dead `vy > 0` guard removed, right-wall boundary pinned; autopilot clears the default wall on 10 runs without losing a life (a fixed-centre autopilot loops forever on one run: classic periodic path, a player varies the hit point) |
| T3 | S7, S11 | inline | done (uncommitted) | RED: `WideTicks` undefined → GREEN `powerup_test.go`; all 7 mutants caught |
| T4 | S1, rendering | inline | done (uncommitted) | RED: build failed → GREEN `render_test.go`; 80x24 preview checked (wall, drop, win overlay) |
| T5 | S8, S9 | inline | done (uncommitted) | RED: `NewModule` undefined → GREEN `module_test.go`; `-race` clean |
| T7 | S12 | inline | done (uncommitted) | RED: `levels_test.go` build failed → GREEN; autopilot clears all 10 levels × 3 seeds (fails at level 5 if 3-hit blocks could not break); sealed-pocket rule RED on level 9 → GREEN after moving solid blocks; `-race` suite green |
| T8 | S2, S5, S7, S9 | inline | done (uncommitted) | RED: `starting lives = 3, want 2`; `TickPeriod`/`PaddleGlide` undefined → GREEN; test velocities doubled to keep crossing one cell per tick at 200 sub-cells |
| T9 | S2, S9 | inline | done (uncommitted) | RED: `tick period = 25ms, want 60`, `PaddleSpeed` undefined, frozen virus at (15,10) → GREEN; autopilot clears 10 levels × 3 seeds |
| T10 | S2 (rendering) | inline | done (uncommitted) | RED: `glyphVirusLow` undefined → GREEN `TestVirusIsDrawnInTheHalfOfTheCellItIsIn`; half-cell boundary mutant caught; 80x24 preview checked |
| T6 | S10 | inline | done (uncommitted) | RED: 3 shell tests (5 games, `play breach`, controls hint) → GREEN; README run command added |

## Follow-ups (not done, out of scope)

- Manual play-test in a real terminal not done by the agent (needs a TTY).

Routing: every unit depends on T2's model; no independent parallel units, no writer delegation.

## Log

- **L1** (2026-11-08, user): "El Firewall aunque el nombre sería mejor otro estilo un virus intentando penetrsr, no siempre tenernos que ser los buenos"
- **L2**: Name "Breach". Angles Steep (25,55), Diagonal (45,45), Shallow (60,30) sub-cells/tick; injector 7 wide (Wide 11 for 300 ticks), moves 2 cells per press; drops 1 in 5, fall 30 sub-cells/tick. Playfield 40x16, 50 ms ticks.
- **L3**: Final checks: `go vet ./...` clean, `go test -race ./...` fully green (16 packages). Not committed: commits wait for the user's request.
- **L4** (2026-11-08, user): "Sí y mete más niveles m, al menos 10 con dificultad progresiva y evidente"
- **L5**: Committed the first version as `a14ad31` (session OverTitler), `28d184b` (Breach), `a09a4d5` (launcher). Levels go in T7.
- **L6**: Campaign stats (rows / hits to clear / indestructible / speed / drops): L1 3/30/0/100%/1-in-4 → L10 8/184/10/154%/1-in-10, every lever monotonic and pinned by `TestEveryLevelIsHarderThanTheOneBefore`. HP glyphs: ▐██▌ 3, ▐▓▓▌ 2, ▐░░▌ 1, ▐╳╳▌ indestructible. Launch hint moved from the HUD to under the `LEVEL N` banner (HUD would overflow 40 cells).
- **L7**: The autopilot found blocks sealed between two indestructible blocks in one column (levels 9 and 10), reachable only through a 2-row side corridor. New design rule `TestNoBlockIsSealedBetweenIndestructibleOnes`; solid blocks moved, strength unchanged. Autopilot now picks a seeded random landing point each descent: fixed aim patterns fall into periodic paths.
- **L8** (2026-11-08, user): "el brech tampoco esta mal, pero un poquito mas de fluidez estaria bien y tener máximo 2 vidas si no se acaba el juego"
- **L9**: Committed the levels as `609f336`. Lives: start and cap 2. Fluidity: tick 50 → 25 ms with `SubCells` 200 and `WideTicks` 600, so real-time speeds are unchanged (`TestRealTimeSpeedsDoNotDependOnTheTickRate`); the shell takes the period from `breach.TickPeriod`. The injector glides one cell per tick instead of jumping 2 cells per press; a tap now moves it 3 cells.
- **L10** (2026-11-08, user): "lo noto aun poco fluido, me gustaria mas a 60 a ver si se nota mas, [...] y en breach, el primer toque de desplazamiento tarda un poco en reaccionar, depsues si que es mas fluido, me gustaria que fuera mas natural el movimiento"
- **L11**: 60 ticks/s; Steep (17,37), Diagonal (30,30), Shallow (40,20), DropSpeed 20, WideTicks 900. Injector: sub-cell glide (`PaddleSpeed` 67 per 100 ticks = 40 cells/s), the first tick of a fresh press always moves. The autopilot found a real bug: a virus that slipped past the injector's edge into its row froze forever if the injector slid under it (the bounce changed velocity without moving the virus). Fixed by bouncing only on entry from above; pinned by `TestVirusBesideTheInjectorIsNotCaughtByMovingUnderIt`. The autopilot now moves the injector directly, since the glide is not what it checks. Open: the pause after the first press is the OS key-repeat delay (terminals report no key release).
- **L12** (2026-11-08, user): "probe la 1, va algo mejor, lo tendre en cuenta, pero la fluidez de la bola en breach todavia va un poco a trompicones, por lo demas todo bien"
- **L13**: The virus is drawn as a half block (`▀` upper half, `▄` lower half of its cell) instead of `●`, so on screen it moves in half-cell vertical steps, roughly square with the horizontal ones. Collisions are unchanged (cell-based).
