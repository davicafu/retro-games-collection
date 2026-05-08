# Sokoban Hardening Design

This design stabilizes Sokoban after the shared overlay migration. It fixes user-visible bugs and the immediate architectural causes behind them, in the same flow-oriented order, so the game becomes safer to extend without doing unrelated refactors.

## Quick path

1. Separate level outcomes so quit, restart, next-level, and completion stop sharing ambiguous control flow.
2. Tighten the level loop and timing behavior without forcing Sokoban into Snake's runtime prematurely.
3. Harden pause/help/summary overlay behavior and then lock the flows with deterministic tests.

## Outcome

Sokoban keeps the new reusable overlay system, but its flow contracts become explicit and testable. The game stops relying on implicit `nil/error` conventions, reduces timing ambiguity, and becomes safer to continue building on.

## Scope

This work targets only Sokoban gameplay hardening.

It includes:

- quit / progression / completion semantics
- level loop and timing behavior
- overlay flow behavior inside Sokoban
- tests around the corrected flows

It does not include:

- new Sokoban features
- launcher changes
- a full migration of Sokoban into the shared runtime used by Snake

## Decision summary

| Topic | Decision |
|-------|----------|
| Delivery strategy | Harden by critical flow, not by layer or file |
| First priority | quit / progression / completion |
| Runtime convergence | Do not force Sokoban into Snake's runtime yet |
| Overlay direction | Keep shared overlay package; harden Sokoban behavior on top |
| Testing strategy | Add contract tests around each corrected flow |

## Flow order

### 1. Quit / progression / completion

This is the highest-risk flow because it defines whether Sokoban understands what just happened.

Today the module still risks conflating outcomes that are not equivalent:

- completed level
- restart level
- skip to next level
- quit to menu

The design requirement is to replace implicit `duration + nil/error` conventions with an explicit level result contract. The level loop must return a domain-level outcome that the top-level `Run()` function can interpret without guessing.

Expected states include concepts equivalent to:

- `completed`
- `restart`
- `next-level`
- `quit-to-menu`

The exact type can vary, but the semantics must be explicit.

### 2. Level loop and timing

Once the result contract is explicit, the level loop must stop hiding control flow inside timing and ad hoc redraw logic.

The goal is not to rewrite Sokoban around the Snake runtime now. The goal is to stabilize the existing loop so responsibilities are clearer:

- receive input
- update game state
- render
- measure elapsed time consistently
- return explicit flow outcomes

This block should reduce accidental behavior from manual sleeps and scattered transition handling, while staying intentionally local to Sokoban.

### 3. Pause / help / overlays

Sokoban already uses the shared in-game overlay package. This block hardens behavior rather than redesigning the UI layer.

The design requirement is consistent overlay semantics:

- pause is an action overlay
- help is an informational overlay
- level-complete and final-summary flows must have predictable exit behavior
- dismiss and confirm behavior must stay explicit
- overlay loops must not swallow errors or block forever without a safe exit path

This validates the overlay abstraction through behavior, not only through successful compilation.

### 4. Tests and determinism

Each corrected flow must end with focused tests.

Priority tests:

- direct quit exits without accidental progression
- pause quit exits without accidental progression
- completion records time only on real completion
- restart/next-level preserve the intended semantics
- help/pause/summary overlays exit correctly
- cancellation or no-input cases do not hang the module

The project should prefer deterministic contract tests over fragile visual snapshots or sleep-driven tests.

## Architectural direction

### Result contract over implicit signals

The central architectural change is explicit flow modeling.

If `Run()` must infer meaning from `0, nil`, then the contract is wrong. The design requires a level result type or equivalent structure that makes the outcome readable and safe.

### Local stabilization before shared runtime convergence

Sokoban should not be forced into a broader runtime refactor as part of this hardening pass.

Instead:

- stabilize Sokoban locally
- make its contracts explicit
- reduce timing ambiguity
- leave a cleaner seam for future convergence

This keeps scope disciplined and tied to the real game behavior.

### Shared overlay package stays in place

The reusable overlay package remains the correct abstraction. The work here is about using it coherently inside Sokoban:

- no game logic in overlay code
- no hidden semantics in dismiss behavior
- no local ad hoc modal systems returning through the side door

## Constraints

- Fix the contract, not just the symptom.
- Keep refactors tied to the flows being hardened.
- Do not introduce new features into the same change.
- Prefer small testable boundaries over broad architectural cleanup.

## Review checklist

- [ ] Quit, restart, next-level, and completion are explicitly distinguishable.
- [ ] `Run()` no longer guesses the meaning of level outcomes.
- [ ] Level timing is computed only for the correct outcomes.
- [ ] Overlay behavior is consistent with the shared overlay package.
- [ ] Critical Sokoban flows are protected by deterministic tests.

## Next step

Write an implementation plan that follows the same flow order: quit/progression/completion first, then loop/timing, then overlay hardening, then test reinforcement.
