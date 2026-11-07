# Sokoban Run Mode Design

This design evolves Sokoban from a sequence of loose levels into a single continuous run. It keeps the current session-only scope, removes selector/replay concerns, and turns the game into a linear challenge where finishing all levels in order is the point.

## Quick path

1. Add a run-session state above the current level gameplay.
2. Replace per-level undo logic with a global pool of 4 undos for the full run.
3. Rework pause/help/final overlays and end-of-run summaries around linear run semantics.

## Outcome

Sokoban becomes a strict run mode: one continuous attempt, fixed level order, no replay of cleared levels, four total undos for the entire run, and full reset on quit or failure.

## Scope

This change includes:

- linear session-only progression
- a global pool of 4 undos for the entire run
- run-aware HUD and overlays
- end-of-run summaries for success and failure
- tests for run-state rules

This change does not include:

- persistence across app launches
- level selector UX
- replaying already-cleared levels
- multiple pressure systems beyond the undo pool
- packs stored or unlocked across sessions

## Decision summary

| Topic | Decision |
|-------|----------|
| Progress model | Strict linear run |
| Level access | No selector; levels play in fixed order |
| Session persistence | In-memory only for now |
| Retry model | No replay of already-cleared levels |
| Pressure model | One shared pool of 4 undos for the whole run |
| Run failure | Quitting or failing ends the run and resets progress |
| Metrics | Time, level reached/completed, undos used, undos remaining, final result |

## Core concept

Sokoban Run Mode is not “Sokoban plus counters.” It is a continuous challenge.

The player starts at level 1 and must clear every level in order in a single session. Every undo comes from the same finite global pool. Once a level is cleared, it is behind the player. If the player quits or the run is considered failed, the run ends and the whole session progress is lost.

This gives the summary screen meaning and makes every undo a strategic decision instead of a local convenience.

## Run rules

### Progression

- Run starts at level 1.
- Levels advance automatically in fixed order.
- Cleared levels cannot be replayed inside the same run.
- There is no level selector in this mode.

### Undo economy

- The run starts with **4 total undos**.
- Undo usage is global, not per-level.
- Using undo consumes 1 from the shared pool.
- If the pool reaches 0, the player may continue playing, but without undo support.

### Reset and exit

- Quitting the run ends the run.
- Restarting the run ends the current attempt and begins again from level 1.
- There is no “restart current level and keep run progress” escape hatch in this mode.

### Victory and failure

- Victory = complete all levels in order.
- Failure = abandon the run, or any explicit future fail state introduced later.
- For this phase, the only strong pressure system is the shared undo pool; no extra score/health/lives model is added.

## UX and feedback

## HUD

The gameplay HUD should always show:

- current level / total
- total run time
- undos remaining globally
- that the player is inside an active run

This shifts the UI from “single-level puzzle” toward “ongoing run.”

## Overlays

### Pause overlay

Allowed actions:

- continue
- abandon run

Do not offer:

- next level
- restart current level
- replay earlier levels

### Help overlay

Must explain:

- controls
- that undos are global to the full run
- that quitting ends the run

### Run failed overlay

Shown when the player abandons the run or the run ends unsuccessfully.

Should summarize:

- level reached
- time accumulated
- undos used
- undos remaining

### Run completed overlay / summary

Shown only when the full level chain is completed.

Should summarize:

- total run time
- total undos used
- undos remaining
- full completion result

## Metrics

Initial metrics should stay intentionally simple:

- total run time
- highest level reached
- levels completed
- undos used
- undos remaining
- final result (`completed` or `abandoned`)

Do not introduce a synthetic score formula yet. The run itself is the result.

## Structural design

The run state should live above the current Sokoban level state.

That means the current per-level `Game` should remain focused on board rules, while a run/session structure should own cross-level state such as:

- current level index
- total undos remaining
- accumulated run time
- run result
- summary metrics

This is important: the shared overlay package remains UI infrastructure. It must not own run rules, undo economy, or progression semantics.

## Testing strategy

Add deterministic tests for:

- run starts with 4 undos
- undo consumption spans multiple levels
- cleared levels are not replayable inside a run
- quitting ends the run and resets session progress
- completing all levels yields the completed summary
- abandoning produces the failed summary
- pause overlay only offers run-valid actions

Avoid sleep-heavy tests and frame-by-frame visual tests.

## Review checklist

- [ ] Sokoban is modeled as one continuous run.
- [ ] There is no level selector in this phase.
- [ ] The only pressure system is a global pool of 4 undos.
- [ ] Quitting or restarting the run resets progress.
- [ ] Overlays reinforce run rules instead of weakening them.
- [ ] Metrics are clear and minimal.

## Next step

Implement a `RunSession`-style layer for Sokoban, migrate undo usage to the global pool, simplify pause behavior around run abandonment, and add run-summary tests.
