# Records: high scores for every game, save and continue for Hack

Objective: a top-5 high score table with player names for every game, and save and continue for Hack, all reachable from the launcher menu.
Branch: `feat/hack` · Delivery: `ask-on-risk` · Test runner: `go test ./...`

## Specs

Source: the user's message of 2026-11-17 (L1). Everything marked *(assumption)* is a design default, not a user requirement.

- **S1 — High scores.** "para los demás y para este quiero que se establezca un ranking con las cinco mejores puntuaciones y que se le puedan poner nombre". Every game keeps its five best results. When a finished round enters its game's top five, the player types a name before the end screen. *(assumption, how)* Names up to 12 characters; an empty name is saved as "ANON". Equal scores keep the older result first. Sokoban has no score: it ranks by levels completed, then by less total time.
- **S2 — Save and continue, Hack only.** "deberá existir solo para este un sistema de guardado que permita continuar la partida empezar una nueva". *(assumption, how)* Hack saves a checkpoint at the start of every level from level 2: the level, the score so far and the time played. Choosing Hack in the launcher with a save offers "Continue (level N)" or "New game". A new game, a won run or a run ended by the trace deletes the save; leaving mid-level keeps the checkpoint of the level's start.
- **S3 — From the main menu.** "Todo ello debe ser consultable desde el menú principal". *(assumption, how)* In the launcher, R opens the high scores, one game at a time (←→ to switch, Esc or Q to go back); Hack's entry tells when there is a saved run to continue.
- **S4 — Storage.** *(assumption)* JSON files with a format version under the user's config directory (`os.UserConfigDir()/retro-games`): `records.json` and `saves.json`. Writes go to a temporary file and are renamed into place. An unreadable file is kept aside as `<name>.bad` and a fresh one starts, so a corrupt file never blocks playing. Tests only ever write under temporary directories.
- **S5 — Architecture and testing.** Per AGENTS.md: storage is an engine capability, not shell or game code; gameplay stays outside Bubble Tea; deterministic tests, no sleeps.

## Tasks

| ID | Specs | Route | Status | Evidence |
| --- | --- | --- | --- | --- |
| T1 | S1, S4 | inline | done | RED: `Book` undefined → GREEN `internal/engine/records` (11 tests); 12 mutants killed. Risk: item 1 (new persisted format) |
| T2 | S2, S4 | inline | done | RED: `Saves` undefined → GREEN `saves_test.go`; mutants killed: misfit run loads, clear not saved. Risk: item 1 |
| T3 | S1, S5 | inline | done | `session/records_test.go` (10 tests): name prompt, esc saves, table with mark, save error; prompt and table extracted to `ui/highscore` for Sokoban; mutants killed. Risk: item 3 (session gains an optional board; rounds without one are unchanged, pinned) |
| T4 | S1 | inline | done | 11 session games: `Result()` + `WithBoard`; Sokoban: `TestAFinishedRunEntersTheHighScoresWithItsName`, `TestTheTableScreenWaitsForAKey` (L2) |
| T5 | S2 | inline | done | `hack/save_test.go` (8 tests); 8 mutants killed |
| T6 | S3 | inline | done | `shell/records_test.go` (8 tests), `app_test.go` with a temporary HOME; 6 mutants killed |
| T7 | docs | inline | done | README "High scores and saved runs" |

Routing: each unit builds on the store of T1/T2; no independent parallel units, so no writer delegation. Forecast: about 1500 authored changed lines.

## Log

- **L1** (2026-11-17, user): "Y revisa los demás por si alguna mejora / Además deberá existir solo para este un sistema de guardado que permita continuar la partida empezar una nueva para los demás y para este quiero que se establezca un ranking con las cinco mejores puntuaciones y que se le puedan poner nombre. Todo ello debe ser consultable desde el menú principal". Read as: save and continue only for Hack; a named top-5 ranking for every game including Hack; both reachable from the main menu.
- **L2** (2026-11-18): Design decisions. Storage is the engine package `records` (Book for tables, Saves for one run per game), JSON with `version: 1`, written to a temp file and renamed; tests write only under temp dirs (the app test sets HOME to one). The name prompt and table text live in `ui/highscore`, used by `session` (11 games) and by Sokoban, which has its own loop. Sokoban's final summary already fills the screen with per-level times, so its table shows on its own screen before the summary instead of changing that summary. Hack checkpoints when a level is cleared (next level, score so far, time played), so quitting at the "level cleared" screen keeps the level's points; a run whose checkpoint is impossible starts afresh. `play hack` with a save also asks first. If records cannot be opened, the app keeps them in memory for the session. Full `go test -race ./...` and `go vet ./...` pass.
- **L3** (2026-11-18, user): "si dale y haz commit de todo y push a main" — after flattening `docs/features/` into `docs/`. Committed on `feat/hack` and fast-forwarded to `main`.
