# Rack (game 10)

Objective: tenth game of the roadmap, a platformer inspired by Donkey Kong / Lode Runner: platforms, ladders and gravity on the fixed-tick runtime.
Branch: `feat/rack` (from `feat/debris`) · Delivery: `ask-on-risk`, commits only when the user asks · Test runner: `go test ./...`

## Specs

Source requirements: README roadmap row `| 10 | Donkey Kong / Lode Runner-inspired | Platforms, ladders, gravity |` and the user's theme choice "Data center climb (Recommended)" (L2), whose description was: "Eres un técnico que sube por los racks de un data center: escaleras entre plantas, cables sueltos que caen rodando y bots de vigilancia. Recoges discos duros y llegas a la sala del servidor arriba. Sigue la línea tech." Everything marked *(assumption)* is a proposed design default, not a user requirement.

- **S1 — Theme.** "Eres un técnico que sube por los racks de un data center: escaleras entre plantas, cables sueltos que caen rodando y bots de vigilancia. Recoges discos duros y llegas a la sala del servidor arriba." Original game: no names, characters, levels, sprites or sounds from Donkey Kong or Lode Runner.
- **S2 — Levels.** *(assumption)* Hand-made 36×18 cell maps of rack floors joined by ladders, drawn 2 characters per cell, rotating by level like Leak's maps. Each floor has a gap at one end, alternating sides, so cables zigzag down. Every disk and the exit can be reached from the start.
- **S3 — Movement and gravity.** *(assumption)* Left/right walk, up/down climb ladders. The technician stands on a floor or on top of a ladder; without either it falls and cannot steer until it lands. Space jumps one cell up for a moment, keeping the walk, which clears a cable or a bot. Held keys move smoothly (key-repeat glide).
- **S4 — Cables.** "cables sueltos que caen rodando". *(assumption, how)* A cable spool at the top drops cables at intervals; they roll along a floor, fall through its gap, land and roll back the other way; at a ladder top a cable sometimes goes down the ladder instead. Cables leave at the bottom floor's ends. Jumping over a cable scores 50.
- **S5 — Bots.** "bots de vigilancia". *(assumption, how)* Bots patrol one floor back and forth, walking over ladder tops, turning only at walls and gaps, and never climb.
- **S6 — Disks and exit.** "Recoges discos duros y llegas a la sala del servidor arriba." *(assumption, how)* A disk scores 100. The server room door at the top opens once every disk on the level is collected; going through it scores 500 and starts the next level, where cables come more often and roll faster.
- **S7 — Lives.** *(assumption)* 3 lives. Touching a cable or a bot costs a life; the technician restarts at the level's start with the collected disks kept and the cables cleared. The round ends when lives run out. The HUD shows level, score, disks left and lives.
- **S8 — Viability.** User on Botnet (2026-11-11): "tienen que ser viables". *(assumption: applied to Rack too)* An autopilot that presses a key every 2 ticks (a held key's repeat rate), follows the shortest path to the nearest disk and then the door, and picks each move by trying it a short time ahead, clears the first 3 levels with its 3 lives in at least 27 of 30 seeded games, pinned by a test.
- **S13 — Wrenches.** User (L7): "voy a poder coger cosas y lanzarla a los enemigos o solo puedo recoger puntos y saltar?", then "me parece perfecto" to the proposal: every level has a toolbox; picking it up gives 3 wrenches, thrown with F; a wrench flies straight along the technician's floor, the way it faces, and disables the first bot or cable it touches, for points. Jumping still works. *(assumption, how)* 20 cells/s, over gaps, stopping at walls; one wrench in flight at a time; a bot is worth 300 and a cable 100. Disabled bots come back after a lost life; collected toolboxes and wrenches in hand stay. A new engine input action `fire`, sent by the F key.
- **S9 — Controls and flow.** *(assumption)* Arrows/WASD walk and climb, space jumps, F throws a wrench (S13). Pause, Game Over (score and level, Restart/Menu) and Q come from `internal/engine/session`. End title "ACCESS DENIED".
- **S10 — Launcher.** *(assumption)* "Rack · Donkey Kong-like" appears in the launcher with its controls hint, and `play rack` starts it.
- **S11 — Engine.** Per AGENTS.md roadmap capability "platform structure": the support, ladder and gravity rules on a tile grid live in a new `internal/engine/platform` package, shared by the technician, the cables and the bots.
- **S12 — Determinism and timing.** Per AGENTS.md: injectable clock and RNG, scripted input, headless simulation tests; no sleep-based tests. 60 ticks per second, with real-time speeds pinned by a test.

## Tasks

| ID | Specs | Route | Status | Evidence |
| --- | --- | --- | --- | --- |
| T1 | S11 | inline | done (uncommitted) | RED: `Supported` undefined → GREEN `engine/platform/platform_test.go` |
| T2 | S2 | inline | done (uncommitted) | `maps.go` from the scratch generator; `TestLevelsAreReachableAndShaped` (every disk and the door reachable, 36x18) |
| T3 | S3, S4, S5, S6, S7, S12 | inline | done (uncommitted) | RED: `NewGame` undefined → GREEN `game_test.go`; mutants killed: no reverse on landing, no jump score, bots ignoring gaps, cables never leaving, cables kept on death, jump in the air, no reverse at a ladder foot, no cable speed-up, a jump scored repeatedly (survived first, test added); a redundant glide stop while falling was removed |
| T4 | S8 | inline | done (uncommitted) | 30 of 30 games clear 3 levels (L5) |
| T5 | S1, rendering, HUD | inline | done (uncommitted) | RED: `Draw` undefined → GREEN `render_test.go`; 80x24 preview checked |
| T6 | S9, S10 | inline | done (uncommitted) | `module_test.go` (keys, end title, summary, run and quit); RED: 10 games, `play rack`, menu name, controls hint → GREEN `shell_test.go`; README run command and features index row |
| T7 | S13 | inline | done (uncommitted) | RED: `CellToolbox` undefined → GREEN `game_test.go` (toolbox, throw without wrenches, flight and bot disabled, one in flight, wall and gaps, cable knocked away, bot walking into a wrench, bots back and wrenches kept after a lost life, speed); mutants killed: no wrench spent, several in flight, aim never changing, through walls, no cable points, no check after hazards move and wrench kept in flight on death (the last two survived first; tests added). RED: `input.ActionFire` undefined → GREEN shell `f`/`F` key, round mapping, render and HUD, controls hint; one toolbox per level, reachable |

Routing: each unit builds on the previous one; no independent parallel units, so no writer delegation. Forecast: about 1300 authored changed lines; commits wait for the user's explicit request (user rule).

## Log

- **L1** (2026-11-14, user): "si dale" (to: "¿Sigo con el juego 10, el de plataformas inspirado en Donkey Kong y Lode Runner?")
- **L2** (2026-11-14, user answer): theme = "Data center climb (Recommended)".
- **L3**: Design defaults (all S#-marked assumptions): grid movement like Lode Runner (support = a floor below, a ladder below or a ladder here); walk and climb 8 cells/s, fall 15 cells/s, cables 6.7 cells/s, bots 4 cells/s; jump 0.35 s one cell up; 3 lives; end title "ACCESS DENIED".
- **L4**: S5 refined while building: bots walk over ladder tops (a ladder's top is part of the floor) and turn only at walls and gaps. The cable spool drops a cable every 4 s on level 1, 0.5 s sooner each level down to 2 s; cables roll 6.7 cells/s, one cell/s faster each level up to 10; a cable goes down 1 in 4 ladder tops it passes, and turns back when it lands, at a ladder's foot and at a wall.
- **L5**: Viability. A rule-based autopilot (jump a hazard within 2 cells, wait on ladders) cleared 3 levels in 4 of 30 games; its deaths were landing on a slow bot it had jumped, stepping off a ladder next to a bot, catching up a cable from behind and climbing into a descending cable. Replaced by a look-ahead: each candidate move is played for 24 ticks on a copy of the game, with cables taking every ladder and none, and the first one survived is pressed. Result: 30 of 30 games (100 of 100 measured, no life lost, about 69 s for 3 levels). This shows the levels can be cleared; it says little about how hard they feel to a person, which the play-test decides.
- **L6**: Checks: `go vet ./...` clean, `go test -race ./...` fully green. Not committed: commits wait for the user's request. Not play-tested in a real terminal.
- **L7** (2026-11-16, user): "en el de donkeykong voy a poder coger cosas y lanzarla a los enemigos o solo puedo recoger puntos y saltar?" Answer: only disks and jumps. Proposed: a toolbox per level with 3 wrenches thrown with F, flying along the floor and disabling the first bot or cable hit, for points; jumping unchanged.
- **L8** (2026-11-16, user): "me parece perfecto". Built as S13 (T7). The look-ahead autopilot does not throw; viability still 30 of 30. Checks: `go vet ./...` clean, `go test -race ./...` fully green. Not committed. Not play-tested in a real terminal.
