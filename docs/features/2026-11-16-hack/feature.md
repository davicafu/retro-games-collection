# Hack (game 12)

Objective: twelfth and last game of the roadmap, "Hack the System": a hybrid of puzzle, arcade and narrative built on a simulated terminal, a network map and short arcade breaches.
Branch: `feat/hack`, to be created once the earlier games are committed (commits only when the user asks) · Delivery: `ask-on-risk` · Test runner: `go test ./...`

## Specs

Source requirements: README roadmap row `| 12 | Hack the System | Hybrid puzzle/arcade/narrative systems |`; the user's answer "me gustaria el 2. de fake terminal pero que no se centrase solo en terminal" (L2) to the option "Una terminal simulada donde escribes comandos (ls, cat, ssh…) para explorar un sistema ficticio, encontrar pistas y contraseñas; ciertos comandos lanzan minijuegos arcade contrarreloj. Más narrativo y de texto."; and "por ahora me vale" (L4) to the proposed mix (L3). Everything marked *(assumption)* is a design default, not a user requirement.

- **S1 — Terminal.** "Terminal (exploración y narrativa): escribes comandos (`ls`, `cat`, `scan`, `connect`, `decrypt`…) en un sistema ficticio; los ficheros y mensajes cuentan la historia y esconden pistas y contraseñas." *(assumption, how)* Commands: `help`, `ls`, `cat <file>`, `scan`, `connect <node>`, `decrypt <file> <key>`, `map`, `clear`. Typed text, backspace and enter work like a shell prompt; output scrolls.
- **S2 — Network map.** "Mapa de red (visual): `map` abre una vista gráfica de los nodos de la red: cuáles has comprometido, cuáles faltan y por dónde avanzar; te mueves por ella con las flechas." *(assumption, how)* Enter connects to the selected node when it is linked to one already owned; backspace returns to the terminal.
- **S3 — Breaches.** "Brechas (arcade): entrar en un nodo protegido lanza un minijuego corto y contrarreloj, de varios tipos: esquivar el ICE en una cuadrícula; alinear bits que caen para formar la clave; un puzzle de puertas lógicas." *(assumption, how)* Three kinds: ICE (reach the port across a grid while sentries chase), Sync (stop a sweeping signal inside the key window, three times), Gates (set the input switches so the circuit outputs 1). Each has a time limit; winning owns the node, losing raises the trace and drops the connection.
- **S4 — Trace.** "Rastreo: un medidor sube con los comandos fallidos y las brechas perdidas. Si se llena, te detectan y pierdes." *(assumption, how much)* Unknown or failed command +4%, wrong password or key +10%, lost breach +20%; at 100% the round ends, end title "TRACE COMPLETE".
- **S5 — Goal and story.** "Pasarlo es llegar al servidor central y leer el último secreto." "Todo es original: historia, nombres y nodos inventados." Reading the final file on the central server wins, end title "ROOT ACCESS"; the summary shows the trace and the time.
- **S6 — Viability.** User on Botnet (2026-11-11): "tienen que ser viables". *(assumption: applied to Hack too)* An autopilot that types the solution and plays each breach at human pace wins at least 27 of 30 seeded games, pinned by a test.
- **S7 — Text input in the engine.** *(assumption, how)* The engine's input carries typed text as well as actions; a round that is taking text gets the letters as text only (Q and P do not quit or pause), while Esc still pauses and Ctrl+C still quits. Every other game keeps its keys exactly as before.
- **S8 — Launcher.** *(assumption)* "Hack" appears in the launcher with its controls hint, and `play hack` starts it. It is an original game, so no "-like" reference.
- **S9 — Determinism and timing.** Per AGENTS.md: injectable clock and RNG, scripted input, headless simulation tests; no sleep-based tests. 60 ticks per second, with real-time limits pinned by a test.

## Tasks

| ID | Specs | Route | Status | Evidence |
| --- | --- | --- | --- | --- |
| T1 | S7 | inline | done (uncommitted) | RED: `RuneAction` undefined, `Text` field missing, `fi.Text` undefined → GREEN `engine/input`, `engine/session` (`TextTaker`), `shell/input.go`; session mutants killed. Risk: item 3 (engine input and session contracts, shell keys) |
| T2 | S1, S4, S5 | inline | done (uncommitted) | RED: `Outcome` undefined → GREEN `game_test.go` (prompt, editing, history, help/ls/cat/clear, failed commands, scan, connect, lost breach, masked password, decrypt, core key, win, full trace, elapsed); mutants killed: wrong password free, password shown, lost breach free, never win, no history, never caught, core without key, no route check (survived first; test added) |
| T3 | S2 | inline | done (uncommitted) | `TestTheMapMovesBetweenNodesAndConnects`, `TestTheMapCommandOpensTheMap`, module test `TestOnTheMapArrowsMoveAndEnterConnects` |
| T4 | S3 | inline | done (uncommitted) | `breach_test.go`: ICE (port wins, walls block, sentries chase and catch, solvable deal and re-deal, sentries stay on open cells), Sync (hit/miss, narrowing, 3 hits/3 misses, bounce), Gates (truth tables, solvable and unsolved start, switch keys), time limits 20/20/30 s; mutants killed: window never narrows, sentries anywhere, unsolvable ICE (survived first; test added), gates may start solved (L5) |
| T5 | S6 | inline | done (uncommitted) | `TestAutopilotHacksTheCore`: 30 of 30 (L5) |
| T6 | rendering, S8, S9 | inline | done (uncommitted) | RED: `Draw` undefined → GREEN `render_test.go` (terminal, scroll and wrap, map, three breaches, fit); `module_test.go` (q and space as text, map keys, end titles and summary, run and quit); RED: 3 shell tests (12 games, `play hack`, hint) → GREEN; README run command and features index row (L6) |
Routing: each unit builds on the previous one; no independent parallel units, so no writer delegation. Forecast: about 1800 authored changed lines; commits wait for the user's explicit request (user rule).

## Log

- **L1** (2026-11-16, user): "dale" (to: "¿Sigo con Hack the System, el último del roadmap, mientras pruebas?")
- **L2** (2026-11-16, user): "me gustaria el 2. de fake terminal pero que no se centrase solo en terminal"
- **L3**: Proposal sent: terminal for exploration and story, a graphical network map, timed arcade breaches of several kinds, and a trace meter; original story, names and nodes.
- **L4** (2026-11-16, user): "por ahora me vale"
- **L5**: Design defaults: trace +4/+10/+20; ICE 21x9 grid, 16 walls, 3 sentries starting at least 8 cells away and moving 4 cells/s, 20 s; Sync bar 30 cells, window 5 narrowing by 1 per hit, cursor 20 cells/s, 20 s; Gates 5 switches and 4 two-input gates (AND, OR, XOR, NAND), 30 s. Story nodes gate → mail (Sync) / dev (ICE) → vault (password "pixel", clue in mail) → core (admin key decrypted with "ember", clue in dev; Gates). The first autopilot run lost 30 of 30: `grid.Unreachable` is -1, so sentries "stepped" off the grid and the autopilot chose unreachable moves; fixed and pinned by `TestSentriesNeverStepIntoWallsOrOffTheGrid`. Then 30 of 30; per breach over 200 seeds: ICE 197, Sync 200, Gates 200.
- **L6**: The 80x24 preview showed the Gates header over the timer and the core's map status past the frame; shortened, core moved to column 58, pinned by `TestEveryViewFitsThePlayfield`. At the terminal q is text, so the launcher hint for Hack says `[ctrl+c] quit` (new `Typing` field on launcher entries); the menu description was shortened to fit the 42-column box. Checks: `go vet ./...` clean, `go test -race ./...` fully green. Not committed. Not play-tested in a real terminal.
- **L7** (2026-11-17, user): "Dale a los commits". Before committing, the controls width test turned out to add `[q] quit` for every game, while Hack's real hint says `[ctrl+c] quit` and was 82 columns. The test now uses the real quit key (RED: "Hack controls hint is 82 columns") and the hint was shortened to `[enter] run  [↑↓] history  [arrows/space] breach` (78). Commits: engine input contract `f3b90ae` on `feat/rack` (T1); the game on `feat/hack`.
