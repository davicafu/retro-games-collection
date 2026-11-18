# Snake Playfield Layout Adjustment

## Context

The first attempt at adding playfield borders and score chrome exposed a real design flaw: gameplay coordinates and render coordinates were being mixed together.

That approach caused two concrete problems:

1. visible borders consumed logical playfield space
2. score rendering became difficult to integrate honestly into the real runtime path

This proves the problem is not just cosmetic. The playfield layout needs an explicit rendering model.

## Updated Goal

Snake should render through a layout model that clearly separates:

- logical game coordinates
- visual playfield origin
- score/chrome rendering space

The goal is to keep the Snake domain unchanged while making the rendered playfield correct, legible, and extensible.

## Design Direction

The selected direction is:

- **logical board remains fixed and pure**
- **rendered surface becomes larger than the logical board**
- the renderer is responsible for mapping logical positions into the visual playfield region

This means:

- the game still thinks in pure board coordinates
- the renderer decides where the board appears inside the terminal surface

## Layout Model

### Logical Playfield

The logical board remains unchanged:

- `(0,0)` is the top-left playable cell
- `(width-1,height-1)` is still the bottom-right playable cell
- collision logic and game rules remain expressed in those coordinates

### Rendered Surface

The rendered surface includes extra space for chrome:

- one score row above the board
- one border row/column around the board

For example, a simple layout may be:

- row `0`: score text
- row `1`: top border
- row `2..n`: board interior rows
- final row: bottom border

and similarly for columns:

- column `0`: left border
- column `1..n`: board interior columns
- final column: right border

The exact constants can vary, but the principle must remain explicit.

## Architectural Boundaries

### Snake Domain

The domain remains responsible for:

- movement
- collisions
- score
- food
- death state

The domain must not know anything about visual chrome offsets or terminal layout.

### Snake Renderer

The renderer becomes responsible for:

- drawing score
- drawing border
- projecting logical cells into screen cells

Every game coordinate should pass through an explicit projection step instead of using ad-hoc magic offsets.

### Projection Responsibility

The renderer should conceptually provide a logical-to-screen mapping, such as:

- `toScreen(position)`

This keeps offset logic centralized and reusable.

## Scope Included

- score row rendered outside the logical board
- border rendered outside the logical board
- explicit logical-to-render projection for Snake board cells
- tests proving edge cells render inside the board region instead of on the border

## Scope Excluded

- game-over overlay behavior
- restart/menu/quit routing
- generalized viewport system for all games
- multiple HUD layers
- animated UI transitions

This is a rendering-layout correction, not a full overlay system.

## Acceptance Criteria

This adjustment is done only if all of the following are true:

### Functional

- the score renders outside the board interior
- the border renders outside the board interior
- the logical top-left and bottom-right playable cells both render inside the board area
- no playable cell overlaps border chrome

### Architectural

- Snake domain code remains unchanged for this adjustment
- offset logic is centralized in rendering/projection logic
- no scattered `+1/+2` style coordinate math across multiple call sites

### Testing

- tests verify border visibility
- tests verify score visibility
- tests verify correct projection of at least one edge cell and one bottom-right cell

## Consequence for Planning

The previous UX Task 2 should be replaced or amended so it no longer treats chrome as a simple offset tweak.

The next plan revision should implement:

- explicit layout/projection logic
- score + border rendering outside the logical board
- tests for edge-cell correctness
