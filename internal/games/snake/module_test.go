package snake

import (
	"context"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/tcellscreen"
)

type scriptedProvider struct {
	frames []input.FrameInput
	idx    int
}

func (p *scriptedProvider) Poll() input.FrameInput {
	if p.idx >= len(p.frames) {
		return input.FrameInput{}
	}

	frame := p.frames[p.idx]
	p.idx++
	return frame
}

func TestRuntimeGame_UpdatePreservesDirectionAcrossEmptyFrame(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 12, Height: 12, Start: Position{X: 1, Y: 1}, Food: []Position{{X: 4, Y: 1}}})}

	if err := g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionDown}}); err != nil {
		t.Fatalf("expected first update to succeed, got %v", err)
	}

	if err := g.Update(context.Background(), input.FrameInput{}); err != nil {
		t.Fatalf("expected empty-frame update to succeed, got %v", err)
	}

	if got := g.Direction(); got != DirectionDown {
		t.Fatalf("expected direction to remain down on empty input frame, got %#v", got)
	}
}

func TestRuntimeGame_UpdateQuitDoesNotMoveOnQuitTick(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 12, Height: 12, Start: Position{X: 1, Y: 1}, Food: []Position{{X: 4, Y: 1}}})}

	if err := g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionDown}}); err != nil {
		t.Fatalf("expected first update to succeed, got %v", err)
	}

	headBeforeQuit := g.Head()
	if err := g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionQuit}}); err != nil {
		t.Fatalf("expected quit update to succeed, got %v", err)
	}

	if got := g.Head(); got != headBeforeQuit {
		t.Fatalf("expected quit tick to not move the snake, got head=%v want=%v", got, headBeforeQuit)
	}

	if !g.ShouldQuit() {
		t.Fatal("expected runtime game to quit after quit action")
	}
}

func TestModule_RunRendersAndQuitsGracefully(t *testing.T) {
	manual := clock.NewManual()
	screen := tcellscreen.NewSimulationScreenForTest(14, 15)
	provider := &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionDown}}, {}, {Actions: []input.Action{input.ActionQuit}}}}
	module := NewModule(manual, screen, provider, nil)

	errCh := make(chan error, 1)
	go func() { errCh <- module.Run(context.Background()) }()
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)

	if err := <-errCh; err != nil {
		t.Fatalf("expected graceful quit, got %v", err)
	}

	snapshot := screen.Snapshot()
	if got := snapshot[5][2]; got != '█' {
		t.Fatalf("expected snake head rendered at projected playfield cell (2,5), got %q", got)
	}

	if got := snapshot[3][5]; got != '•' {
		t.Fatalf("expected food rendered at projected playfield cell (5,3), got %q", got)
	}

	if got := snapshot[1][0]; got != '+' {
		t.Fatalf("expected top-left border corner, got %q", got)
	}
}

func TestModule_RunContinuesAfterFirstFoodPickup(t *testing.T) {
	manual := clock.NewManual()
	screen := tcellscreen.NewSimulationScreenForTest(12, 12)
	provider := &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionRight}}, {}, {}, {}, {Actions: []input.Action{input.ActionQuit}}}}
	module := NewModule(manual, screen, provider, nil)

	errCh := make(chan error, 1)
	go func() { errCh <- module.Run(context.Background()) }()
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)

	if err := <-errCh; err != nil {
		t.Fatalf("expected graceful exit after continued play, got %v", err)
	}

	snapshot := screen.Snapshot()
	if snapshot[1][5] != '█' {
		t.Fatalf("expected snake to advance one more tick after first pickup, got %q at (5,1)", snapshot[1][5])
	}
}

func TestRuntimeGame_DeathActivatesOverlayNotForceQuit(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 3, Height: 3, Start: Position{X: 1, Y: 0}})}

	// Start at top edge: Up drives into wall at y=-1
	if err := g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionUp}}); err != nil {
		t.Fatalf("expected move to succeed, got %v", err)
	}

	if !g.Dead() {
		t.Fatal("expected snake to be dead after wall collision")
	}
	if g.forceQuit {
		t.Fatal("expected death to NOT set forceQuit")
	}
	if g.overlay == nil {
		t.Fatal("expected overlay to be active after death")
	}
	if g.ShouldQuit() {
		t.Fatal("expected ShouldQuit to return false when overlay is active")
	}
}

func TestRuntimeGame_OverlayNavigationChangesSelection(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 3, Height: 3, Start: Position{X: 1, Y: 0}})}

	// Die first to activate overlay
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionUp}})

	if idx := g.overlay.SelectedIndex(); idx != 0 {
		t.Fatalf("expected initial selection at index 0, got %d", idx)
	}

	// Navigate down
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionDown}})
	if idx := g.overlay.SelectedIndex(); idx != 1 {
		t.Fatalf("expected selection at index 1 after Down, got %d", idx)
	}

	// Navigate up
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionUp}})
	if idx := g.overlay.SelectedIndex(); idx != 0 {
		t.Fatalf("expected selection at index 0 after Up, got %d", idx)
	}
}

func TestRuntimeGame_OverlayConfirmQuitExitsGame(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 3, Height: 3, Start: Position{X: 1, Y: 0}})}

	// Die first, then navigate to Quit (index 2)
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionUp}})
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionDown}})
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionDown}})

	// Confirm quit via ActionQuit while overlay is active
	if err := g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionQuit}}); err != nil {
		t.Fatalf("expected overlay confirm to succeed, got %v", err)
	}

	if !g.forceQuit {
		t.Fatal("expected forceQuit after overlay confirm quit")
	}
	if !g.ShouldQuit() {
		t.Fatal("expected ShouldQuit to return true after overlay confirm quit")
	}
}

func TestRuntimeGame_OverlayConfirmRestartResetsState(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 3, Height: 3, Start: Position{X: 1, Y: 0}})}

	// Die first
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionUp}})

	// Navigate to Restart (index 0 is already selected)
	if err := g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionQuit}}); err != nil {
		t.Fatalf("expected overlay confirm restart to succeed, got %v", err)
	}

	if g.forceQuit {
		t.Fatal("expected forceQuit to be false after restart")
	}
	if g.overlay != nil {
		t.Fatal("expected overlay to be nil after restart")
	}
	if g.Dead() {
		t.Fatal("expected game to be alive after restart")
	}
	if g.Score() != 0 {
		t.Fatalf("expected score reset to 0 after restart, got %d", g.Score())
	}
}

func TestRuntimeGame_OverlaySelectActionConfirms(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 3, Height: 3, Start: Position{X: 1, Y: 0}})}

	// Die first, then navigate to Quit (index 2)
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionUp}})
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionDown}})
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionDown}})

	// Confirm with ActionSelect (Enter/Space)
	if err := g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionSelect}}); err != nil {
		t.Fatalf("expected overlay confirm via Select to succeed, got %v", err)
	}

	if !g.forceQuit {
		t.Fatal("expected forceQuit after overlay confirm with ActionSelect")
	}
}

func TestRuntimeGame_OverlayEmptyInputDoesNotCrash(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 3, Height: 3, Start: Position{X: 1, Y: 0}})}

	// Die first
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionUp}})

	// Empty frame while overlay active
	if err := g.Update(context.Background(), input.FrameInput{}); err != nil {
		t.Fatalf("expected empty frame to succeed during overlay, got %v", err)
	}

	if g.overlay == nil {
		t.Fatal("expected overlay to remain active after empty frame")
	}
	if g.ShouldQuit() {
		t.Fatal("expected ShouldQuit to return false after empty overlay frame")
	}
}

func TestRuntimeGame_PauseActivatesOverlay(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 8, Height: 8, Start: Position{X: 1, Y: 1}})}
	headBefore := g.Head()

	if err := g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionPause}}); err != nil {
		t.Fatalf("expected pause to succeed, got %v", err)
	}

	if !g.paused {
		t.Fatal("expected paused to be true")
	}
	if g.overlay == nil {
		t.Fatal("expected overlay active after pause")
	}
	if g.overlay.Title() != "PAUSED" {
		t.Fatalf("expected overlay title PAUSED, got %q", g.overlay.Title())
	}
	if g.Head() != headBefore {
		t.Fatal("expected game state not to advance on pause frame")
	}
}

func TestRuntimeGame_PauseOverlayHasContinueFirst(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 8, Height: 8, Start: Position{X: 1, Y: 1}})}

	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionPause}})

	if len(g.overlay.Options()) < 1 {
		t.Fatal("expected at least one option in pause overlay")
	}
	if g.overlay.Options()[0].Action != OverlayActionContinue {
		t.Fatalf("expected first option to be Continue, got %q", g.overlay.Options()[0].Action)
	}
}

func TestRuntimeGame_PauseEscapeResumes(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 8, Height: 8, Start: Position{X: 1, Y: 1}})}

	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionPause}})

	// Escape during pause resumes
	if err := g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionQuit}}); err != nil {
		t.Fatalf("expected escape during pause to succeed, got %v", err)
	}

	if g.paused {
		t.Fatal("expected paused to be false after escape")
	}
	if g.overlay != nil {
		t.Fatal("expected overlay cleared after escape during pause")
	}
}

func TestRuntimeGame_PauseToggleResumes(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 8, Height: 8, Start: Position{X: 1, Y: 1}})}

	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionPause}})

	// 'p' during pause resumes (ActionPause)
	if err := g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionPause}}); err != nil {
		t.Fatalf("expected pause toggle to succeed, got %v", err)
	}

	if g.paused {
		t.Fatal("expected paused to be false after second pause toggle")
	}
	if g.overlay != nil {
		t.Fatal("expected overlay cleared after pause toggle")
	}
}

func TestRuntimeGame_PauseContinueViaSelect(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 8, Height: 8, Start: Position{X: 1, Y: 1}})}

	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionPause}})

	// Continue is selected by default (index 0), Enter/Space confirms
	if err := g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionSelect}}); err != nil {
		t.Fatalf("expected continue via select to succeed, got %v", err)
	}

	if g.paused {
		t.Fatal("expected paused to be false after continue")
	}
	if g.overlay != nil {
		t.Fatal("expected overlay cleared after continue")
	}
}

func TestRuntimeGame_GameOverOverlayNoContinue(t *testing.T) {
	g := &runtimeGame{Game: NewGame(Config{Width: 3, Height: 3, Start: Position{X: 1, Y: 0}})}

	// Die via wall collision
	g.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionUp}})

	for _, opt := range g.overlay.Options() {
		if opt.Action == OverlayActionContinue {
			t.Fatal("game-over overlay should NOT have Continue option")
		}
	}
}

func TestModule_RunUsesInjectedSelectorForRespawnPath(t *testing.T) {
	manual := clock.NewManual()
	screen := tcellscreen.NewSimulationScreenForTest(12, 12)
	provider := &scriptedProvider{frames: []input.FrameInput{{Actions: []input.Action{input.ActionRight}}, {}, {}, {Actions: []input.Action{input.ActionQuit}}}}
	module := NewModule(manual, screen, provider, NewSequenceSelector([]int{5}))

	errCh := make(chan error, 1)
	go func() { errCh <- module.Run(context.Background()) }()
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)
	manual.Advance(time.Millisecond)

	if err := <-errCh; err != nil {
		t.Fatalf("expected graceful exit after respawn verification, got %v", err)
	}

	snapshot := screen.Snapshot()
	if got := snapshot[0][5]; got != '•' {
		t.Fatalf("expected respawned food selected via factory at (5,0), got %q", got)
	}

	if got := snapshot[0][0]; got == '•' {
		t.Fatalf("expected non-fallback selection, but fallback cell (0,0) contains food")
	}
}
