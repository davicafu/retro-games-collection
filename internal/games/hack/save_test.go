package hack

import (
	"errors"
	"math/rand"
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/records"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/session"
)

// clearLevel reads the first level's secret: the level is cleared.
func clearLevel(g *Game) {
	at(g, 0, "core")
	run(g, "cat lantern.txt")
}

func TestACheckpointResumesAtTheNextLevelWithItsScoreAndTime(t *testing.T) {
	g := newTestGame(Won)
	g.ticks = 600
	clearLevel(g)
	cp, ok := g.Checkpoint()
	if !ok || cp != (Checkpoint{Level: 1, Completed: g.Score(), Ticks: g.Ticks()}) {
		t.Fatalf("a cleared level checkpoints the next one, got %+v ok %v", cp, ok)
	}
	resumed := NewGame(Config{Rand: rand.New(rand.NewSource(1)), Resume: &cp})
	if resumed.Level() != 1 || resumed.Score() != cp.Completed+ScoreStart || resumed.Elapsed() != g.Elapsed() || resumed.Owned("core") {
		t.Fatalf("resumed at level %d score %d elapsed %v", resumed.Level()+1, resumed.Score(), resumed.Elapsed())
	}
}

func TestOnlyAClearedLevelCheckpoints(t *testing.T) {
	if _, ok := newTestGame(Won).Checkpoint(); ok {
		t.Fatalf("a level under way has no checkpoint")
	}
}

func TestAnImpossibleCheckpointStartsANewGame(t *testing.T) {
	for _, cp := range []Checkpoint{{Level: 0}, {Level: len(levels)}, {Level: 2, Completed: -1}, {Level: 2, Ticks: -1}} {
		g := NewGame(Config{Rand: rand.New(rand.NewSource(1)), Resume: &cp})
		if g.Level() != 0 || g.Score() != ScoreStart || g.Ticks() != 0 {
			t.Fatalf("%+v must start a new game", cp)
		}
	}
}

func savedSession(t *testing.T, resume bool) (*session.Session, records.Slot) {
	t.Helper()
	saves, _ := records.OpenSaves("")
	slot := saves.Slot("hack")
	return savedSessionOf(t, slot, resume), slot
}

func savedSessionOf(t *testing.T, slot Slot, resume bool) *session.Session {
	t.Helper()
	m := NewModule(nil, nil, nil, rand.New(rand.NewSource(1))).WithSave(slot, resume)
	return session.New(m.rounds())
}

func TestClearingALevelSavesTheRunAndTheEndDeletesIt(t *testing.T) {
	s, slot := savedSession(t, false)
	g := s.Round().(savedRound).Game
	g.newBreach = func(BreachKind) Breach { return &stubBreach{outcome: Won} }
	at(g, 0, "core")
	update(t, s, typed("cat lantern.txt"), press(input.ActionSelect))
	var cp Checkpoint
	if ok, _ := slot.Load(&cp); !ok || cp.Level != 1 {
		t.Fatalf("clearing a level saves the run, got %+v ok %v", cp, ok)
	}
	g.raise(100)
	update(t, s, press())
	if slot.Exists() {
		t.Fatalf("a run ended by the trace deletes the save")
	}
}

func TestContinueResumesTheSavedRunAndNewGameDeletesIt(t *testing.T) {
	saves, _ := records.OpenSaves("")
	slot := saves.Slot("hack")
	slot.Store(Checkpoint{Level: 3, Completed: 2500, Ticks: 60})
	s := savedSessionOf(t, slot, true)
	if g := s.Round().(savedRound).Game; g.Level() != 3 || g.Score() != 2500+ScoreStart {
		t.Fatalf("continue resumes at level %d, score %d", g.Level()+1, g.Score())
	}
	if !slot.Exists() {
		t.Fatalf("continuing keeps the save until the next checkpoint or the end")
	}
	s = savedSessionOf(t, slot, false)
	if g := s.Round().(savedRound).Game; g.Level() != 0 || slot.Exists() {
		t.Fatalf("a new game starts at level 1 and deletes the save")
	}
}

func TestRestartStartsANewGame(t *testing.T) {
	saves, _ := records.OpenSaves("")
	slot := saves.Slot("hack")
	slot.Store(Checkpoint{Level: 3, Completed: 2500})
	s := savedSessionOf(t, slot, true)
	update(t, s, press(input.ActionPause), press(input.ActionDown), press(input.ActionSelect))
	if g := s.Round().(savedRound).Game; g.Level() != 0 || slot.Exists() {
		t.Fatalf("restart begins a new game at level 1, level %d, saved %v", g.Level()+1, slot.Exists())
	}
}

// brokenSlot never saves.
type brokenSlot struct{}

func (brokenSlot) Load(any) (bool, error) { return false, nil }
func (brokenSlot) Store(any) error        { return errors.New("disk full") }
func (brokenSlot) Clear() error           { return nil }

func TestAFailedSaveIsToldAtTheEnd(t *testing.T) {
	s := savedSessionOf(t, brokenSlot{}, false)
	g := s.Round().(savedRound).Game
	at(g, 0, "core")
	update(t, s, typed("cat lantern.txt"), press(input.ActionSelect))
	g.raise(100)
	update(t, s, press())
	m, _ := s.Overlay()
	var text []string
	for _, sec := range m.Sections() {
		text = append(text, sec.Lines...)
	}
	if !strings.Contains(strings.Join(text, "\n"), "No se pudo guardar") {
		t.Fatalf("the end must say the run was not saved: %+v", m.Sections())
	}
}
