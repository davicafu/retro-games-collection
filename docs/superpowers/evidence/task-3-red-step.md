# Task 3 — Red/Green Evidence (Snake playfield layout)

Task-specific strict-compliance evidence for:

- `TestModule_RunRendersAndQuitsGracefully`
- the projected-coordinate expectation update (`snapshot[5][2]` for head render)

## Red setup (old coordinate expectation in isolated proof copy)

```bash
rm -rf "/tmp/opencode/task3-playfield-layout-red-proof"
cp -a "/mnt/c/Users/david.barreiro/Downloads/cli-games-collection/.worktrees/snake-playfield-layout" "/tmp/opencode/task3-playfield-layout-red-proof"
perl -0pi -e 's/snapshot\[5\]\[2\]/snapshot[4][2]/' "/tmp/opencode/task3-playfield-layout-red-proof/internal/games/snake/module_test.go"
```

## Red command

```bash
GOWORK=off go test ./internal/games/snake -run TestModule_RunRendersAndQuitsGracefully -v
```

## Red output

```text
=== RUN   TestModule_RunRendersAndQuitsGracefully
    module_test.go:83: expected snake head rendered at projected playfield cell (2,5), got ' '
--- FAIL: TestModule_RunRendersAndQuitsGracefully (0.00s)
FAIL
FAIL	github.com/davidbarreiro/cli-games-collection/internal/games/snake	0.004s
FAIL
```

## Green confirmation on current Task 3 state

```bash
go test ./internal/games/snake -run TestModule_RunRendersAndQuitsGracefully -v
```

```text
=== RUN   TestModule_RunRendersAndQuitsGracefully
--- PASS: TestModule_RunRendersAndQuitsGracefully (0.00s)
PASS
ok  	github.com/davidbarreiro/cli-games-collection/internal/games/snake	(cached)
```
