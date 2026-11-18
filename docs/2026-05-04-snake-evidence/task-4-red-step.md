# Task 4 (redo) — Red Step Evidence

This redo targets wiring from launcher selection to real gameplay execution.

## Red command

`go test ./internal/app -run 'TestRunGameWith_DelegatesSnakeRunner|TestRunGameWith_ReturnsErrorForUnsupportedGame|TestBuildLauncher_BridgesSelectionToRunner|TestLauncherGames_OnlyListsImplementedGames' -v`

## Red output

```text
# github.com/davidbarreiro/cli-games-collection/internal/app [github.com/davidbarreiro/cli-games-collection/internal/app.test]
internal/app/app_test.go:11:9: undefined: RunGameWith
internal/app/app_test.go:25:9: undefined: RunGameWith
internal/app/app_test.go:38:14: undefined: BuildLauncher
internal/app/app_test.go:56:11: undefined: LauncherGames
FAIL	github.com/davidbarreiro/cli-games-collection/internal/app [build failed]
FAIL
```
