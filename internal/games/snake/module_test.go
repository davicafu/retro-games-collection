package snake

import (
	"context"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/session"
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

func newSession(cfg Config) *session.Session {
	return session.New(func() session.Round { return round{NewGame(cfg)} })
}

func snakeOf(s *session.Session) *Game { return s.Round().(round).Game }

// wallConfig places the snake on the top edge: moving up dies on the first tick.
var wallConfig = Config{Width: 3, Height: 3, Start: Position{X: 1, Y: 0}}

func TestRound_UpdatePreservesDirectionAcrossEmptyFrame(t *testing.T) {
	r := round{NewGame(Config{Width: 12, Height: 12, Start: Position{X: 1, Y: 1}, Food: []Position{{X: 4, Y: 1}}})}

	if err := r.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionDown}}); err != nil {
		t.Fatalf("expected first update to succeed, got %v", err)
	}
	if err := r.Update(context.Background(), input.FrameInput{}); err != nil {
		t.Fatalf("expected empty-frame update to succeed, got %v", err)
	}

	if got := r.Direction(); got != DirectionDown {
		t.Fatalf("expected direction to remain down on empty input frame, got %#v", got)
	}
}

func TestSession_QuitDoesNotMoveOnQuitTick(t *testing.T) {
	s := newSession(Config{Width: 12, Height: 12, Start: Position{X: 1, Y: 1}, Food: []Position{{X: 4, Y: 1}}})

	if err := s.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionDown}}); err != nil {
		t.Fatalf("expected first update to succeed, got %v", err)
	}
	headBeforeQuit := snakeOf(s).Head()
	if err := s.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionQuit}}); err != nil {
		t.Fatalf("expected quit update to succeed, got %v", err)
	}

	if got := snakeOf(s).Head(); got != headBeforeQuit {
		t.Fatalf("expected quit tick to not move the snake, got head=%v want=%v", got, headBeforeQuit)
	}
	if !s.ShouldQuit() {
		t.Fatal("expected session to quit after quit action")
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

func TestSession_DeathShowsGameOverWithScore(t *testing.T) {
	s := newSession(wallConfig)

	if err := s.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionUp}}); err != nil {
		t.Fatalf("expected move to succeed, got %v", err)
	}

	if !snakeOf(s).Dead() {
		t.Fatal("expected snake to be dead after wall collision")
	}
	m, open := s.Overlay()
	if !open || m.Title() != "Game Over" {
		t.Fatal("expected Game Over overlay after death")
	}
	if got := m.Sections()[0].Lines[0]; got != "Score: 0" {
		t.Fatalf("expected score in Game Over, got %q", got)
	}
	if s.ShouldQuit() {
		t.Fatal("expected ShouldQuit to return false when overlay is active")
	}
}

func TestSession_GameOverRestartResetsState(t *testing.T) {
	s := newSession(wallConfig)
	s.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionUp}})

	// Restart is selected first; Q confirms the selection on Game Over.
	if err := s.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionQuit}}); err != nil {
		t.Fatalf("expected overlay confirm restart to succeed, got %v", err)
	}

	if s.ShouldQuit() {
		t.Fatal("expected restart to keep playing")
	}
	if _, open := s.Overlay(); open {
		t.Fatal("expected overlay to be closed after restart")
	}
	if snakeOf(s).Dead() || snakeOf(s).Score() != 0 {
		t.Fatal("expected a fresh snake after restart")
	}
}

func TestSession_GameOverMenuExitsGame(t *testing.T) {
	s := newSession(wallConfig)
	s.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionUp}})
	s.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionDown}})

	if err := s.Update(context.Background(), input.FrameInput{Actions: []input.Action{input.ActionSelect}}); err != nil {
		t.Fatalf("expected overlay confirm via Select to succeed, got %v", err)
	}

	if !s.ShouldQuit() {
		t.Fatal("expected Menu to end the game")
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
