package main

import (
	"fmt"
	"os"

	"github.com/davidbarreiro/cli-games-collection/cmd"
	"github.com/davidbarreiro/cli-games-collection/internal/app"
)

func main() {
	if err := cmd.Execute(
		app.RunShell,
		app.RunGame,
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
