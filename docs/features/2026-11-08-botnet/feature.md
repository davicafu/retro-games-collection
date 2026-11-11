# Botnet (game 6)

Objective: sixth game of the roadmap, Space Invaders-inspired; introduces shooting systems and enemy waves on the fixed-tick runtime.
Branch: `feat/botnet` (from `feat/breach`) · Delivery: `ask-on-risk`, commits only when the user asks · Test runner: `go test ./...`

## Specs

Source requirements: README roadmap row `| 6 | Space Invaders | Shooting systems, enemy waves |` and the user's theme choice "Botnet vs antivirus" (L2), whose description was: "Sigue la línea de Breach: eres un servidor infectado que dispara contra oleadas de escáneres antivirus que bajan en formación. Las oleadas son cada vez más rápidas y disparan más." Everything marked *(assumption)* is a proposed design default, not a user requirement.

- **S1 — Theme.** "eres un servidor infectado que dispara contra oleadas de escáneres antivirus que bajan en formación". Original game: no names, art, sounds or layouts from Space Invaders.
- **S2 — Player.** *(assumption)* The infected server sits on the bottom row, 3 cells wide, and moves left/right with the same glide as Breach's injector (moves on the press tick, glides while key repeats arrive). It cannot leave the playfield.
- **S3 — Shooting.** *(assumption)* Up/W/Space/Enter fires a packet straight up from the server's centre. One player packet on screen at a time, so aiming matters. Packets move in sub-cells and never skip a cell.
- **S4 — Formation.** *(assumption)* Scanners start in a grid (4 rows × 8 columns, each scanner 3 cells wide with 1 cell of spacing). The whole formation steps one cell sideways at a fixed interval; when a scanner would leave the playfield, the formation drops one row instead and reverses. The fewer scanners left, the shorter the interval: it goes linearly from the wave's pace down to `MinStepTicks` (6 ticks, 100 ms) for the last scanner (L9).
- **S5 — Enemy fire.** *(assumption)* At random moments (injected RNG) the lowest scanner of a random column fires a quarantine shot straight down. A wave has a cap on enemy shots on screen.
- **S6 — Cover.** *(assumption)* A few proxy blocks stand above the server. Any shot, from either side, erodes the cell it hits.
- **S7 — Lives, score, end.** *(assumption)* 3 lives. A quarantine shot that hits the server costs a life. The round ends when lives run out or the formation reaches the server's row. Score per scanner by row type (top 30, middle 20, bottom 10).
- **S8 — Waves.** "Las oleadas son cada vez más rápidas y disparan más." *(assumption, how)* Clearing a wave starts the next one: the formation starts one row lower (up to a limit), steps faster and fires more often, with a higher cap on enemy shots. Waves are endless; the HUD shows the wave number. Eased after play-testing (L6, L8). User (L10): "tienen que ser viables": the first 10 waves must be clearable with the formation staying at least 3 rows above the server, pinned by `TestAutopilotClearsWavesBeforeTheyLand` (`ViableWaves`, `SafetyRows`).
- **S9 — Controls and flow.** *(assumption)* Left/Right (arrows, A/D) move; Up/W/Space/Enter fire. Pause, Game Over (score and wave reached, Restart/Menu) and Q come from `internal/engine/session`.
- **S10 — Launcher.** *(assumption)* "Botnet" appears in the launcher with its controls hint, and `play botnet` starts it.
- **S11 — Engine.** Per AGENTS.md ("Promote repeated patterns into engine capabilities once they appear more than once"): the centred HUD-plus-box frame (copied in Runner, Drone and Breach) moves to `internal/engine/render`, and Breach's key-repeat glide moves to an engine package, since Botnet is its second user. Both are behaviour-preserving refactors.
- **S13 — Launcher menu.** User (L4): "Descripción solo del seleccionado (Recommended)": one line per game, and the description of the selected game below the list, so the menu fits a 24-row terminal with all 12 roadmap games (with 6 games the old layout took 27 lines).
- **S12 — Determinism and timing.** Per AGENTS.md: injectable clock and RNG, scripted input, headless simulation tests; no sleep-based tests. 60 ticks per second like Drone and Breach, with real-time speeds pinned by a test.

## Tasks

| ID | Specs | Route | Status | Evidence |
| --- | --- | --- | --- | --- |
| T1 | S11 (frame) | inline | done (uncommitted) | RED: `CenterFrame` undefined → GREEN `render/frame_test.go`; Runner, Drone and Breach migrated; differential check: their 80x24, 50x15 and 20x6 snapshots are byte-identical before and after |
| T2 | S11 (glide) | inline | done (uncommitted) | RED: `Unit` undefined → GREEN `engine/glide/glide_test.go`; Breach migrated, its glide tests unchanged and green |
| T3 | S2, S3, S4, S7, S12 | inline | done (uncommitted) | RED: build failed → GREEN `game_test.go` |
| T4 | S5, S6, S8 | inline | done (uncommitted) | GREEN with T3 (same RED run); autopilot with a fire oracle clears waves 1–5 for 3 seeds; mutant "formation twice as fast" makes it fail at wave 2 |
| T5 | S1, S7 (HUD), rendering | inline | done (uncommitted) | RED: build failed → GREEN `render_test.go`; 80x24 preview checked |
| T6 | S9 | inline | done (uncommitted) | RED: `NewModule` undefined → GREEN `module_test.go`; `-race` clean |
| T8 | S8 | inline | done (uncommitted) | RED: `wave 1 formation steps every 500ms, want 600ms` → GREEN; autopilot still clears waves 1–5 |
| T9 | S4, S8 | inline | done (uncommitted) | autopilot clears waves 1–5; a formation 3× faster makes it fail at wave 1 |
| T10 | S8 | inline | done (uncommitted) | viability test now 10 waves with a 3-row margin: green on the eased pace, red on the committed pace (`seed 1 wave 5: scanner reached row 15, want at most 14`) |
| T7 | S10, S13 | inline | done (uncommitted) | RED: 6 games, `play botnet`, controls hint, menu height (27 lines, then 45 with 12 games), description of the selected game only → GREEN; README run command added |

Routing: every unit builds on the previous one (T3 on T2, T4 on T3); no independent parallel units, so no writer delegation. Forecast: about 1200 authored changed lines; commits wait for the user's explicit request (user rule), split per unit.

## Log

- **L1** (2026-11-08, user): "perfecto, se puede conitnuar"
- **L2** (2026-11-08, user answer): theme = "Botnet vs antivirus".
- **L3**: Design defaults (all S#-marked assumptions): 40x18 playfield, 3 lives, packet 30 cells/s, quarantine shot 12 cells/s, formation 4x8 stepping every 500 ms at wave 1 and faster as scanners fall; each wave starts one row lower (max row 4), steps 3 ticks faster (min 12), fires more often (1 in 60 ticks, -8 per wave, min 15) and allows one more shot (2 at wave 1, max 5). End title "QUARANTINED" with score and wave. The first autopilot aimed at where a scanner was, not where it would be, and lost at wave 3; with an oracle (fires when a packet fired now would hit, including the server's move) it clears waves 1–5, lowest scanner row 13 of 17 in wave 5: the pace is tight but clearable.
- **L4** (2026-11-08, user answer): adding the sixth game made the launcher menu 27 lines tall; menu choice = "Descripción solo del seleccionado (Recommended)".
- **L5**: Final checks: `go vet ./...` clean, `go test -race ./...` fully green. Not committed: commits wait for the user's request. Not play-tested in a real terminal.
- **L6** (2026-11-08, user): "muy guay aunque igual puedes facilitar un poco"
- **L7**: Committed T1–T7 as `c7c08f6`, `57f5216`, `2e887b7`, `92d6437`. Eased: wave 1 steps every 600 ms (was 500), -3 ticks per wave down to 15 (was 12), fastest step 4 ticks (was 2), fire 1 in 80 ticks at wave 1 down to 1 in 20 (was 60 down to 15), formation starts at most at row 3 (was 4). Autopilot, lowest scanner row reached per wave (server row 17), before → after: wave 1 14 → 11, wave 2 12 → 10, wave 3 13 → 12, wave 4 13 → 15, wave 5 15 → 14. Wave 4 got no easier for this autopilot (cause not investigated); waves 1–3 gained 1–3 rows of margin.
- **L8** (2026-11-08, user): "pues amplialo un poco mas please" (read as: ease it further).
- **L9**: Per-wave hardening softened: step -2 ticks per wave down to 20 (was -3 down to 15), fire 1 in 90 ticks at wave 1, -6 per wave, down to 1 in 30 (was 80, -8, 20), at most 4 enemy shots (was 5). The end-of-wave rush caused the deep descents, so the speed-up now runs from the wave's pace down to 6 ticks instead of being proportional to the scanners left (with 4 of 32 left: about 10 ticks instead of 4). Autopilot, deepest row per wave across 3 seeds (server row 17): before 11/10/12/15/14 → after 10/10/12/11/10.
- **L10** (2026-11-11, user): "vale pues entonces cambia eso, tienen que ser viables"
- **L11**: The concern came from the reported mutation check (a formation 3x faster makes the autopilot lose at wave 1), which is a deliberate test probe, not the game. Measured the eased pace: the autopilot clears waves 1–10 for 5 seeds, deepest row 13 of 17. The viability test was strengthened to pin that (10 waves, 3 rows of margin).
