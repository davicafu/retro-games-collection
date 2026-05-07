package cmd

import (
	"github.com/davidbarreiro/cli-games-collection/internal/app"
	"github.com/spf13/cobra"
)

func NewPlayCommand(runGame app.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "play [game]",
		Short: "Run a game directly",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGame(args[0])
		},
	}
}
