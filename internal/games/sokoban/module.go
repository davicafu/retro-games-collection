package sokoban

import (
	"context"
	"fmt"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	gameoverlay "github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"
	"github.com/mattn/go-runewidth"
)

const (
	restartPenaltyIncrement = 5 * time.Second
	nextLevelSkipPenalty    = 25 * time.Second
)

const (
	pauseOptionIDContinue  = "continue"
	pauseOptionIDRestart   = "restart-level"
	pauseOptionIDNextLevel = "next-level"
	pauseOptionIDQuit      = "back-to-menu"

	summaryOptionIDBackToMenu = "back-to-menu"
)

var finalSummaryInitialDelay = 5 * time.Second

type levelResultKind int

const (
	levelResultUnknown levelResultKind = iota
	levelResultCompleted
	levelResultRestart
	levelResultNextLevel
	levelResultQuitToMenu
)

type levelResult struct {
	kind    levelResultKind
	elapsed time.Duration
}

type completedLevel struct {
	levelIdx int
	elapsed  time.Duration
}

type runSession struct {
	currentLevelIdx int
	remainingUndos  int
	active          bool
	completedLevels []completedLevel
	totalElapsed    time.Duration
	totalLevels     int
}

type runResult string

const (
	runResultCompleted runResult = "completada"
	runResultAbandoned runResult = "abandonada"
)

type runSummary struct {
	result          runResult
	totalTime       time.Duration
	levelsCompleted int
	levelReached    int
	undosUsed       int
	undosRemaining  int
	levelTimes      []time.Duration
}

func newRunSession(totalLevels int) *runSession {
	return &runSession{
		currentLevelIdx: 0,
		remainingUndos:  4,
		active:          totalLevels > 0,
		completedLevels: make([]completedLevel, 0, totalLevels),
		totalLevels:     totalLevels,
	}
}

func (s *runSession) completeLevel(elapsed time.Duration) {
	if !s.active {
		return
	}
	if s.currentLevelIdx >= s.totalLevels {
		s.active = false
		return
	}

	s.completedLevels = append(s.completedLevels, completedLevel{levelIdx: s.currentLevelIdx, elapsed: elapsed})
	s.totalElapsed += elapsed
	s.currentLevelIdx++

	if s.currentLevelIdx >= s.totalLevels {
		s.active = false
	}
}

func (s *runSession) isActive() bool {
	return s.active
}

func (s *runSession) currentLevel() int {
	return s.currentLevelIdx
}

func (s *runSession) consumeUndo() bool {
	if s.remainingUndos == 0 {
		return false
	}
	s.remainingUndos--
	return true
}

func (s *runSession) levelTimes() []time.Duration {
	times := make([]time.Duration, 0, len(s.completedLevels))
	for _, completed := range s.completedLevels {
		times = append(times, completed.elapsed)
	}
	return times
}

func (s *runSession) summary(result runResult) runSummary {
	levelReached := s.currentLevelIdx + 1
	if levelReached > s.totalLevels {
		levelReached = s.totalLevels
	}
	if levelReached < 1 && s.totalLevels > 0 {
		levelReached = 1
	}

	return runSummary{
		result:          result,
		totalTime:       s.totalElapsed,
		levelsCompleted: len(s.completedLevels),
		levelReached:    levelReached,
		undosUsed:       4 - s.remainingUndos,
		undosRemaining:  s.remainingUndos,
		levelTimes:      s.levelTimes(),
	}
}

var (
	runLevelFn = func(m Module, ctx context.Context, game *Game, levelIdx int, initialPenalty time.Duration, session *runSession) (levelResult, error) {
		return m.runLevel(ctx, game, levelIdx, initialPenalty, session)
	}
	showLevelCompleteFn = func(m Module, ctx context.Context, levelIdx int, elapsed time.Duration) error {
		return m.showLevelComplete(ctx, levelIdx, elapsed)
	}
	showFinalSummaryFn = func(m Module, ctx context.Context, summary runSummary) error {
		return m.showFinalSummary(ctx, summary)
	}
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
	session := newRunSession(len(Levels))

	for session.isActive() {
		levelIdx := session.currentLevel()
		restartPenalty := time.Duration(0)

		for {
			game, err := ParseLevelString(Levels[levelIdx])
			if err != nil {
				return fmt.Errorf("failed to load level %d: %w", levelIdx, err)
			}

			result, err := runLevelFn(m, ctx, game, levelIdx, restartPenalty, session)
			if err != nil {
				return err
			}

			switch result.kind {
			case levelResultQuitToMenu:
				return nil
			case levelResultCompleted:
				session.completeLevel(result.elapsed)
				if err := showLevelCompleteFn(m, ctx, levelIdx, result.elapsed); err != nil {
					return err
				}
			case levelResultRestart:
				restartPenalty += restartPenaltyIncrement
				continue
			case levelResultNextLevel:
				session.completeLevel(result.elapsed)
			default:
				return fmt.Errorf("unknown level result kind %d", result.kind)
			}

			break
		}
	}

	return showFinalSummaryFn(m, ctx, session.summary(runResultCompleted))
}

// runLevel runs a single level until the player wins or quits.
// Returns explicit level exit semantics and elapsed time when completed.
func (m Module) runLevel(ctx context.Context, game *Game, levelIdx int, initialPenalty time.Duration, session *runSession) (levelResult, error) {
	start := time.Now()
	elapsedWithPenalty := initialPenalty

	m.drawLevel(game, levelIdx, elapsedWithPenalty, session)

	if err := m.Screen.Present(); err != nil {
		return levelResult{}, err
	}

	for {
		select {
		case <-ctx.Done():
			return levelResult{}, ctx.Err()
		default:
		}

		frame := m.Provider.Poll()

		for _, action := range frame.Actions {
			switch action {
			case input.ActionUp, input.ActionDown, input.ActionLeft, input.ActionRight:
				game.Move(action)
				m.drawLevel(game, levelIdx, time.Since(start)+initialPenalty, session)

			case input.ActionUndo:
				if session.remainingUndos > 0 && game.Undo() {
					session.consumeUndo()
				}
				m.drawLevel(game, levelIdx, time.Since(start)+initialPenalty, session)

			case input.ActionHelp:
				if err := m.showHelpOverlay(ctx, game, levelIdx, time.Since(start)+initialPenalty, session); err != nil {
					return levelResult{}, err
				}

			case input.ActionQuit:
				return levelResult{kind: levelResultQuitToMenu}, nil

			case input.ActionPause:
				pausedAt := time.Now()
				result, err := m.showPauseMenu(ctx, levelIdx)
				if err != nil {
					return levelResult{}, err
				}
				start = start.Add(time.Since(pausedAt))
				switch result {
				case pauseContinue:
					m.drawLevel(game, levelIdx, time.Since(start)+initialPenalty, session)
				case pauseRestart:
					return levelResult{kind: levelResultRestart}, nil
				case pauseNext:
					return levelResult{kind: levelResultNextLevel, elapsed: time.Since(start) + initialPenalty + nextLevelSkipPenalty}, nil
				case pauseQuit:
					return levelResult{kind: levelResultQuitToMenu}, nil
				}
			}
		}

		if game.Won() {
			elapsed := time.Since(start) + initialPenalty
			return levelResult{kind: levelResultCompleted, elapsed: elapsed}, nil
		}

		if err := m.Screen.Present(); err != nil {
			return levelResult{}, err
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// drawLevel renders the game state with level info and elapsed time at the top.
func (m Module) drawLevel(game *Game, levelIdx int, elapsed time.Duration, session *runSession) {
	m.Screen.Clear()

	// Header: Level, Time, Undo
	undoStatus := session.remainingUndos > 0
	var undoRune rune
	if undoStatus {
		undoRune = '↶'
	} else {
		undoRune = ' '
	}
	header := fmt.Sprintf("L%d/%d ⏱%s U:%d %c", levelIdx+1, len(Levels), fmtDuration(elapsed), session.remainingUndos, undoRune)
	render.WriteText(m.Screen, 0, 0, header)

	// Draw the game board offset by 1 row down
	DrawAt(m.Screen, game, 0, 1)

	// Dibujar indicadores en la parte inferior
	hints := " [↑↓←→] Mover  [U] Deshacer  [P] Pausa  [?] Ayuda"
	render.WriteText(m.Screen, 0, 22, hints)
}

// showLevelComplete displays a completion message with elapsed time.
func (m Module) showLevelComplete(ctx context.Context, levelIdx int, elapsed time.Duration) error {
	m.Screen.Clear()
	gameoverlay.Draw(m.Screen, m.buildLevelCompleteOverlay(levelIdx, elapsed))

	if err := m.Screen.Present(); err != nil {
		return err
	}

	return sleepWithContext(ctx, 2*time.Second)
}

// showFinalSummary displays a summary table with all level times and total as overlay.
func (m Module) showFinalSummary(ctx context.Context, summary runSummary) error {
	model := m.buildFinalSummaryOverlay(summary)
	showSummaryOnce := false
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m.Screen.Clear()
		gameoverlay.Draw(m.Screen, model)
		if err := m.Screen.Present(); err != nil {
			return err
		}

		if !showSummaryOnce {
			showSummaryOnce = true
			if err := sleepWithContext(ctx, finalSummaryInitialDelay); err != nil {
				return err
			}
		}

		frame := m.Provider.Poll()
		for _, action := range frame.Actions {
			switch action {
			case input.ActionSelect:
				result := model.Apply(gameoverlay.IntentConfirm)
				if result.Kind == gameoverlay.ResultKindConfirmed && result.OptionID == summaryOptionIDBackToMenu {
					return nil
				}
			case input.ActionQuit, input.ActionPause:
				return nil
			}
		}

		if err := sleepWithContext(ctx, 50*time.Millisecond); err != nil {
			return err
		}
	}
}

// showHelpOverlay displays the help screen as overlay.
func (m Module) showHelpOverlay(ctx context.Context, game *Game, levelIdx int, elapsed time.Duration, session *runSession) error {
	m.Screen.Clear()
	gameoverlay.Draw(m.Screen, m.buildHelpOverlay())
	if err := m.Screen.Present(); err != nil {
		return err
	}

	// Wait for any key to dismiss
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		frame := m.Provider.Poll()
		if len(frame.Actions) > 0 {
			break
		}
		if err := sleepWithContext(ctx, 50*time.Millisecond); err != nil {
			return err
		}
	}

	// Redraw game
	m.drawLevel(game, levelIdx, elapsed, session)
	if err := m.Screen.Present(); err != nil {
		return err
	}

	return nil
}

type pauseResult int

const (
	pauseContinue pauseResult = iota
	pauseRestart
	pauseNext
	pauseQuit
)

// showPauseMenu displays the pause menu using the overlay system.
func (m Module) showPauseMenu(ctx context.Context, levelIdx int) (pauseResult, error) {
	model := m.buildPauseOverlay(levelIdx).WithDismissPolicy(gameoverlay.DismissPolicyClose)

	m.Screen.Clear()
	gameoverlay.Draw(m.Screen, model)
	if err := m.Screen.Present(); err != nil {
		return pauseContinue, err
	}

	for {
		select {
		case <-ctx.Done():
			return pauseContinue, ctx.Err()
		default:
		}

		frame := m.Provider.Poll()

		for _, action := range frame.Actions {
			switch action {
			case input.ActionUp:
				model.Apply(gameoverlay.IntentMoveUp)
			case input.ActionDown:
				model.Apply(gameoverlay.IntentMoveDown)
			case input.ActionSelect:
				return m.resolvePauseOverlayResult(model.Apply(gameoverlay.IntentConfirm), levelIdx), nil
			case input.ActionPause:
				return m.resolvePauseOverlayResult(model.Apply(gameoverlay.IntentClose), levelIdx), nil
			case input.ActionQuit:
				return pauseQuit, nil
			}
		}

		m.Screen.Clear()
		gameoverlay.Draw(m.Screen, model)
		if err := m.Screen.Present(); err != nil {
			return pauseContinue, err
		}
		if err := sleepWithContext(ctx, 50*time.Millisecond); err != nil {
			return pauseContinue, err
		}
	}
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (m Module) buildPauseOverlay(levelIdx int) gameoverlay.Model {
	options := []gameoverlay.Option{
		{ID: pauseOptionIDContinue, Label: "Continuar", Enabled: true},
		{ID: pauseOptionIDRestart, Label: "Reiniciar (+5s)", Enabled: true},
	}
	if levelIdx < len(Levels)-1 {
		options = append(options, gameoverlay.Option{ID: pauseOptionIDNextLevel, Label: "Pasar nivel (+25s)", Enabled: true})
	}
	options = append(options, gameoverlay.Option{ID: pauseOptionIDQuit, Label: "Abandonar partida", Enabled: true})

	return gameoverlay.NewAction("PAUSA", []gameoverlay.Section{gameoverlay.Text("Juego pausado")}, options)
}

func (m Module) buildHelpOverlay() gameoverlay.Model {
	return gameoverlay.NewInfo("AYUDA", []gameoverlay.Section{
		gameoverlay.Text("Objetivo: llevar cajas a objetivos (.)"),
		gameoverlay.Text("WASD/Flechas: Mover", "U: Deshacer (4 por partida)", "P/ESC: Menú de pausa", "?/h: Esta ayuda", "Q: Salir al menú"),
		gameoverlay.Text("Presiona cualquier tecla..."),
	})
}

func (m Module) buildLevelCompleteOverlay(levelIdx int, elapsed time.Duration) gameoverlay.Model {
	title := fmt.Sprintf("Level %d complete!", levelIdx+1)
	subtitle := "All levels complete!"
	if levelIdx < len(Levels)-1 {
		subtitle = "Next level..."
	}

	return gameoverlay.NewInfo(
		title,
		[]gameoverlay.Section{gameoverlay.Text(fmt.Sprintf("Time: %s", fmtDuration(elapsed)), subtitle)},
	)
}

func (m Module) buildFinalSummaryOverlay(summary runSummary) gameoverlay.Model {
	lines := make([]string, 0, len(summary.levelTimes)+5)
	lines = append(lines,
		fmt.Sprintf("Resultado: %s", summary.result),
		fmt.Sprintf("Tiempo total: %s", fmtDuration(summary.totalTime)),
		fmt.Sprintf("Niveles completados: %d/%d", summary.levelsCompleted, len(Levels)),
		fmt.Sprintf("Undos usados: %d", summary.undosUsed),
		fmt.Sprintf("Undos restantes: %d", summary.undosRemaining),
	)
	for i, t := range summary.levelTimes {
		lines = append(lines, fmt.Sprintf("Nivel %d: %s", i+1, fmtDuration(t)))
	}
	lines = append(lines, fmt.Sprintf("TOTAL: %s", fmtDuration(summary.totalTime)))
	maxWidth := 0
	for _, line := range lines {
		if width := runewidth.StringWidth(line); width > maxWidth {
			maxWidth = width
		}
	}
	for i := range lines {
		lines[i] = padRightDisplay(lines[i], maxWidth)
	}

	return gameoverlay.NewAction(
		"¡¡FIN DEL JUEGO!!",
		[]gameoverlay.Section{gameoverlay.Text(lines...)},
		[]gameoverlay.Option{{ID: summaryOptionIDBackToMenu, Label: "Volver al menú", Enabled: true}},
	)
}
func padRightDisplay(s string, width int) string {
	current := runewidth.StringWidth(s)
	if current >= width {
		return s
	}
	return s + fmt.Sprintf("%*s", width-current, "")
}

func (m Module) resolvePauseOverlayResult(result gameoverlay.Result, _ int) pauseResult {
	if result.Kind != gameoverlay.ResultKindConfirmed {
		return pauseContinue
	}

	switch result.OptionID {
	case pauseOptionIDContinue:
		return pauseContinue
	case pauseOptionIDRestart:
		return pauseRestart
	case pauseOptionIDNextLevel:
		return pauseNext
	case pauseOptionIDQuit:
		return pauseQuit
	default:
		return pauseContinue
	}
}

// fmtDuration formats a duration as MM:SS.mmm.
func fmtDuration(d time.Duration) string {
	total := d.Milliseconds()
	min := total / 60000
	sec := (total % 60000) / 1000
	millis := total % 1000
	return fmt.Sprintf("%02d:%02d.%03d", min, sec, millis)
}
