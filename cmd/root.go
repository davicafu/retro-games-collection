package cmd

import (
	"github.com/davidbarreiro/cli-games-collection/internal/app"
	"github.com/spf13/cobra"
)

func NewRootCommand(runLauncher app.Launcher, runGame app.Runner) *cobra.Command {
	safeRunner := runGame
	if safeRunner == nil {
		safeRunner = app.RunGame
	}

	cmd := &cobra.Command{
		Use:           "retro-games",
		Short:         "Progressive terminal games laboratory",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if runLauncher == nil {
				return nil
			}
			return runLauncher()
		},
	}

	cmd.AddCommand(NewPlayCommand(safeRunner))
	return cmd
}

func Execute(runLauncher app.Launcher, runGame app.Runner) error {
	return NewRootCommand(runLauncher, runGame).Execute()
}
