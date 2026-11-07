# Endless Runner (game 3)

Objective: third game of the roadmap; introduces scrolling and speed scaling on top of the existing fixed-tick runtime.
Branch: `feat/endless-runner` · Delivery: `ask-on-risk` · Test runner: `go test ./...`

## Specs

Source requirements: README roadmap row `| 3 | Endless Runner | Scrolling systems, speed scaling |` and the user's mechanic choice "Jump + duck, scripted (Recommended)" (L2). Everything marked *(assumption)* is a proposed design default, not a user requirement.

- **S1 — Mechanic.** "Side-scroller; player jumps over ground obstacles and ducks under flying ones. Jump is a fixed tick-by-tick height table, no gravity. Focus stays on scrolling, obstacle spawning and speed scaling; real physics is left for Flappy."
- **S2 — Jump.** *(assumption)* Jump follows a fixed height table indexed by tick (`[1,2,3,4,4,4,3,2,1]`, then back on the ground). Jumping is only possible from the ground; pressing jump mid-air does nothing.
- **S3 — Duck.** *(assumption)* Terminals report key presses but not key releases, so duck is timed: pressing down ducks for a fixed number of ticks; pressing again restarts the timer. Duck is ignored while airborne. Jumping from a duck cancels it.
- **S4 — Bodies and collisions.** *(assumption)* Positions are measured in cells above the ground. A standing player occupies heights 0–1, a ducking player height 0, a jumping player heights `h`–`h+1`. Ground obstacles occupy heights `0..tall-1`; flying obstacles occupy height 1. Any overlap between the player column and an obstacle cell ends the run.
- **S5 — Scrolling and speed scaling.** *(assumption)* Speed is an integer number of cells per 100 ticks. Each tick adds the speed to an accumulator and the world scrolls one cell per 100 accumulated. Speed starts at 35, rises by 5 every 40 cells travelled (L10: "en el runner pon para que el cambio de velocidad sea en 40"; was 20 per L6), and caps at 90 (reached at 440 cells). No floating point, so the simulation stays deterministic.
- **S6 — Spawning.** *(assumption)* Obstacles are generated from an injected RNG (`Intn`). The gap between obstacles always lets the player land and react: at least the cells covered by a full jump at the current speed, plus a margin.
- **S7 — Score.** *(assumption)* Score is the distance travelled in cells. The HUD shows score and speed multiplier (speed / 35, one decimal).
- **S8 — Controls and flow.** *(assumption)* Up/W/Space jump; Down/S duck; Esc/P pause overlay (Continue / Restart / Menu); Q returns to the menu; collision shows a Game Over overlay with the score (Restart / Menu). Same overlay package as Snake; back-to-menu ends the loop with `nil`, like Sokoban. The playfield is centered on the screen so overlays sit on top of it.
- **S9 — Launcher.** *(assumption)* "Runner" appears in the launcher menu and `play runner` starts it directly.
- **S11 — Shared session in the engine.** User (L6): "si es buena decision quitar lo comun al motor". Assessed as good for runtime-based games: pause, game over, restart and back-to-menu move to `internal/engine/session`; Snake and Runner provide only a `Round`. Sokoban stays out (own blocking loop, not on the runtime).
- **S10 — Determinism.** Per AGENTS.md: injectable clock and RNG, scripted input, headless simulation tests; no sleep-based tests.

## Architecture decision

Scrolling stays inside `internal/games/runner` for now. AGENTS.md promotes a pattern into the engine once it appears more than once; Flappy (game 4) is the second scrolling game, so the scroll accumulator is kept small and isolated so it can move to `internal/engine` then, with a real second use to shape its interface.

## Tasks

| ID | Specs | Route | Status | Evidence |
| --- | --- | --- | --- | --- |
| T1 | S1–S5, S7, S10 | inline | done (uncommitted) | RED: build failed, `Game`/`Action` undefined → GREEN `game_test.go`; mutation check caught flying height and duck duration changes; added tunnelling guard `MaxSpeed < 100` |
| T2 | S6, S10 | inline | done (uncommitted) | RED: `NewRandomSpawner` undefined → GREEN `spawn_test.go` (5000-cell gap property, seeded RNG) |
| T3 | S7 (HUD), rendering | inline | done (uncommitted) | First pass written together with its tests (no separate RED). Centering fix: RED `newLayout` undefined → GREEN |
| T4 | S8, S10 | inline | done (uncommitted) | RED: `runtimeGame`/`NewModule` undefined → GREEN `module_test.go`; `-race` clean |
| T6 | S5 | inline | done (uncommitted) | RED: `speed at distance 20 = 35, want 40` → GREEN with `SpeedDistance = 20` |
| T7 | S11 | inline | done (uncommitted) | RED: `Session`/`Round` undefined → GREEN `session_test.go` (`-race` clean); Runner and Snake migrated, their suites green |
| T5 | S9 | inline | done (uncommitted) | RED: `TestNew_CreatesMenuStateWithDefaults` (3 games), `TestNewWithGame_RunnerSetsPlayingState` → GREEN; `-race` clean |

Routing: all units depend on the T1 domain model, so there are no independent parallel units; no writer delegation.

## Baseline failures (resolved, see L9)

Were failing on `main` (5832c18) before this feature:
- `internal/engine/ui/overlay`: `TestDrawCentersEmojiWrappedTitleWithoutExtraGap`
- `internal/games/sokoban`: `TestBuildFinalSummaryOverlayUsesNormalizedTitleSpacing`

## Follow-ups (not done, out of scope)

- Sokoban keeps its own pause/overlay loop; moving it onto the runtime + session is a separate, larger change.
- Manual play-test in a real terminal not done by the agent (needs a TTY).

## Log

- **L1** (2026-11-07, user): "ok pues continua con lo siguiente que toque"
- **L2** (2026-11-07, user answer): mechanic = "Jump + duck, scripted (Recommended)" with preview `jump: [0,1,2,3,3,2,1,0]`.
- **L3**: Explored runtime (fixed-tick `Update`), input actions, shell wiring, Snake module/overlays. Back-to-menu follows Sokoban (return `nil`). Baseline `go test ./...`: two pre-existing failures listed above.
- **L4**: T1–T5 implemented inline with tests. S8 revised: the shell returns to the menu whenever a game ends, so Game Over "Quit" duplicated "Menu" and was dropped. A rendered 80x24 preview showed the Game Over overlay hanging below the playfield and covering the ground line; fixed by centering the playfield (`newLayout`).
- **L5**: Final checks: `go vet ./...` clean; `go test ./...` passes except the two baseline failures; `go test -race` on runner and shell clean. Not committed: per the user's AGENTS.md, commits wait for an explicit request.
- **L6** (2026-11-07, user): "baja el cambio de velocidad a cada 20 celdas, si es buena decision quitar lo comun al motor lo demas lo vi todo bien"
- **L7**: Speed step lowered to 20 cells. Extracted `internal/engine/session` (Session + Round). Behavior changes for Snake, all deliberate: Game Over offers Restart/Menu (its "Quit" had the same effect as Menu); Menu ends with `nil` instead of `snake.ErrBackToMenu` (removed; the shell returns to the menu either way); the overlay is drawn even on screens smaller than Snake's board (only reachable in tests; the shell screen is 80x24); a direction pressed in the same frame as Pause is no longer kept. Kept from Snake: Q confirms the selected Game Over option, Q/Esc resume from pause. Runner now also gets Q-confirms on Game Over.
- **L8** (2026-11-07, user): "perfecto, continua". Closed follow-ups: per-game controls hint in the shell (`gameEntry.Controls`; Runner shows `[↑/W/space] jump  [↓/S] duck`, Snake/Sokoban unchanged), README run commands fixed to `go run .` / `go run . play <game>`.
- **L9**: Baseline failures fixed. Overlay test sliced the row by bytes inside the 3-byte `│`, inflating the left pad by 2; after fixing it the test still failed on the real rule (odd centering space must go right), so the title formula dropped its `+1` (affects every overlay title with an odd remainder: one cell further left). Sokoban title: user chose "🎉 FIN DEL JUEGO!🎉 (test)", restored in code. `go test ./...` and `-race` fully green.
- **L10** (2026-11-08, user): "Cambia ese fallo en el motor y en el runner pon para que el cambio de velocidad sea en 40"
- **L11**: `SpeedDistance` 20 → 40. RED: `speed at distance 39 = 40, want 35` → GREEN; render HUD test moved to score 40 (x1.1). Done on branch `feat/drone`, uncommitted.
