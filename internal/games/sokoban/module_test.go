package sokoban

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/records"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/ui/highscore"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"
	"github.com/mattn/go-runewidth"
)

func TestBuildPauseOverlayOffersRestartAndNextLevelBeforeLastLevel(t *testing.T) {
	module := Module{}
	model := module.buildPauseOverlay(0)

	if got := model.Kind(); got != overlay.KindAction {
		t.Fatalf("expected action overlay, got %q", got)
	}

	optionIDs := optionIDs(model.Options())
	if len(optionIDs) != 4 {
		t.Fatalf("expected four pause options before last level, got %v", optionIDs)
	}
	if !contains(optionIDs, pauseOptionIDContinue) {
		t.Fatalf("expected continue option in pause overlay, got %v", optionIDs)
	}
	if !contains(optionIDs, pauseOptionIDRestart) {
		t.Fatalf("expected restart option in pause overlay, got %v", optionIDs)
	}
	if !contains(optionIDs, pauseOptionIDNextLevel) {
		t.Fatalf("expected next-level option before last level, got %v", optionIDs)
	}
	if !contains(optionIDs, pauseOptionIDQuit) {
		t.Fatalf("expected quit option in pause overlay, got %v", optionIDs)
	}
}

func TestNewRunSessionStartsAtLevelZeroWithFourUndosAndActiveStatus(t *testing.T) {
	session := newRunSession(len(Levels))

	if session.currentLevelIdx != 0 {
		t.Fatalf("expected run session to start at level index 0, got %d", session.currentLevelIdx)
	}
	if session.remainingUndos != 4 {
		t.Fatalf("expected run session to start with 4 total undos, got %d", session.remainingUndos)
	}
	if !session.active {
		t.Fatal("expected run session to start as active")
	}
	if len(session.completedLevels) != 0 {
		t.Fatalf("expected no completed levels at session start, got %d", len(session.completedLevels))
	}
	if session.totalElapsed != 0 {
		t.Fatalf("expected zero total elapsed at session start, got %v", session.totalElapsed)
	}
}

func TestRunSessionCompleteLevelAdvancesAndRecordsProgress(t *testing.T) {
	session := newRunSession(len(Levels))
	elapsed := 7*time.Second + 350*time.Millisecond

	session.completeLevel(elapsed)

	if session.currentLevelIdx != 1 {
		t.Fatalf("expected next level index to be 1 after completing first level, got %d", session.currentLevelIdx)
	}
	if len(session.completedLevels) != 1 {
		t.Fatalf("expected one completed level entry, got %d", len(session.completedLevels))
	}
	if session.completedLevels[0].levelIdx != 0 {
		t.Fatalf("expected completed level record for level 0, got %d", session.completedLevels[0].levelIdx)
	}
	if session.completedLevels[0].elapsed != elapsed {
		t.Fatalf("expected completed level elapsed %v, got %v", elapsed, session.completedLevels[0].elapsed)
	}
	if session.totalElapsed != elapsed {
		t.Fatalf("expected total elapsed %v, got %v", elapsed, session.totalElapsed)
	}
}

func TestRunSessionCompletingLastLevelDeactivatesAtBoundary(t *testing.T) {
	session := newRunSession(2)
	firstElapsed := 3*time.Second + 100*time.Millisecond
	lastElapsed := 4*time.Second + 250*time.Millisecond

	session.completeLevel(firstElapsed)
	if !session.active {
		t.Fatal("expected session to remain active before last level completion")
	}

	session.completeLevel(lastElapsed)

	if session.active {
		t.Fatal("expected session to deactivate after completing last level")
	}
	if session.currentLevelIdx != 2 {
		t.Fatalf("expected current level index to stop at completion boundary 2, got %d", session.currentLevelIdx)
	}
	if len(session.completedLevels) != 2 {
		t.Fatalf("expected exactly two completed levels, got %d", len(session.completedLevels))
	}
}

func TestRunSessionTotalElapsedMatchesRecordedCompletedLevels(t *testing.T) {
	session := newRunSession(3)
	firstElapsed := 2*time.Second + 150*time.Millisecond
	secondElapsed := 5*time.Second + 900*time.Millisecond

	session.completeLevel(firstElapsed)
	session.completeLevel(secondElapsed)

	var expectedTotal time.Duration
	for _, completed := range session.completedLevels {
		expectedTotal += completed.elapsed
	}

	if session.totalElapsed != expectedTotal {
		t.Fatalf("expected total elapsed %v to match sum of completed levels, got %v", expectedTotal, session.totalElapsed)
	}
}

func TestRunSessionConsumeUndoUsesCanonicalRemainingPool(t *testing.T) {
	session := newRunSession(len(Levels))

	for i := 0; i < 4; i++ {
		if ok := session.consumeUndo(); !ok {
			t.Fatalf("expected undo #%d consumption to succeed", i+1)
		}
	}

	if ok := session.consumeUndo(); ok {
		t.Fatal("expected consuming undo beyond pool to fail")
	}
	if session.remainingUndos != 0 {
		t.Fatalf("expected no remaining undos after consuming full pool, got %d", session.remainingUndos)
	}
}

func TestRunSessionUndoPoolPersistsAcrossLevels(t *testing.T) {
	session := newRunSession(len(Levels))

	if ok := session.consumeUndo(); !ok {
		t.Fatal("expected first undo consumption to succeed")
	}
	if ok := session.consumeUndo(); !ok {
		t.Fatal("expected second undo consumption to succeed")
	}

	session.completeLevel(5 * time.Second)

	if session.remainingUndos != 2 {
		t.Fatalf("expected remaining undos to persist after level completion, got %d", session.remainingUndos)
	}
	if ok := session.consumeUndo(); !ok {
		t.Fatal("expected undo to remain available after advancing level")
	}
	if session.remainingUndos != 1 {
		t.Fatalf("expected remaining undos to continue from global pool, got %d", session.remainingUndos)
	}
}

func TestRunSessionUndoConsumesGlobalPool(t *testing.T) {
	TestRunSessionConsumeUndoUsesCanonicalRemainingPool(t)
}

func TestBuildPauseOverlayOmitsNextLevelOnLastLevel(t *testing.T) {
	module := Module{}
	lastLevelIdx := len(Levels) - 1
	model := module.buildPauseOverlay(lastLevelIdx)

	optionIDs := optionIDs(model.Options())
	if !contains(optionIDs, pauseOptionIDRestart) {
		t.Fatalf("expected restart option on last level, got %v", optionIDs)
	}
	if contains(optionIDs, pauseOptionIDNextLevel) {
		t.Fatalf("expected no next-level option on last level, got %v", optionIDs)
	}
}

func TestBuildHelpOverlayUsesInfoKind(t *testing.T) {
	module := Module{}
	model := module.buildHelpOverlay()

	if got := model.Kind(); got != overlay.KindInfo {
		t.Fatalf("expected info overlay kind, got %q", got)
	}

	if got := len(model.Options()); got != 0 {
		t.Fatalf("expected info overlay with no options, got %d", got)
	}
}

func TestBuildHelpOverlayDescribesGlobalUndoPoolForRunMode(t *testing.T) {
	module := Module{}
	model := module.buildHelpOverlay()

	joined := strings.Join(sectionLines(model.Sections()), "\n")
	if !strings.Contains(joined, "U: Deshacer (4 por partida)") {
		t.Fatalf("expected help overlay to document 4 global undos per run, got %q", joined)
	}
	if strings.Contains(joined, "U: Deshacer (1 por nivel)") {
		t.Fatalf("expected help overlay to avoid outdated per-level undo copy, got %q", joined)
	}
}

func TestBuildLevelCompleteOverlayUsesInfoKind(t *testing.T) {
	module := Module{}
	model := module.buildLevelCompleteOverlay(1, 7*time.Second+120*time.Millisecond)

	if got := model.Kind(); got != overlay.KindInfo {
		t.Fatalf("expected level-complete overlay to be info kind, got %q", got)
	}

	joined := strings.Join(sectionLines(model.Sections()), "\n")
	if got := model.Title(); got != "Level 2 complete!" {
		t.Fatalf("expected level-complete title, got %q", got)
	}
	if strings.Contains(joined, "Level 2 complete") {
		t.Fatalf("expected level-complete body to omit duplicated title, got %q", joined)
	}
	if !strings.Contains(joined, "00:07.120") {
		t.Fatalf("expected formatted elapsed time in overlay sections, got %q", joined)
	}
}

func TestBuildFinalSummaryOverlayUsesActionKindAndMenuOption(t *testing.T) {
	module := Module{}
	model := module.buildFinalSummaryOverlay(runSummary{
		result:          runResultCompleted,
		totalTime:       3 * time.Second,
		levelsCompleted: 2,
		levelReached:    2,
		undosUsed:       3,
		undosRemaining:  1,
		levelTimes:      []time.Duration{time.Second, 2 * time.Second},
	})

	if got := model.Kind(); got != overlay.KindAction {
		t.Fatalf("expected final-summary overlay to be action kind, got %q", got)
	}

	optionIDs := optionIDs(model.Options())
	if len(optionIDs) != 1 || optionIDs[0] != summaryOptionIDBackToMenu {
		t.Fatalf("expected only back-to-menu option, got %v", optionIDs)
	}
}

func TestBuildFinalSummaryOverlayReportsRunMetrics(t *testing.T) {
	module := Module{}
	model := module.buildFinalSummaryOverlay(runSummary{
		result:          runResultAbandoned,
		totalTime:       45 * time.Second,
		levelsCompleted: 3,
		levelReached:    4,
		undosUsed:       3,
		undosRemaining:  1,
		levelTimes:      []time.Duration{10 * time.Second, 15 * time.Second, 20 * time.Second},
	})

	joined := strings.Join(sectionLines(model.Sections()), "\n")
	for _, expected := range []string{"Resultado: abandonada", "Tiempo total: 00:45.000", "Niveles completados: 3/10", "Undos usados: 3", "Undos restantes: 1"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected final summary to include %q, got %q", expected, joined)
		}
	}
	if strings.Contains(joined, "Nivel alcanzado") {
		t.Fatalf("expected final summary to omit nivel alcanzado, got %q", joined)
	}
}

func TestBuildFinalSummaryOverlayPadsSummaryRowsToCommonWidth(t *testing.T) {
	module := Module{}
	model := module.buildFinalSummaryOverlay(runSummary{
		result:          runResultCompleted,
		totalTime:       2*time.Minute + 20*time.Second + 68*time.Millisecond,
		levelsCompleted: 3,
		levelReached:    3,
		undosUsed:       2,
		undosRemaining:  2,
		levelTimes:      []time.Duration{1234 * time.Millisecond, 12*time.Second + 45*time.Millisecond, 2*time.Minute + 6*time.Second + 789*time.Millisecond},
	})

	lines := sectionLines(model.Sections())
	if len(lines) < 2 {
		t.Fatalf("expected multiple summary lines, got %v", lines)
	}

	targetWidth := runewidth.StringWidth(lines[0])
	for i, line := range lines {
		if width := runewidth.StringWidth(line); width != targetWidth {
			t.Fatalf("expected summary line %d to have width %d, got %d (%q)", i, targetWidth, width, line)
		}
	}
}

func TestBuildFinalSummaryOverlayUsesNormalizedTitleSpacing(t *testing.T) {
	module := Module{}
	model := module.buildFinalSummaryOverlay(runSummary{
		result:          runResultCompleted,
		totalTime:       3 * time.Second,
		levelsCompleted: 2,
		undosUsed:       1,
		undosRemaining:  3,
		levelTimes:      []time.Duration{time.Second, 2 * time.Second},
	})

	if got := model.Title(); got != "🎉 FIN DEL JUEGO!🎉" {
		t.Fatalf("expected normalized title spacing, got %q", got)
	}
}

func TestDrawLevelHeaderShowsGlobalUndoCount(t *testing.T) {
	module := Module{Screen: render.NewBufferScreen(60, 24)}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}
	session := newRunSession(len(Levels))
	session.remainingUndos = 4

	module.drawLevel(game, 0, 12*time.Second, session)
	snapshot := module.Screen.(*render.BufferScreen).Snapshot()
	if !strings.Contains(snapshot, "U:4") {
		t.Fatalf("expected HUD to show remaining global undos, got %q", snapshot)
	}
}
func TestResolvePauseOverlayResultMapsOptionIDs(t *testing.T) {
	module := Module{}

	cases := []struct {
		name  string
		id    string
		want  pauseResult
		kind  overlay.ResultKind
		level int
	}{
		{name: "continue", id: pauseOptionIDContinue, want: pauseContinue, kind: overlay.ResultKindConfirmed, level: 0},
		{name: "restart", id: pauseOptionIDRestart, want: pauseRestart, kind: overlay.ResultKindConfirmed, level: 0},
		{name: "next", id: pauseOptionIDNextLevel, want: pauseNext, kind: overlay.ResultKindConfirmed, level: 0},
		{name: "quit", id: pauseOptionIDQuit, want: pauseQuit, kind: overlay.ResultKindConfirmed, level: 0},
		{name: "closed defaults continue", id: "", want: pauseContinue, kind: overlay.ResultKindClosed, level: 0},
		{name: "none defaults continue", id: "", want: pauseContinue, kind: overlay.ResultKindNone, level: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := module.resolvePauseOverlayResult(overlay.Result{Kind: tc.kind, OptionID: tc.id}, tc.level)
			if got != tc.want {
				t.Fatalf("expected pause result %v, got %v", tc.want, got)
			}
		})
	}
}

func TestShowPauseMenuNavigatesAndConfirmsSelection(t *testing.T) {
	module := Module{
		Screen:   render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionDown}}, {Actions: []input.Action{input.ActionSelect}}}},
	}

	if got, err := module.showPauseMenu(context.Background(), 0); err != nil || got != pauseRestart {
		if err != nil {
			t.Fatalf("expected pause menu navigation to succeed, got error %v", err)
		}
		t.Fatalf("expected pause restart after moving down and selecting, got %v", got)
	}
}

func TestShowPauseMenuCloseIntentReturnsContinue(t *testing.T) {
	module := Module{
		Screen:   render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionPause}}}},
	}

	if got, err := module.showPauseMenu(context.Background(), 0); err != nil || got != pauseContinue {
		if err != nil {
			t.Fatalf("expected close intent to succeed, got error %v", err)
		}
		t.Fatalf("expected pause continue on close intent, got %v", got)
	}
}

func TestShowPauseMenuReturnsContextErrorWhenCanceledWhileWaiting(t *testing.T) {
	module := Module{
		Screen:   render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	got, err := module.showPauseMenu(ctx, 0)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled from pause menu wait loop, got %v", err)
	}
	if got != pauseContinue {
		t.Fatalf("expected pauseContinue when pause menu exits on cancel, got %v", got)
	}
}

func TestShowHelpOverlayDismissesAndRedrawsLevel(t *testing.T) {
	trackedScreen := newTrackedScreen(render.NewBufferScreen(60, 24))
	module := Module{
		Screen:   trackedScreen,
		Provider: &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionLeft}}}},
	}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	if err := module.showHelpOverlay(context.Background(), game, 0, 2*time.Second, newRunSession(len(Levels))); err != nil {
		t.Fatalf("expected help overlay to dismiss cleanly, got %v", err)
	}
	if trackedScreen.PresentCalls() != 2 {
		t.Fatalf("expected help overlay flow to present twice (overlay + gameplay redraw), got %d", trackedScreen.PresentCalls())
	}

	snapshot := trackedScreen.Snapshot()
	if !strings.Contains(snapshot, "L1/") {
		t.Fatalf("expected gameplay header to be redrawn after help dismiss, got %q", snapshot)
	}
}

func TestShowHelpOverlayReturnsContextErrorWithoutInput(t *testing.T) {
	module := Module{
		Screen:   render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{},
	}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := module.showHelpOverlay(ctx, game, 0, 0, newRunSession(len(Levels))); err == nil {
		t.Fatal("expected canceled help overlay to return context error")
	}
}

func TestShowLevelCompleteReturnsContextErrorWhenCanceled(t *testing.T) {
	module := Module{Screen: render.NewBufferScreen(60, 24)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := module.showLevelComplete(ctx, 0, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled from level-complete delay, got %v", err)
	}
}

func TestShowHelpOverlayReturnsPresentErrorWhenGameplayRedrawFails(t *testing.T) {
	wantErr := errors.New("present failed on gameplay redraw")
	trackedScreen := newTrackedScreen(render.NewBufferScreen(60, 24)).FailPresentOnCall(2, wantErr)
	module := Module{
		Screen:   trackedScreen,
		Provider: &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionLeft}}}},
	}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	if err := module.showHelpOverlay(context.Background(), game, 0, 0, newRunSession(len(Levels))); !errors.Is(err, wantErr) {
		t.Fatalf("expected redraw present error %v, got %v", wantErr, err)
	}
}

func TestShowPauseMenuReturnsPresentErrorDuringRedraw(t *testing.T) {
	wantErr := errors.New("present failed on pause redraw")
	trackedScreen := newTrackedScreen(render.NewBufferScreen(60, 24)).FailPresentOnCall(2, wantErr)
	module := Module{
		Screen: trackedScreen,
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionDown}},
		}},
	}

	if _, err := module.showPauseMenu(context.Background(), 0); !errors.Is(err, wantErr) {
		t.Fatalf("expected pause redraw present error %v, got %v", wantErr, err)
	}
}

func TestShowFinalSummaryReturnsPresentError(t *testing.T) {
	wantErr := errors.New("present failed on summary")
	trackedScreen := newTrackedScreen(render.NewBufferScreen(60, 24)).FailPresentOnCall(1, wantErr)
	module := Module{
		Screen:   trackedScreen,
		Provider: &scriptedProvider{},
	}

	if err := module.showFinalSummary(context.Background(), runSummary{result: runResultCompleted, totalTime: time.Second, levelsCompleted: 1, levelReached: 1, undosRemaining: 4, levelTimes: []time.Duration{time.Second}}); !errors.Is(err, wantErr) {
		t.Fatalf("expected summary present error %v, got %v", wantErr, err)
	}
}

func TestShowFinalSummaryReturnsContextErrorWhenCanceledWithoutInput(t *testing.T) {
	originalDelay := finalSummaryInitialDelay
	finalSummaryInitialDelay = 0
	defer func() { finalSummaryInitialDelay = originalDelay }()

	module := Module{
		Screen:   render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := module.showFinalSummary(ctx, runSummary{result: runResultCompleted, totalTime: 3 * time.Second, levelsCompleted: 2, levelReached: 2, undosRemaining: 4, levelTimes: []time.Duration{time.Second, 2 * time.Second}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled from final summary wait loop, got %v", err)
	}
}

func TestShowFinalSummaryExitsOnSelectQuitAndPause(t *testing.T) {
	originalDelay := finalSummaryInitialDelay
	finalSummaryInitialDelay = 0
	defer func() { finalSummaryInitialDelay = originalDelay }()

	tests := []struct {
		name   string
		action input.Action
	}{
		{name: "select", action: input.ActionSelect},
		{name: "quit", action: input.ActionQuit},
		{name: "pause", action: input.ActionPause},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			module := Module{
				Screen:   render.NewBufferScreen(60, 24),
				Provider: &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{tc.action}}}},
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			errCh := make(chan error, 1)
			go func() {
				errCh <- module.showFinalSummary(ctx, runSummary{result: runResultCompleted, totalTime: 3 * time.Second, levelsCompleted: 2, levelReached: 2, undosRemaining: 4, levelTimes: []time.Duration{time.Second, 2 * time.Second}})
			}()

			select {
			case err := <-errCh:
				if err != nil {
					t.Fatalf("expected final summary to exit cleanly on %s, got %v", tc.name, err)
				}
			case <-ctx.Done():
				t.Fatalf("expected final summary to exit on %s", tc.name)
			}
		})
	}
}

func TestRunLevelQuitReturnsExplicitQuitResult(t *testing.T) {
	module := Module{
		Screen:   render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionQuit}}}},
	}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	result, err := module.runLevel(context.Background(), game, 0, 0, newRunSession(len(Levels)))
	if err != nil {
		t.Fatalf("expected quit to menu without runtime error, got %v", err)
	}
	if result.kind != levelResultQuitToMenu {
		t.Fatalf("expected quit result kind, got %v", result.kind)
	}
}

func TestRunLevelCompletedReturnsExplicitCompletedResultWithElapsed(t *testing.T) {
	module := Module{
		Screen:   render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionRight}}}},
	}
	game, err := ParseLevelString("#####\n#@$.#\n#####")
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	result, err := module.runLevel(context.Background(), game, 0, time.Second, newRunSession(len(Levels)))
	if err != nil {
		t.Fatalf("expected completed level without runtime error, got %v", err)
	}
	if result.kind != levelResultCompleted {
		t.Fatalf("expected completed result kind, got %v", result.kind)
	}
	if result.elapsed < time.Second {
		t.Fatalf("expected elapsed time to include completion duration, got %v", result.elapsed)
	}
}

func TestRunLevelPauseQuitReturnsExplicitQuitResult(t *testing.T) {
	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionPause}},
			{Actions: []input.Action{input.ActionQuit}},
		}},
	}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	result, err := module.runLevel(context.Background(), game, 0, 0, newRunSession(len(Levels)))
	if err != nil {
		t.Fatalf("expected pause quit to menu without runtime error, got %v", err)
	}
	if result.kind != levelResultQuitToMenu {
		t.Fatalf("expected quit result kind after pause quit, got %v", result.kind)
	}
}

func TestRunLevelPauseDoesNotAccumulatePausedTime(t *testing.T) {
	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &delayedProvider{frames: []delayedFrame{
			{actions: []input.Action{input.ActionPause}},
			{delay: 200 * time.Millisecond, actions: []input.Action{input.ActionPause}},
			{actions: []input.Action{input.ActionRight}},
		}},
	}
	game, err := ParseLevelString("#####\n#@$.#\n#####")
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	timedCtx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	result, err := module.runLevel(timedCtx, game, 0, time.Second, newRunSession(len(Levels)))
	if err != nil {
		t.Fatalf("expected completed level without runtime error, got %v", err)
	}
	if result.kind != levelResultCompleted {
		t.Fatalf("expected completed result kind, got %v", result.kind)
	}
	if result.elapsed >= time.Second+100*time.Millisecond {
		t.Fatalf("expected pause time to be excluded from elapsed total, got %v", result.elapsed)
	}
}

func TestRunLevelReturnsContextErrorWhenCanceledWhileIdle(t *testing.T) {
	module := Module{
		Screen:   render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{},
	}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	if _, err := module.runLevel(ctx, game, 0, 0, newRunSession(len(Levels))); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled from idle runLevel, got %v", err)
	}
}

func TestRunLevelPauseQuitSelectionReturnsQuitResult(t *testing.T) {
	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionPause}},
			{Actions: []input.Action{input.ActionQuit}},
		}},
	}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}

	result, err := module.runLevel(context.Background(), game, 0, 0, newRunSession(len(Levels)))
	if err != nil {
		t.Fatalf("expected pause quit selection without runtime error, got %v", err)
	}
	if result.kind != levelResultQuitToMenu {
		t.Fatalf("expected quit result kind after pause quit selection, got %v", result.kind)
	}
}

func TestRunAbandonFromPauseStopsSessionWithoutProgression(t *testing.T) {
	originalShowLevelCompleteFn := showLevelCompleteFn
	originalShowFinalSummaryFn := showFinalSummaryFn
	defer func() {
		showLevelCompleteFn = originalShowLevelCompleteFn
		showFinalSummaryFn = originalShowFinalSummaryFn
	}()

	showLevelCompleteCalls := 0
	showFinalSummaryCalls := 0

	showLevelCompleteFn = func(m Module, ctx context.Context, levelIdx int, elapsed time.Duration) error {
		showLevelCompleteCalls++
		return nil
	}
	showFinalSummaryFn = func(m Module, ctx context.Context, summary runSummary) error {
		showFinalSummaryCalls++
		return nil
	}

	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionPause}},
			{Actions: []input.Action{input.ActionQuit}},
		}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := module.Run(ctx); err != nil {
		t.Fatalf("expected Run to stop cleanly on run abandon from pause, got %v", err)
	}
	if showLevelCompleteCalls != 0 {
		t.Fatalf("expected no level-complete progression when abandoning run, got %d calls", showLevelCompleteCalls)
	}
	if showFinalSummaryCalls != 0 {
		t.Fatalf("expected no final-summary progression when abandoning run, got %d calls", showFinalSummaryCalls)
	}
}

func TestRunLevelUndoConsumesGlobalRunPool(t *testing.T) {
	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionLeft}},
			{Actions: []input.Action{input.ActionUndo}},
			{Actions: []input.Action{input.ActionQuit}},
		}},
	}
	game, err := ParseLevelString(Levels[1])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}
	start := game.Player
	session := newRunSession(len(Levels))

	result, err := module.runLevel(context.Background(), game, 1, 0, session)
	if err != nil {
		t.Fatalf("expected run level to quit without runtime error, got %v", err)
	}
	if result.kind != levelResultQuitToMenu {
		t.Fatalf("expected quit result kind, got %v", result.kind)
	}
	if session.remainingUndos != 3 {
		t.Fatalf("expected one undo consumption from global pool, got %d remaining", session.remainingUndos)
	}
	if game.Player != start {
		t.Fatalf("expected undo to restore player position to %v, got %v", start, game.Player)
	}
}

func TestRunLevelUndoBlockedWhenSessionPoolEmptyEvenWithGameHistory(t *testing.T) {
	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionLeft}},
			{Actions: []input.Action{input.ActionUndo}},
			{Actions: []input.Action{input.ActionQuit}},
		}},
	}
	game, err := ParseLevelString(Levels[1])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}
	session := newRunSession(len(Levels))
	session.remainingUndos = 0
	start := game.Player
	afterMove := grid.Point{X: start.X - 1, Y: start.Y}

	result, err := module.runLevel(context.Background(), game, 1, 0, session)
	if err != nil {
		t.Fatalf("expected run level to quit without runtime error, got %v", err)
	}
	if result.kind != levelResultQuitToMenu {
		t.Fatalf("expected quit result kind, got %v", result.kind)
	}
	if session.remainingUndos != 0 {
		t.Fatalf("expected empty global undo pool to remain empty, got %d", session.remainingUndos)
	}
	if game.Player != afterMove {
		t.Fatalf("expected player to remain at post-move position %v when undo is blocked, got %v", afterMove, game.Player)
	}
}

func TestRunLevelUndoWithEmptyHistoryDoesNotConsumeGlobalRunPool(t *testing.T) {
	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionUndo}},
			{Actions: []input.Action{input.ActionQuit}},
		}},
	}
	game, err := ParseLevelString(Levels[0])
	if err != nil {
		t.Fatalf("expected level parse to succeed, got %v", err)
	}
	start := game.Player
	session := newRunSession(len(Levels))

	result, err := module.runLevel(context.Background(), game, 0, 0, session)
	if err != nil {
		t.Fatalf("expected run level to quit without runtime error, got %v", err)
	}
	if result.kind != levelResultQuitToMenu {
		t.Fatalf("expected quit result kind, got %v", result.kind)
	}
	if session.remainingUndos != 4 {
		t.Fatalf("expected undo pool to stay at 4 when undo history is empty, got %d", session.remainingUndos)
	}
	if game.Player != start {
		t.Fatalf("expected player position unchanged without undo history, got %v", game.Player)
	}
}

func TestModuleRunReturnsNilOnExplicitQuitWithoutCompletionFlow(t *testing.T) {
	originalRunLevelFn := runLevelFn
	originalShowLevelCompleteFn := showLevelCompleteFn
	originalShowFinalSummaryFn := showFinalSummaryFn
	defer func() {
		runLevelFn = originalRunLevelFn
		showLevelCompleteFn = originalShowLevelCompleteFn
		showFinalSummaryFn = originalShowFinalSummaryFn
	}()

	showLevelCompleteCalls := 0
	showFinalSummaryCalls := 0
	runLevelFn = func(m Module, ctx context.Context, game *Game, levelIdx int, initialPenalty time.Duration, session *runSession) (levelResult, error) {
		return levelResult{kind: levelResultQuitToMenu}, nil
	}
	showLevelCompleteFn = func(m Module, ctx context.Context, levelIdx int, elapsed time.Duration) error {
		showLevelCompleteCalls++
		return nil
	}
	showFinalSummaryFn = func(m Module, ctx context.Context, summary runSummary) error {
		showFinalSummaryCalls++
		return nil
	}

	module := Module{Screen: render.NewBufferScreen(60, 24), Provider: &scriptedProvider{}}
	if err := module.Run(context.Background()); err != nil {
		t.Fatalf("expected Run to return nil on explicit quit, got %v", err)
	}
	if showLevelCompleteCalls != 0 {
		t.Fatalf("expected no completion flow on quit, got %d calls", showLevelCompleteCalls)
	}
	if showFinalSummaryCalls != 0 {
		t.Fatalf("expected no final summary flow on quit, got %d calls", showFinalSummaryCalls)
	}
}

func TestModuleRunProcessesExplicitCompletedResult(t *testing.T) {
	originalRunLevelFn := runLevelFn
	originalShowLevelCompleteFn := showLevelCompleteFn
	originalShowFinalSummaryFn := showFinalSummaryFn
	defer func() {
		runLevelFn = originalRunLevelFn
		showLevelCompleteFn = originalShowLevelCompleteFn
		showFinalSummaryFn = originalShowFinalSummaryFn
	}()

	completedElapsed := 3*time.Second + 250*time.Millisecond
	showLevelCompleteCalls := 0
	showFinalSummaryCalls := 0
	seenElapsed := time.Duration(0)
	runCalls := 0
	runLevelFn = func(m Module, ctx context.Context, game *Game, levelIdx int, initialPenalty time.Duration, session *runSession) (levelResult, error) {
		runCalls++
		if runCalls == 1 {
			return levelResult{kind: levelResultCompleted, elapsed: completedElapsed}, nil
		}
		return levelResult{kind: levelResultQuitToMenu}, nil
	}
	showLevelCompleteFn = func(m Module, ctx context.Context, levelIdx int, elapsed time.Duration) error {
		showLevelCompleteCalls++
		seenElapsed = elapsed
		if levelIdx != 0 {
			t.Fatalf("expected completion callback for level 0, got %d", levelIdx)
		}
		return nil
	}
	showFinalSummaryFn = func(m Module, ctx context.Context, summary runSummary) error {
		showFinalSummaryCalls++
		return nil
	}

	module := Module{Screen: render.NewBufferScreen(60, 24), Provider: &scriptedProvider{}}
	if err := module.Run(context.Background()); err != nil {
		t.Fatalf("expected Run to succeed, got %v", err)
	}
	if showLevelCompleteCalls != 1 {
		t.Fatalf("expected one completion callback, got %d", showLevelCompleteCalls)
	}
	if seenElapsed != completedElapsed {
		t.Fatalf("expected elapsed %v forwarded to completion flow, got %v", completedElapsed, seenElapsed)
	}
	if showFinalSummaryCalls != 0 {
		t.Fatalf("expected no final summary after subsequent quit, got %d calls", showFinalSummaryCalls)
	}
}

func TestModuleRunHandlesRestartAndNextLevelWithoutCompletionFlow(t *testing.T) {
	originalRunLevelFn := runLevelFn
	originalShowLevelCompleteFn := showLevelCompleteFn
	originalShowFinalSummaryFn := showFinalSummaryFn
	defer func() {
		runLevelFn = originalRunLevelFn
		showLevelCompleteFn = originalShowLevelCompleteFn
		showFinalSummaryFn = originalShowFinalSummaryFn
	}()

	type runCall struct {
		levelIdx int
		penalty  time.Duration
	}
	var calls []runCall
	showLevelCompleteCalls := 0
	showFinalSummaryCalls := 0

	runLevelFn = func(m Module, ctx context.Context, game *Game, levelIdx int, initialPenalty time.Duration, session *runSession) (levelResult, error) {
		calls = append(calls, runCall{levelIdx: levelIdx, penalty: initialPenalty})
		switch len(calls) {
		case 1:
			return levelResult{kind: levelResultRestart}, nil
		case 2:
			return levelResult{kind: levelResultNextLevel, elapsed: 45 * time.Second}, nil
		case 3:
			return levelResult{kind: levelResultQuitToMenu}, nil
		default:
			t.Fatalf("unexpected extra runLevel call %d", len(calls))
			return levelResult{}, nil
		}
	}
	showLevelCompleteFn = func(m Module, ctx context.Context, levelIdx int, elapsed time.Duration) error {
		showLevelCompleteCalls++
		return nil
	}
	showFinalSummaryFn = func(m Module, ctx context.Context, summary runSummary) error {
		showFinalSummaryCalls++
		return nil
	}

	module := Module{Screen: render.NewBufferScreen(60, 24), Provider: &scriptedProvider{}}
	if err := module.Run(context.Background()); err != nil {
		t.Fatalf("expected Run to succeed, got %v", err)
	}

	if showLevelCompleteCalls != 0 {
		t.Fatalf("expected no completion flow for restart/next-level path, got %d calls", showLevelCompleteCalls)
	}
	if showFinalSummaryCalls != 0 {
		t.Fatalf("expected no final summary on quit, got %d calls", showFinalSummaryCalls)
	}

	if len(calls) != 3 {
		t.Fatalf("expected three runLevel calls, got %d", len(calls))
	}
	if calls[0].levelIdx != 0 || calls[0].penalty != 0 {
		t.Fatalf("expected first call level 0 penalty 0, got level %d penalty %v", calls[0].levelIdx, calls[0].penalty)
	}
	if calls[1].levelIdx != 0 || calls[1].penalty != 5*time.Second {
		t.Fatalf("expected restart rerun at level 0 with +5s penalty, got level %d penalty %v", calls[1].levelIdx, calls[1].penalty)
	}
	if calls[2].levelIdx != 1 || calls[2].penalty != 0 {
		t.Fatalf("expected next-level advance to level 1 with reset penalty, got level %d penalty %v", calls[2].levelIdx, calls[2].penalty)
	}
}

func TestTopLevelProgressionHandlesRestartAndNextLevel(t *testing.T) {
	TestModuleRunHandlesRestartAndNextLevelWithoutCompletionFlow(t)
}

func TestRunReturnsNilOnQuitToMenuWithoutLevelComplete(t *testing.T) {
	originalShowLevelCompleteFn := showLevelCompleteFn
	originalShowFinalSummaryFn := showFinalSummaryFn
	defer func() {
		showLevelCompleteFn = originalShowLevelCompleteFn
		showFinalSummaryFn = originalShowFinalSummaryFn
	}()

	showLevelCompleteCalls := 0
	showFinalSummaryCalls := 0

	showLevelCompleteFn = func(m Module, ctx context.Context, levelIdx int, elapsed time.Duration) error {
		showLevelCompleteCalls++
		return nil
	}
	showFinalSummaryFn = func(m Module, ctx context.Context, summary runSummary) error {
		showFinalSummaryCalls++
		return nil
	}

	module := Module{
		Screen:   render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionQuit}}}},
	}
	if err := module.Run(context.Background()); err != nil {
		t.Fatalf("expected Run to stop cleanly on quit-to-menu, got %v", err)
	}
	if showLevelCompleteCalls != 0 {
		t.Fatalf("expected no level-complete progression on quit, got %d calls", showLevelCompleteCalls)
	}
	if showFinalSummaryCalls != 0 {
		t.Fatalf("expected no final-summary progression on quit, got %d calls", showFinalSummaryCalls)
	}
}

func TestRunPauseQuitStopsWithoutProgression(t *testing.T) {
	originalShowLevelCompleteFn := showLevelCompleteFn
	originalShowFinalSummaryFn := showFinalSummaryFn
	defer func() {
		showLevelCompleteFn = originalShowLevelCompleteFn
		showFinalSummaryFn = originalShowFinalSummaryFn
	}()

	showLevelCompleteCalls := 0
	showFinalSummaryCalls := 0

	showLevelCompleteFn = func(m Module, ctx context.Context, levelIdx int, elapsed time.Duration) error {
		showLevelCompleteCalls++
		return nil
	}
	showFinalSummaryFn = func(m Module, ctx context.Context, summary runSummary) error {
		showFinalSummaryCalls++
		return nil
	}

	module := Module{
		Screen: render.NewBufferScreen(60, 24),
		Provider: &scriptedProvider{frames: []input.FrameInput{
			{Actions: []input.Action{input.ActionPause}},
			{Actions: []input.Action{input.ActionQuit}},
		}},
	}
	if err := module.Run(context.Background()); err != nil {
		t.Fatalf("expected Run to stop cleanly on pause quit, got %v", err)
	}
	if showLevelCompleteCalls != 0 {
		t.Fatalf("expected no level-complete progression on pause quit, got %d calls", showLevelCompleteCalls)
	}
	if showFinalSummaryCalls != 0 {
		t.Fatalf("expected no final-summary progression on pause quit, got %d calls", showFinalSummaryCalls)
	}
}

func optionIDs(options []overlay.Option) []string {
	ids := make([]string, 0, len(options))
	for _, opt := range options {
		ids = append(ids, opt.ID)
	}
	return ids
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sectionLines(sections []overlay.Section) []string {
	lines := make([]string, 0)
	for _, section := range sections {
		lines = append(lines, section.Lines...)
	}
	return lines
}

type scriptedProvider struct {
	frames []input.FrameInput
	index  int
}

func (p *scriptedProvider) Poll() input.FrameInput {
	if p.index >= len(p.frames) {
		return input.FrameInput{}
	}
	frame := p.frames[p.index]
	p.index++
	return frame
}

type delayedFrame struct {
	delay   time.Duration
	actions []input.Action
}

type delayedProvider struct {
	frames []delayedFrame
	index  int
}

func (p *delayedProvider) Poll() input.FrameInput {
	if p.index >= len(p.frames) {
		return input.FrameInput{}
	}
	frame := p.frames[p.index]
	p.index++
	if frame.delay > 0 {
		time.Sleep(frame.delay)
	}
	return input.FrameInput{Actions: append([]input.Action(nil), frame.actions...)}
}

type trackedScreen struct {
	inner       *render.BufferScreen
	presentCall int
	failOnCall  int
	failErr     error
}

func newTrackedScreen(inner *render.BufferScreen) *trackedScreen {
	return &trackedScreen{inner: inner}
}

func (s *trackedScreen) Size() (int, int) {
	return s.inner.Size()
}

func (s *trackedScreen) Clear() {
	s.inner.Clear()
}

func (s *trackedScreen) SetCell(x, y int, cell render.Cell) {
	s.inner.SetCell(x, y, cell)
}

func (s *trackedScreen) Present() error {
	s.presentCall++
	if s.failOnCall > 0 && s.presentCall == s.failOnCall {
		return s.failErr
	}
	return s.inner.Present()
}

func (s *trackedScreen) PresentCalls() int {
	return s.presentCall
}

func (s *trackedScreen) Snapshot() string {
	return s.inner.Snapshot()
}

func (s *trackedScreen) FailPresentOnCall(call int, err error) *trackedScreen {
	s.failOnCall = call
	s.failErr = err
	return s
}

func TestAFinishedRunEntersTheHighScoresWithItsName(t *testing.T) {
	originalRunLevelFn, originalShowLevelCompleteFn, originalShowFinalSummaryFn, originalShowTableFn := runLevelFn, showLevelCompleteFn, showFinalSummaryFn, showTableFn
	defer func() {
		runLevelFn, showLevelCompleteFn, showFinalSummaryFn, showTableFn = originalRunLevelFn, originalShowLevelCompleteFn, originalShowFinalSummaryFn, originalShowTableFn
	}()
	runLevelFn = func(Module, context.Context, *Game, int, time.Duration, *runSession) (levelResult, error) {
		return levelResult{kind: levelResultCompleted, elapsed: 30 * time.Second}, nil
	}
	showLevelCompleteFn = func(Module, context.Context, int, time.Duration) error { return nil }
	var shown []string
	summaries := 0
	showTableFn = func(_ Module, _ context.Context, table []string) error {
		shown = table
		return nil
	}
	showFinalSummaryFn = func(Module, context.Context, runSummary) error {
		summaries++
		return nil
	}
	book, _ := records.Open("")
	book.Add("sokoban", records.Entry{Name: "slow", Score: len(Levels), Seconds: 99999})
	module := NewModule(render.NewBufferScreen(60, 24), &scriptedProvider{frames: []input.FrameInput{
		{Text: []rune("bo")},
		{Actions: []input.Action{input.ActionSelect}},
	}}).WithBoard(book.Board("sokoban"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := module.Run(ctx); err != nil {
		t.Fatal(err)
	}
	want := records.Entry{Name: "bo", Score: len(Levels), Seconds: 30 * len(Levels)}
	if top := book.Top("sokoban"); len(top) != 2 || top[0] != want {
		t.Fatalf("a finished run ranks by levels, then less time: %+v", top)
	}
	table := strings.Join(shown, "\n")
	if !strings.Contains(table, "1. bo") || !strings.Contains(table, "◂") || !strings.Contains(table, "2. slow") || summaries != 1 {
		t.Fatalf("the table shows before the final summary:\n%s", table)
	}
}

func TestTheTableScreenWaitsForAKey(t *testing.T) {
	screen := render.NewBufferScreen(80, 24)
	module := NewModule(screen, &scriptedProvider{frames: []input.FrameInput{{}, {Actions: []input.Action{input.ActionSelect}}}})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := module.showTable(ctx, highscore.Lines([]records.Entry{{Name: "bo", Score: 10, Seconds: 61}}, 1, nil)); err != nil {
		t.Fatal(err)
	}
	if out := screen.Snapshot(); !strings.Contains(out, "HIGH SCORES") || !strings.Contains(out, "1. bo") || !strings.Contains(out, "01:01") {
		t.Fatalf("the table screen:\n%s", out)
	}
}
