# Drone (game 4, Flappy-inspired)

Objective: fourth game of the roadmap; introduces gravity and impulse physics, and promotes scrolling into the engine now that a second scrolling game exists.
Branch: `feat/drone` (from `feat/endless-runner`) · Delivery: `ask-on-risk` · Test runner: `go test ./...`

## Specs

Source requirements: README roadmap row `| 4 | Flappy-inspired | Gravity, impulse physics, obstacle generation |`, AGENTS.md ("Promote repeated patterns into engine capabilities once they appear more than once") and the user's theme choice (L1): "Me gusta más el dron entre antenas". Everything marked *(assumption)* is a proposed design default, not a user requirement.

- **S1 — Theme.** A drone flies right through gaps between antenna masts that grow from the ceiling and the ground. Original game: no names, art or levels from Flappy Bird.
- **S2 — Physics.** *(assumption)* Integer fixed-point, 100 sub-cells per cell, no floating point. Each tick gravity adds to the vertical velocity, capped at a terminal fall speed; thrust replaces the velocity with a fixed upward impulse. The drone's row is its position divided by 100.
- **S3 — Ready state.** *(assumption)* The drone hovers in place until the first thrust, so a round does not start falling before the player reacts. The world does not scroll while hovering.
- **S4 — Bounds.** *(assumption)* The ceiling stops the drone (position clamped to row 0, velocity reset); touching the ground ends the round.
- **S5 — Antennas.** *(assumption)* Each antenna is 2 columns wide with a vertical gap of 5 rows. Antennas are spaced at a fixed distance; the gap position comes from an injected RNG and moves at most a fixed number of rows from the previous antenna's gap, so every gap is reachable.
- **S6 — Collisions.** *(assumption)* The drone occupies one cell; overlapping any mast cell ends the round.
- **S7 — Score.** *(assumption)* One point per antenna passed (when its right edge goes past the drone column). HUD shows the score.
- **S8 — Scrolling and speed in the engine.** *(assumption)* The scroll accumulator and the speed ramp used by the Runner move to `internal/engine/scroll`; Runner and Drone both use it. Runner behavior stays identical (pinned by its existing tests). Drone speed ramps with distance and stays below one cell per tick.
- **S9 — Controls and flow.** *(assumption)* Up/W/Space/Enter thrust. Pause, Game Over (score summary, Restart/Menu) and Q come from `internal/engine/session`.
- **S10 — Launcher.** *(assumption)* "Drone" appears in the launcher menu with its controls hint, and `play drone` starts it directly.
- **S11 — Determinism.** Per AGENTS.md: injectable clock and RNG, scripted input, headless simulation tests; no sleep-based tests.

## Tasks

| ID | Specs | Route | Status | Evidence |
| --- | --- | --- | --- | --- |
| T1 | S8 | inline | done (uncommitted) | RED: `Unit`/`Scroller`/`Ramp` undefined → GREEN `scroll_test.go`; Runner migrated, its unchanged suite green |
| T2 | S2–S4, S6, S7, S11 | inline | done (uncommitted) | RED: build failed → GREEN `game_test.go`; mutation check: ground and gap-bottom boundaries survived, pinned with two edge tests, then caught |
| T3 | S5, S11 | inline | done (uncommitted) | RED: `MaxGapShift` undefined → GREEN `spawn_test.go`; autopilot survives 20 seeds × 8000 ticks at MaxSpeed; mutants `MaxGapShift = 7` and `AntennaSpacing = 5` make it crash |
| T4 | S7 (HUD), rendering | inline | done (uncommitted) | RED: build failed → GREEN `render_test.go`; 80x24 preview checked (playfield and Game Over overlay) |
| T5 | S9 | inline | done (uncommitted) | RED: `NewModule` undefined → GREEN `module_test.go`; `-race` clean |
| T7 | L4 | inline | done (uncommitted) | RED: `TestWriteTextPlacesOneRunePerCell`, `TestPadRightCountsRunes` → GREEN; full suite `-race` green |
| T6 | S10 | inline | done (uncommitted) | RED: 3 shell tests (4 games, `play drone`, controls hint) → GREEN; README run command added |

## Follow-ups (not done, out of scope)

- Screen-centering `layout` is now duplicated in Runner and Drone (Snake has its own); a candidate for `internal/engine/render` when a third game needs it.
- Manual play-test in a real terminal not done by the agent (needs a TTY).

Routing: every unit depends on T1/T2; no independent parallel units, no writer delegation. Commits wait for the user's explicit request (user rule).

## Log

- **L1** (2026-11-08, user): "Me gusta más el dron entre antenas, me parece bien"
- **L2**: Physics tuned to Gravity 5, Thrust 45, MaxFall 60 (sub-cells/tick, 50 ms ticks): one thrust climbs about 2.25 rows; a thrust every 19 ticks holds altitude. Speed 40→70 (+5 every 30 cells). Antennas 2 wide, gap 5 rows, spacing 16, gap shift ≤ 3 rows.
- **L3**: Found the byte-indexed `render.WriteText` bug while writing the HUD hint; reported instead of fixed (out of scope, changes Sokoban output). Final checks: `go vet ./...` clean, `go test -race ./...` fully green. Not committed: commits wait for the user's request.
- **L4** (2026-11-08, user): "Cambia ese fallo en el motor y en el runner pon para que el cambio de velocidad sea en 40"
- **L5**: `render.WriteText` and `render.PadRight` now count runes (one rune per cell, matching `BufferScreen`). Visible effect: Sokoban's header (`⏱`) and hints (`[↑↓←→]`) no longer have gaps. Runner speed step change logged in its own feature doc (L10–L11).
