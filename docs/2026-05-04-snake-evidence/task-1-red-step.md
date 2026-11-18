# Task 1 — Red/Green Evidence (Snake playfield layout model without render-contract change)

This artifact provides task-specific red/green proof for:

`TestDraw_PreservesCurrentRuntimeCoordinateContract`

## Red setup (projection applied too early)

```bash
rm -rf "/tmp/opencode/task1-playfield-layout-red-proof"
cp -a "/mnt/c/Users/david.barreiro/Downloads/cli-games-collection/.worktrees/snake-playfield-layout" "/tmp/opencode/task1-playfield-layout-red-proof"
# In the proof copy, render.go is switched back to Draw(...) using layout.toScreen(...) for body/food.
```

## Red command

```bash
GOWORK=off go test ./internal/games/snake -run TestDraw_PreservesCurrentRuntimeCoordinateContract -v
```

## Red output

```text
=== RUN   TestDraw_PreservesCurrentRuntimeCoordinateContract
    render_test.go:41: expected snake segment rendered at logical coordinates, got '\x00'
--- FAIL: TestDraw_PreservesCurrentRuntimeCoordinateContract (0.00s)
FAIL
FAIL	github.com/davidbarreiro/cli-games-collection/internal/games/snake	0.004s
FAIL
```

## Green confirmation on corrected Task 1 implementation

```bash
go test ./internal/games/snake -run TestDraw_PreservesCurrentRuntimeCoordinateContract -v
```

```text
=== RUN   TestDraw_PreservesCurrentRuntimeCoordinateContract
--- PASS: TestDraw_PreservesCurrentRuntimeCoordinateContract (0.00s)
PASS
ok  	github.com/davidbarreiro/cli-games-collection/internal/games/snake	0.005s
```
