# Snake Continuous Food Adjustment

## Context

The current Snake implementation is structurally correct for movement, growth, and score, but it still behaves like a short technical demo instead of a real game session.

At the moment, the game starts with a single food item. When the snake eats that food, the game quickly reaches a terminal condition because no new food is spawned. As a result, `play snake` opens a visible session but exits almost immediately.

That is not acceptable for the intended milestone. The first playable game should behave like a real playable loop, not like a one-food proof of concept.

## Updated Goal

Snake must become a **continuous playable loop** by spawning new food whenever the current food is consumed.

This change must preserve deterministic testing while still allowing real gameplay behavior in the interactive runtime.

## Product Behavior

The updated Snake rules are:

1. The game starts with an initial food item
2. When the snake eats food:
   - the snake grows
   - the score increases
   - a new valid food position is generated
3. The game ends only when:
   - the player quits
   - the snake dies
   - the board becomes full and no valid food cell remains

This turns Snake from a short demo into an actual repeatable play session.

## Design Direction

The selected approach is a hybrid food selection model:

- **runtime gameplay:** choose the next food position pseudo-randomly
- **tests:** inject deterministic food selection behavior

This avoids hard-wiring randomness into the domain while still making the live game feel like a real game.

## Architectural Boundaries

### Snake Domain

The domain is responsible for:

- deciding when new food is needed
- knowing which cells are occupied
- determining which cells are valid spawn candidates
- handling the “no space left” terminal condition

The domain must not depend directly on terminal code or UI code.

### Food Selection Strategy

The food selection strategy is responsible only for:

- choosing one valid cell from the available candidates

It must not decide game rules, movement, score, or rendering.

### Runtime / App Layer

The runtime or app layer is responsible for selecting the concrete strategy:

- pseudo-random strategy in real gameplay
- deterministic strategy in tests

## Scope Included

- new food appears after the current one is eaten
- food never spawns on the snake body
- tests prove deterministic multi-step continuation
- full-board condition is handled explicitly
- interactive `play snake` no longer exits after the first food

## Scope Excluded

- bonus food types
- timed food
- weighted spawn rules
- advanced difficulty scaling
- leaderboard or meta progression

This adjustment is intentionally narrow. It exists to make the first Snake loop honest and playable.

## Acceptance Criteria

This adjustment is done only if all of the following are true:

### Functional

- eating food causes new food to appear
- the game continues after eating
- food never overlaps the snake body
- full board/no-cell-left condition ends the game explicitly

### Testing

- tests can inject deterministic food selection
- multi-step Snake simulations remain deterministic
- existing Snake rule tests remain green

### Product

- `play snake` no longer exits immediately after the first food
- the player can continue playing across multiple food pickups

## Consequence for Planning

This should be treated as a small but real follow-up unit of work before claiming Snake is fully playable.

The next implementation plan revision should include:

- domain support for continuous food spawning
- injectable food selection strategy
- deterministic tests for repeated food consumption
- explicit full-board termination behavior
