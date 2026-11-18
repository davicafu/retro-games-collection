# Defrag (game 11)

Objective: eleventh game of the roadmap, a falling-block puzzle inspired by Tetris: spatial manipulation, rotations and line clearing on the fixed-tick runtime.
Branch: `feat/defrag`, to be created from `feat/rack` once Rack is committed (commits only when the user asks) · Delivery: `ask-on-risk` · Test runner: `go test ./...`

## Specs

Source requirements: README roadmap row `| 11 | Tetris | Spatial manipulation, rotations, line clearing |` and the user's theme choice "Disk defrag (Recommended)" (L2), whose description was: "Caen fragmentos de archivos sobre un disco duro; los giras y encajas, y cada fila completa es un sector desfragmentado que se libera. Sigue la línea tech y conecta con los discos de Rack." Everything marked *(assumption)* is a proposed design default, not a user requirement.

- **S1 — Theme.** "Caen fragmentos de archivos sobre un disco duro; los giras y encajas, y cada fila completa es un sector desfragmentado que se libera." Original game: no name, music, logo, colour scheme or rotation tables from Tetris; only the genre's falling blocks of four cells.
- **S2 — Disk and fragments.** *(assumption)* The disk is 10 cells wide and 20 tall, drawn 2 characters per cell. Fragments are the seven shapes of four cells, each in its own colour, dealt in shuffled sets of seven so no shape is missing for long. The next fragment, the level and the sectors are shown beside the disk; the score above it.
- **S3 — Moving and rotating.** *(assumption)* Left/right shift a cell (held keys glide), up rotates clockwise, nudging up to two cells sideways or one up when the rotation would overlap, down drops faster, space drops at once. A shadow shows where the fragment will land and dots mark the empty cells; G turns both on and off for the whole session, on by default (L7).
- **S4 — Gravity and locking.** *(assumption)* Fragments fall a row every 0.8 s on level 1, 0.1 s faster each level, reaching 15 rows/s on level 9 (L7). A fragment that rests locks after 0.5 s; a move or rotation while resting restarts that wait, at most 15 times; a hard drop locks at once.
- **S5 — Sectors.** "cada fila completa es un sector desfragmentado que se libera". *(assumption, how)* Full rows flash for a moment, then are removed and everything above falls. Points per clear of 1/2/3/4 rows: 100/400/900/1600 times the level, the square of the rows, so clearing rows together pays a bonus; for 2 s the side panel shows the points, titled DOUBLE!, TRIPLE! or DEFRAG! for 2, 3 or 4 rows (L8); soft drop 1 and hard drop 2 per row. Every 8 rows is a level (L7).
- **S6 — End.** *(assumption)* The round ends when a new fragment cannot appear. End title "DISK FULL"; the summary shows score, level and sectors.
- **S7 — Viability.** User on Botnet (2026-11-11): "tienen que ser viables". *(assumption: applied to Defrag too)* An autopilot that presses one key every 6 ticks (10 keys a second) clears at least 100 rows in at least 27 of 30 seeded games, pinned by a test.
- **S8 — Controls and flow.** *(assumption)* ←→/AD shift, ↑/W rotate, ↓/S soft drop, space hard drop, G guides; the launcher hint reads `[←→↓/ASD] move  [↑/W] rotate  [space] slam  [g] guides` to fit 80 columns. Pause, Game Over (Restart/Menu) and Q come from `internal/engine/session`.
- **S9 — Launcher.** *(assumption)* "Defrag · Tetris-like" appears in the launcher with its controls hint, and `play defrag` starts it.
- **S10 — Engine.** Per AGENTS.md roadmap capability "complex spatial manipulation": shapes on a grid, their rotation in a bounding box, and a well that checks fits, places shapes and clears full rows live in a new `internal/engine/polyomino` package.
- **S11 — Determinism and timing.** Per AGENTS.md: injectable clock and RNG, scripted input, headless simulation tests; no sleep-based tests. 60 ticks per second, with real-time speeds pinned by a test.

## Tasks

| ID | Specs | Route | Status | Evidence |
| --- | --- | --- | --- | --- |
| T1 | S10 | inline | done | RED: `Shape` undefined → GREEN `engine/polyomino/polyomino_test.go` (rotation, fits, drop distance, row clearing, clone) |
| T2 | S2, S3, S4, S5, S6, S11 | inline | done | RED: `Kind` undefined → GREEN `game_test.go`; mutants killed: unlimited lock restarts, no kicks, no level up, no soft drop points, never over, no level multiplier, lock wait kept while falling, unshuffled deal (the last three survived first; tests added) |
| T3 | S7 | inline | done | 30 of 30 games clear 100 rows (L4) |
| T4 | S1, rendering, HUD | inline | done | RED: `Draw` undefined → GREEN `render_test.go`; 80x24 preview found the HUD cut off (L5) |
| T5 | S8, S9 | inline | done | `module_test.go` (keys, end title and summary, run and quit); RED: 4 shell tests (11 games, `play defrag`, menu name, controls hint) → GREEN; README run command and features index row |
| T6 | S5 | inline | done | RED: `g.Bonus` undefined → GREEN `TestClearingRowsTogetherPaysABonus`, `TestTheBonusShowsForTwoSeconds`, `TestDrawTheBonusBesideTheDisk`; four-row test updated from 800 to 1600 |

Routing: each unit builds on the previous one; no independent parallel units, so no writer delegation. Forecast: about 1200 authored changed lines; commits wait for the user's explicit request (user rule).

## Log

- **L1** (2026-11-16, user): "genial, dale dale al siguiente" (to: "¿Empiezo con Tetris mientras tanto?")
- **L2** (2026-11-16, user answer): theme = "Disk defrag (Recommended)".
- **L3**: Design defaults (all S#-marked assumptions): 10x20 disk; shuffled sets of seven shapes; level 1 gravity 48 ticks a row, 5 fewer each level down to 4; lock delay 30 ticks with up to 15 resets; full rows flash for 18 ticks; end title "DISK FULL".
- **L4**: Viability. The autopilot weighs every rotation and reachable column with a well-known hand-tuned board heuristic (landing height, cleared rows, row and column transitions, holes, wells) and presses one key every 6 ticks: rotate, shift, then hard drop. 30 of 30 games clear 100 rows, past level 10 where fragments fall 15 rows/s.
- **L5**: The 80x24 preview showed the HUD cut off: the frame is 32 columns, too narrow for level, score and sectors on one line once the score grows (the first test used score 0). The score stays above the disk; level and sectors moved to the side panel under NEXT, pinned with a score of 1234567.
- **L6**: Checks: `go vet ./...` clean, `go test -race ./...` fully green. Not committed (Rack and the Debris fixes are not committed yet either). Not play-tested in a real terminal.
- **L7** (2026-11-16, user): "casi perfecto ya, cada cuanto se sube de nivel? me gustaria que se acelarara un poco porque eso entiendo que añadira velocidad, y quier que las guias y la sombra de la pieza dodne cae sea una ayuda que se pueda quitar o poner". Answer: every 10 rows, 5 ticks faster a level, top speed on level 10 (90 rows). Changed to every 8 rows and 6 ticks a level: top speed on level 9 (64 rows), pinned by `TestRealTimeSpeedsDoNotDependOnTheTickRate` (I first said level 8; the arithmetic gives 6 ticks on level 8 and the 4-tick top on level 9). Guides: G (new engine input action `guides`, shell keys g/G) hides the dots and the shadow; the setting is shared by every round of a session (`newRounds`), tested by `TestGuidesOffHideTheDotsAndTheShadow` and `TestGTogglesTheGuidesForEveryRoundOfTheSession`. The shell key test passed on first run because the key was wired before the test was written. A new shell test keeps every game's controls hint within 80 columns; Defrag's first hint was 90. Viability still 30 of 30.
- **L8** (2026-11-17, user): "en tetris me gustaria que cuando ejecutas varias lineas a la vez te de mas puntos o un bono o similar". The table already paid more (4 rows 800 vs 4 x 100) but the gain was small and nothing on screen showed it. Now 100/400/900/1600 x level, plus a 2 s note beside the disk. Viability still 30 of 30 (it counts rows, not points).
