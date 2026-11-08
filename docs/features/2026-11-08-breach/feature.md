# Breach (game 5, Breakout-inspired)

Objective: fifth game of the roadmap; introduces bounce physics with angles and power-ups. The player is the attacker: a virus bouncing off an injector to break through a firewall.
Branch: `feat/breach` (from `feat/drone`) · Delivery: `ask-on-risk` · Test runner: `go test ./...`

## Specs

Source requirements: README roadmap row `| 5 | Breakout / Arkanoid | Bounce physics, angles, power-ups |` and the user's theme choice (L1): "El Firewall aunque el nombre sería mejor otro estilo un virus intentando penetrsr, no siempre tenernos que ser los buenos". Everything marked *(assumption)* is a proposed design default, not a user requirement.

- **S1 — Theme.** The ball is a virus, the paddle an injector at the bottom, the bricks firewall blocks. Breaking every block breaches the firewall. Original game: no names, art or levels from Breakout or Arkanoid.
- **S2 — Physics.** *(assumption)* Integer fixed-point, 100 sub-cells per cell. The ball moves one axis at a time and each velocity component stays below one cell per tick, so it never tunnels through a block.
- **S3 — Bounces.** *(assumption)* Walls and ceiling reflect the matching velocity component. A block hit reflects the component of the axis that hit it.
- **S4 — Angles.** *(assumption)* Where the virus lands on the injector sets the exit angle: outer, middle or inner thirds of each half give shallow, diagonal or steep angles, mirrored left and right. The exact centre keeps the current horizontal direction.
- **S5 — Launch and lives.** *(assumption)* The virus rides on the injector until Up/Space/Enter launches it. Missing it costs a life; 3 lives. Losing the last one ends the round.
- **S6 — Blocks and score.** *(assumption)* 5 rows of 4-cell blocks; the top row is hardened (2 hits). Each hit that damages a block scores 10.
- **S7 — Power-ups.** *(assumption)* A destroyed block may drop a power-up (chance from an injected RNG). It falls; catching it with the injector applies it: **Wide** (wider injector for a while) or **Life** (+1 life, max 5).
- **S8 — Win.** *(assumption)* Destroying every block of the last level (S12) ends the round with the title "FIREWALL BREACHED"; losing all lives shows "Game Over". Requires the session's Game Over title to come from the round when the round provides one.
- **S9 — Controls and flow.** *(assumption)* Left/Right (arrows, A/D) move the injector 2 cells per press; Up/W/Space/Enter launch. Pause, Restart and Menu come from `internal/engine/session`.
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
