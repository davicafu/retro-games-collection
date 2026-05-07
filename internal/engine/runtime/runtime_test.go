package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/platform/terminal/headless"
)

type fakeGame struct{ updates int }

func (g *fakeGame) Init(context.Context) error                     { return nil }
func (g *fakeGame) Update(context.Context, input.FrameInput) error { g.updates++; return nil }
func (g *fakeGame) Render(context.Context, Screen) error           { return nil }
func (g *fakeGame) ShouldQuit() bool                               { return g.updates >= 3 }

type scriptedGame struct {
	initErr   error
	updateErr error
	renderErr error
	updates   int
	quitAfter int
}

func (g *scriptedGame) Init(context.Context) error { return g.initErr }
func (g *scriptedGame) Update(context.Context, input.FrameInput) error {
	g.updates++
	if g.updateErr != nil {
		return g.updateErr
	}
	return nil
}
func (g *scriptedGame) Render(context.Context, Screen) error { return g.renderErr }
func (g *scriptedGame) ShouldQuit() bool {
	if g.quitAfter <= 0 {
		return false
	}
	return g.updates >= g.quitAfter
}

type presentErrScreen struct{ err error }

func (s presentErrScreen) Size() (int, int)                        { return 20, 10 }
func (s presentErrScreen) Clear()                                  {}
func (s presentErrScreen) SetCell(int, int, render.Cell)           {}
func (s presentErrScreen) Present() error                          { return s.err }

func TestRuntime_RunAdvancesUntilGameQuits(t *testing.T) {
	manual := clock.NewManual()
	screen := headless.NewScreen(20, 10)
	rt := New(Config{Clock: manual, Screen: screen, Input: input.Static(nil)})

	game := &fakeGame{}
	errCh := make(chan error, 1)
	go func() { errCh <- rt.Run(context.Background(), game) }()

	for range 3 {
		manual.Advance(time.Millisecond)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("expected no run error, got %v", err)
	}

	if game.updates != 3 {
		t.Fatalf("expected 3 updates, got %d", game.updates)
	}
}

func TestRuntime_RunPropagatesInitError(t *testing.T) {
	manual := clock.NewManual()
	screen := headless.NewScreen(20, 10)
	rt := New(Config{Clock: manual, Screen: screen, Input: input.Static(nil)})

	wantErr := errors.New("init failed")
	game := &scriptedGame{initErr: wantErr}

	err := rt.Run(context.Background(), game)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected init error %v, got %v", wantErr, err)
	}
}

func TestRuntime_RunPropagatesUpdateError(t *testing.T) {
	manual := clock.NewManual()
	screen := headless.NewScreen(20, 10)
	rt := New(Config{Clock: manual, Screen: screen, Input: input.Static(nil)})

	wantErr := errors.New("update failed")
	game := &scriptedGame{updateErr: wantErr}

	errCh := make(chan error, 1)
	go func() { errCh <- rt.Run(context.Background(), game) }()
	manual.Advance(time.Millisecond)

	err := <-errCh
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected update error %v, got %v", wantErr, err)
	}
}

func TestRuntime_RunPropagatesRenderError(t *testing.T) {
	manual := clock.NewManual()
	screen := headless.NewScreen(20, 10)
	rt := New(Config{Clock: manual, Screen: screen, Input: input.Static(nil)})

	wantErr := errors.New("render failed")
	game := &scriptedGame{renderErr: wantErr}

	errCh := make(chan error, 1)
	go func() { errCh <- rt.Run(context.Background(), game) }()
	manual.Advance(time.Millisecond)

	err := <-errCh
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected render error %v, got %v", wantErr, err)
	}
}

func TestRuntime_RunPropagatesPresentError(t *testing.T) {
	manual := clock.NewManual()
	wantErr := errors.New("present failed")
	rt := New(Config{Clock: manual, Screen: presentErrScreen{err: wantErr}, Input: input.Static(nil)})

	game := &scriptedGame{}
	errCh := make(chan error, 1)
	go func() { errCh <- rt.Run(context.Background(), game) }()
	manual.Advance(time.Millisecond)

	err := <-errCh
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected present error %v, got %v", wantErr, err)
	}
}

func TestRuntime_RunReturnsContextCancellation(t *testing.T) {
	manual := clock.NewManual()
	screen := headless.NewScreen(20, 10)
	rt := New(Config{Clock: manual, Screen: screen, Input: input.Static(nil)})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := rt.Run(ctx, &scriptedGame{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
