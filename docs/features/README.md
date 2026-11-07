# Feature Docs

One folder per feature, named `YYYY-MM-DD-<feature>`. This is the single home for specs, designs, plans, and verification evidence.

## Conventions for new features

- Create `docs/features/<date>-<feature>/` before the first code change.
- Keep goal, scope, non-goals, checklist, acceptance criteria, and evidence in that folder.
- Do not use `openspec/`, `docs/superpowers/`, or `odd/tasks/`.

## Index

| Date | Feature | Contents |
| --- | --- | --- |
| 2026-05-04 | [foundation-and-snake](2026-05-04-foundation-and-snake/) | design, spec summary, plan, task 2 red evidence |
| 2026-05-04 | [snake-interactive](2026-05-04-snake-interactive/) | spec, plan |
| 2026-05-04 | [snake-playfield-layout](2026-05-04-snake-playfield-layout/) | spec, plan |
| 2026-05-04 | [snake-playfield-and-game-over](2026-05-04-snake-playfield-and-game-over/) | spec, plan |
| 2026-05-04 | [snake-continuous-food](2026-05-04-snake-continuous-food/) | spec, plan |
| 2026-05-04 | [snake-evidence](2026-05-04-snake-evidence/) | red/green evidence for the Snake interactive and playfield-layout work |
| 2026-05-07 | [sokoban](2026-05-07-sokoban/) | proposal, exploration, design, grid-puzzle spec, tasks, verify report |
| 2026-05-07 | [sokoban-improvements](2026-05-07-sokoban-improvements/) | spec, plan |
| 2026-05-08 | [sokoban-ui-improvements](2026-05-08-sokoban-ui-improvements/) | spec, plan |
| 2026-05-12 | [in-game-overlay-architecture](2026-05-12-in-game-overlay-architecture/) | design, plan |
| 2026-05-12 | [sokoban-hardening](2026-05-12-sokoban-hardening/) | design, plan |
| 2026-05-12 | [sokoban-run-mode](2026-05-12-sokoban-run-mode/) | design, plan |

## Historical notes

- Snake and later Sokoban docs were written with the Superpowers workflow; their plans still reference `superpowers:*` sub-skills, which are no longer installed. Treat those banners as historical.
- The first Sokoban feature was written with OpenSpec; `grid-puzzle-spec.md` is the former `openspec/specs/grid-puzzle/spec.md` (identical to the archived change delta).
- `snake-evidence` files are numbered by task, but their tests span both the interactive and playfield-layout plans, so they are kept together rather than guessed into one folder.
