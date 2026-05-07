# TDD RED Evidence — Task 2

- Step: RED (test written before implementation)
- Command:

```bash
go test ./internal/engine/runtime -run TestRuntime_RunAdvancesUntilGameQuits -v
```

- Observed output:

```text
# github.com/davidbarreiro/cli-games-collection/internal/engine/runtime
internal/engine/runtime/runtime_test.go:8:2: no required module provides package github.com/davidbarreiro/cli-games-collection/internal/engine/clock; to add it:
	go get github.com/davidbarreiro/cli-games-collection/internal/engine/clock
FAIL	github.com/davidbarreiro/cli-games-collection/internal/engine/runtime [setup failed]
FAIL
```
