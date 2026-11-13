# Packet (game 8)

Objective: eighth game of the roadmap, "Metro Survival": lane systems and rapid decision gameplay on the fixed-tick runtime.
Branch: to be cut from `feat/leak` once Leak is committed (Leak is still uncommitted in the same tree; Packet touches its own package plus the launcher list, README and features index) · Delivery: `ask-on-risk`, commits only when the user asks · Test runner: `go test ./...`

## Specs

Source requirements: README roadmap row `| 8 | Metro Survival | Lane systems, rapid decision gameplay |` and the user's theme choice "Packet in transit (Recommended)" (L2), whose description was: "Eres un paquete de red que viaja por 3-4 enlaces de fibra. Cambias de enlace para esquivar firewalls y colisiones, y recoges ancho de banda. Sigue la línea tech de Drone, Breach, Botnet y Leak." Everything marked *(assumption)* is a proposed design default, not a user requirement.

- **S1 — Theme.** "Eres un paquete de red que viaja por 3-4 enlaces de fibra. Cambias de enlace para esquivar firewalls y colisiones, y recoges ancho de banda." Original game: no names, characters, tracks, sprites or sounds from any existing lane runner.
- **S2 — Links.** *(assumption)* 4 horizontal fiber links (lanes) scroll right to left past the packet, which sits at a fixed column. Each link is drawn 3 rows high; dashed separators between links scroll with the world.
- **S3 — Switching.** *(assumption)* Up and down move the packet one link per press, at once, never past the outer links.
- **S4 — Firewalls.** *(assumption)* Firewalls come in gates: one gate blocks 1 to 3 of the 4 links (never all) over a length of 1 to 10 cells. Touching a firewall, by scrolling into it or by switching into it, drops the packet and ends the round.
- **S5 — Fair gates.** *(assumption)* Every gate is passable from wherever the previous one let the packet be: from every free link of a gate, a free link of the next one is at most `reach` links away, where `reach` is the gap between them divided by the cells the world scrolls during a reaction window (0.3 s), at least 1, at the speed the gate will arrive with. Gates are at least one reaction window apart.
- **S6 — Bandwidth.** *(assumption)* Some gates carry a bandwidth token in one of their free links; taking it scores 50.
- **S7 — Speed and score.** *(assumption)* The world scrolls at 15 cells/s and speeds up with distance to 40 cells/s. Score is the distance in cells plus the tokens; the HUD shows score and speed.
- **S8 — Viability.** User on Botnet (2026-11-11): "tienen que ser viables". *(assumption: applied to Packet too)* An autopilot that only acts once per reaction window, one switch at a time, survives 3 minutes (past top speed) in every one of 30 seeded games, pinned by a test.
- **S9 — Controls and flow.** *(assumption)* Arrows/WS switch links. Pause, Game Over (score, Restart/Menu) and Q come from `internal/engine/session`. End title "PACKET DROPPED".
- **S10 — Launcher.** *(assumption)* "Packet" appears in the launcher with its controls hint, and `play packet` starts it.
- **S11 — Engine.** Scrolling and the speed ramp reuse `internal/engine/scroll`. Lanes stay in the game package: no other game uses them yet (AGENTS.md: promote a pattern once it repeats).
- **S12 — Determinism and timing.** Per AGENTS.md: injectable clock and RNG, scripted input, headless simulation tests; no sleep-based tests. 60 ticks per second like Drone, Breach, Botnet and Leak, with real-time speeds pinned by a test.

## Tasks

| ID | Specs | Route | Status | Evidence |
| --- | --- | --- | --- | --- |
| T1 | S2, S3, S4, S7, S12 | inline | done (uncommitted) | code written before its tests (no RED observed); mutants killed instead: no clamp, firewalls dropped too early, no check right after a switch (survived first, killed by `TestSwitchingOntoTheLastCellOfAFirewallDropsThePacket`) |
| T2 | S5, S6, S8 | inline | done (uncommitted) | `spawn_test.go` properties over 20 seeds; viability 30 of 30 games survive 3 min at top speed; mutants killed: no fairness, reach ×3, gap without the reaction window (each drops all 30 autopilot games), token placement |
| T3 | S1, S2, S7 (HUD), rendering | inline | done (uncommitted) | `render_test.go`; 80x24 preview checked |
| T4 | S9 | inline | done (uncommitted) | `module_test.go`: last press wins, end title, summary, run and quit |
| T5 | S10 | inline | done (uncommitted) | RED: 8 games, `play packet`, controls hint → GREEN; README run command and features index row |

Routing: each unit builds on the previous one; no independent parallel units, so no writer delegation. Forecast: about 900 authored changed lines; commits wait for the user's explicit request (user rule).

## Log

- **L1** (2026-11-13, user): "mas tarede, continua con el siguietne mientras pruebo"
- **L2** (2026-11-13, user answer): theme = "Packet in transit (Recommended)".
- **L3**: Design defaults (all S#-marked assumptions): 4 links, instant one-link switch, one hit ends the round (like Runner and Drone), 15 to 40 cells/s, reaction window 0.3 s, token 50 points, end title "PACKET DROPPED".
- **L4**: Fairness (S5) is enforced by construction: the spawner picks the free links of each gate among every set that is fair for the previous gate, so no rejection loop; the speed used is the one the world will have when the gate reaches the packet (the ramp only rises). The autopilot plans one decision per reaction window at the next speed step and moves at most one link per decision.
- **L5**: The world starts empty: the first gate appears at the right edge and reaches the packet after about 4 s, a lead-in to settle on a link. Checks: `go vet ./...` clean, `go test -race ./...` fully green. Not committed: commits wait for the user's request; Leak (still uncommitted) shares the launcher list, its test, README and the features index, so those hunks need splitting per game when committing. Not play-tested in a real terminal.
- **L6** (2026-11-13, user): "la velocidad podria cambiar un poquito mas rapido? no mucho mas". Speed steps now every 90 cells instead of 120: the top speed comes after about 76 s instead of 101 s (a step every 6 s at the start instead of 8 s); base and top speeds unchanged. `go test -race ./internal/games/packet/` green, viability still 30 of 30.
- **L7** (2026-11-13, user): "para el packet la velocidad tiene que subir mas rapido, es demasiado facil aun asi". Speed steps of 3 points every 60 cells (were 2 every 90): top speed after about 34 s instead of 76 s, a step every 4 s at the start; base 15 and top about 40 cells/s unchanged. `go test -race ./internal/games/packet/` green, viability still 30 of 30.
- **L8**: With the faster ramp the viability test first dropped seed 16. Traces showed the autopilot's model, not the gates: it sized each reaction window at the speed at its start, and a speed step inside a window scrolled one cell more than planned, so a planned switch came a cell late. Windows are now sized at the speed by their end. Over 300 seeds no game drops (before the fix, 5 of 300 did); unfair spawners (no fairness, reach ×3, no reaction gap) still drop all 30 test games. `go vet ./...` clean, `go test -race ./...` fully green.
- **L9** (2026-11-13, user): "el packet deberia llegar a mas velocidad, no pararse en 2.6". Top speed 96 points per tick (57.6 cells/s, x3.84 on the HUD) instead of 66 (39.6 cells/s, x2.6); it stays below one cell per tick, the limit that keeps firewalls from passing through the packet. With the same ramp the top comes after about 47 s. Viability still 30 of 30 (3 minutes, past top speed); `go test -race ./internal/games/packet/` green.
- **L10** (2026-11-13, user): "para packet haz el campos si tal no tan ancho entonces para tener un como menos de prevision". Playfield 44 columns instead of 64: firewalls are seen 38 cells ahead of the packet instead of 58, 2.5 s at the start speed and 0.66 s at top speed (were 3.9 s and 1 s). Viability still 30 of 30.
