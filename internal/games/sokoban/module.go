package sokoban

import (
	"context"
	"fmt"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// Module is the runnable entry point for Sokoban.
type Module struct {
	Screen   render.Screen
	Provider input.Provider
}

// NewModule creates a new Sokoban module.
func NewModule(screen render.Screen, provider input.Provider) Module {
	return Module{
		Screen:   screen,
		Provider: provider,
	}
}

// Run starts the Sokoban game loop.
// It loads levels sequentially and advances on win.
func (m Module) Run(ctx context.Context) error {
	for levelIdx := 0; levelIdx < len(Levels); levelIdx++ {
		game, err := ParseLevelString(Levels[levelIdx])
		if err != nil {
			return fmt.Errorf("failed to load level %d: %w", levelIdx, err)
		}

		// Game loop for this level with elapsed timer
		elapsed, err := m.runLevel(ctx, game, levelIdx)
		if err != nil {
			return err
		}

		if err := m.showLevelComplete(levelIdx, elapsed); err != nil {
			return err
		}
	}

	return nil
}

// runLevel runs a single level until the player wins or quits.
// Returns the elapsed time on win, or an error if the user quit.
func (m Module) runLevel(ctx context.Context, game *Game, levelIdx int) (time.Duration, error) {
	start := time.Now()
	m.drawLevel(game, levelIdx, 0)

	if err := m.Screen.Present(); err != nil {
		return 0, err
	}

	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		frame := m.Provider.Poll()

		for _, action := range frame.Actions {
			switch action {
			case input.ActionUp, input.ActionDown, input.ActionLeft, input.ActionRight:
				game.Move(action)
				m.drawLevel(game, levelIdx, time.Since(start))

			case input.ActionUndo:
				game.Undo()
				m.drawLevel(game, levelIdx, time.Since(start))

			case input.ActionQuit, input.ActionPause:
				return 0, nil
			}
		}

		if game.Won() {
			elapsed := time.Since(start)
			return elapsed, nil
		}

		if err := m.Screen.Present(); err != nil {
			return 0, err
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// drawLevel renders the game state with level info and elapsed time at the top.
func (m Module) drawLevel(game *Game, levelIdx int, elapsed time.Duration) {
	m.Screen.Clear()

	header := fmt.Sprintf("Level %d/%d   ⏱ %s", levelIdx+1, len(Levels), fmtDuration(elapsed))
	render.WriteText(m.Screen, 0, 0, header)

	// Draw the game board offset by 1 row down
	DrawAt(m.Screen, game, 0, 1)
}

// showLevelComplete displays a completion message with elapsed time.
func (m Module) showLevelComplete(levelIdx int, elapsed time.Duration) error {
	m.Screen.Clear()

	title := fmt.Sprintf("Level %d complete!", levelIdx+1)
	timeStr := fmt.Sprintf("Time: %s", fmtDuration(elapsed))
	var subtitle string
	if levelIdx < len(Levels)-1 {
		subtitle = "Next level..."
	} else {
		subtitle = "All levels complete!"
	}

	// Calculate box dimensions (3 content rows: title, time, subtitle)
	boxW := 28
	boxH := 3
	sw, sh := m.Screen.Size()
	boxX := (sw - boxW - 2) / 2
	boxY := (sh - boxH - 2) / 2

	render.DrawBox(m.Screen, boxX, boxY, boxW, boxH)

	render.WriteText(m.Screen, boxX+1+(boxW-len(title))/2, boxY+1, title)
	render.WriteText(m.Screen, boxX+1+(boxW-len(timeStr))/2, boxY+2, timeStr)
	render.WriteText(m.Screen, boxX+1+(boxW-len(subtitle))/2, boxY+3, subtitle)

	if err := m.Screen.Present(); err != nil {
		return err
	}

	time.Sleep(2 * time.Second)
	return nil
}

// fmtDuration formats a duration as MM:SS.mmm.
func fmtDuration(d time.Duration) string {
	total := d.Milliseconds()
	min := total / 60000
	sec := (total % 60000) / 1000
	millis := total % 1000
	return fmt.Sprintf("%02d:%02d.%03d", min, sec, millis)
}
