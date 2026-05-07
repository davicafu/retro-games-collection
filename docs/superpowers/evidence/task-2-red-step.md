# Task 2 — Red Step Evidence (simulation screen render contract)

This evidence covers the Task 2 test:

`go test ./internal/platform/terminal/tcellscreen -run TestNewSimulationScreen_ImplementsRenderContract -v`

`internal/platform/terminal/tcellscreen/screen_test.go` is an existing shared test file (introduced earlier) and is extended here for Task 2.

## Red setup (isolated pre-implementation shape)

```bash
rm -rf "/tmp/opencode/task2-tcellscreen-red-proof"
mkdir -p "/tmp/opencode/task2-tcellscreen-red-proof/internal/platform/terminal/tcellscreen"
mkdir -p "/tmp/opencode/task2-tcellscreen-red-proof/internal/engine"
cp "go.mod" "/tmp/opencode/task2-tcellscreen-red-proof/go.mod"
cp "go.sum" "/tmp/opencode/task2-tcellscreen-red-proof/go.sum"
cp -r "internal/engine/input" "/tmp/opencode/task2-tcellscreen-red-proof/internal/engine/"
cp "internal/platform/terminal/tcellscreen/screen_test.go" "/tmp/opencode/task2-tcellscreen-red-proof/internal/platform/terminal/tcellscreen/screen_test.go"
```

## Red command

```bash
GOWORK=off go test ./internal/platform/terminal/tcellscreen -run TestNewSimulationScreen_ImplementsRenderContract -v
```

## Red output

```text
# github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/tcellscreen [github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/tcellscreen.test]
internal/platform/terminal/tcellscreen/screen_test.go:37:15: undefined: mapKeyEvent
internal/platform/terminal/tcellscreen/screen_test.go:54:7: undefined: newSimulationScreen
internal/platform/terminal/tcellscreen/screen_test.go:60:18: undefined: runeCell
internal/platform/terminal/tcellscreen/screen_test.go:67:19: undefined: runeCell
internal/platform/terminal/tcellscreen/screen_test.go:68:19: undefined: runeCell
internal/platform/terminal/tcellscreen/screen_test.go:69:19: undefined: runeCell
internal/platform/terminal/tcellscreen/screen_test.go:70:18: undefined: runeCell
FAIL	github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/tcellscreen [build failed]
FAIL
```

## Green confirmation on implemented code

```bash
go test ./internal/platform/terminal/tcellscreen -run TestNewSimulationScreen_ImplementsRenderContract -v
```

```text
=== RUN   TestNewSimulationScreen_ImplementsRenderContract
--- PASS: TestNewSimulationScreen_ImplementsRenderContract (0.00s)
PASS
ok  	github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/tcellscreen	(cached)
```
