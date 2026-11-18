package session

import (
	"errors"
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/records"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"
)

// rankedRound is a fakeRound with a result for the high score table.
type rankedRound struct {
	fakeRound
	score int
}

func (r *rankedRound) Result() records.Entry { return records.Entry{Score: r.score} }

func newRankedSession(t *testing.T, scores ...int) (*Session, *records.Book) {
	t.Helper()
	book, err := records.Open("")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	s := New(func() Round {
		r := &rankedRound{fakeRound: fakeRound{endOn: 1}, score: scores[min(n, len(scores)-1)]}
		n++
		return r
	}, WithBoard(book.Board("test")))
	return s, book
}

func typed(text string) input.FrameInput {
	f := input.FrameInput{Text: []rune(text)}
	for _, r := range text {
		if a := input.RuneAction(r); a != "" {
			f.Actions = append(f.Actions, a)
		}
	}
	return f
}

func lines(m overlay.Model) string {
	var all []string
	for _, s := range m.Sections() {
		all = append(all, s.Lines...)
	}
	return strings.Join(all, "\n")
}

func TestAHighScoreAsksForANameBeforeTheEnd(t *testing.T) {
	s, book := newRankedSession(t, 120)
	update(t, s, press())
	m, open := s.Overlay()
	if !open || m.Title() != TitleHighScore || !strings.Contains(lines(m), "Score: 120") {
		t.Fatalf("a result that enters the table asks for a name first: %q\n%s", m.Title(), lines(m))
	}
	// Letters that stand for actions are only text here: q does not quit,
	// p does not pause and space does not save.
	update(t, s, typed("p qa"), press(input.ActionBackspace), typed("ux"))
	if m, _ := s.Overlay(); !strings.Contains(lines(m), "Name: p qux_") || s.ShouldQuit() {
		t.Fatalf("typing edits the name:\n%s", lines(m))
	}
	update(t, s, press(input.ActionSelect))
	if top := book.Top("test"); len(top) != 1 || top[0] != (records.Entry{Name: "p qux", Score: 120}) {
		t.Fatalf("enter saves the result, top %+v", top)
	}
	m, _ = s.Overlay()
	if m.Title() != "Game Over" || !strings.Contains(lines(m), "Score: 7") || !strings.Contains(lines(m), "1. p qux") {
		t.Fatalf("then the end shows the summary and the table:\n%s", lines(m))
	}
}

func TestANameStopsAtItsLongestLength(t *testing.T) {
	s, book := newRankedSession(t, 5)
	update(t, s, press(), typed(strings.Repeat("x", records.MaxName+3)))
	if m, _ := s.Overlay(); !strings.Contains(lines(m), "Name: "+strings.Repeat("x", records.MaxName)+"_") {
		t.Fatalf("typing stops at the longest name:\n%s", lines(m))
	}
	update(t, s, press(input.ActionSelect))
	if got := book.Top("test")[0].Name; len(got) != records.MaxName {
		t.Fatalf("name %q", got)
	}
}

func TestEscSavesTheNameAsTyped(t *testing.T) {
	s, book := newRankedSession(t, 5)
	update(t, s, press(), press(input.ActionPause))
	if top := book.Top("test"); len(top) != 1 || top[0].Name != records.DefaultName {
		t.Fatalf("esc saves the result too, top %+v", top)
	}
	if m, _ := s.Overlay(); m.Title() != "Game Over" {
		t.Fatalf("then the end shows")
	}
}

func TestAResultOffTheTableGoesStraightToTheEndWithTheTable(t *testing.T) {
	s, book := newRankedSession(t, 0)
	book.Add("test", records.Entry{Name: "ana", Score: 50})
	update(t, s, press())
	m, _ := s.Overlay()
	if m.Title() != "Game Over" || !strings.Contains(lines(m), "1. ana") {
		t.Fatalf("no name is asked, and the end still shows the table:\n%s", lines(m))
	}
}

func TestTheTableMarksTheNewResult(t *testing.T) {
	s, book := newRankedSession(t, 30)
	book.Add("test", records.Entry{Name: "ana", Score: 50})
	book.Add("test", records.Entry{Name: "bo", Score: 10})
	update(t, s, press(), typed("cy"), press(input.ActionSelect))
	m, _ := s.Overlay()
	table := lines(m)
	if !strings.Contains(table, "2. cy") || strings.Count(table, "◂") != 1 || !strings.Contains(table, "30 ◂\n3. bo") {
		t.Fatalf("the new result is marked:\n%s", table)
	}
}

func TestOnlyTheLastResultIsMarked(t *testing.T) {
	s, _ := newRankedSession(t, 10, 0)
	update(t, s, press(), press(input.ActionSelect), press(input.ActionSelect), press())
	if m, _ := s.Overlay(); m.Title() != "Game Over" || strings.Contains(lines(m), "◂") {
		t.Fatalf("a result off the table marks nothing:\n%s", lines(m))
	}
}

func TestRestartAfterAHighScoreAsksAgain(t *testing.T) {
	s, _ := newRankedSession(t, 10, 20)
	update(t, s, press(), press(input.ActionSelect), press(input.ActionSelect), press())
	if m, _ := s.Overlay(); m.Title() != TitleHighScore {
		t.Fatalf("a better result after a restart asks for a name, got %q", m.Title())
	}
}

func TestARoundWithoutAResultOrASessionWithoutABoardIsUnchanged(t *testing.T) {
	s, _ := newTestSession(1)
	update(t, s, press())
	if m, _ := s.Overlay(); m.Title() != "Game Over" || lines(m) != "Score: 7" {
		t.Fatalf("no table without a board:\n%s", lines(m))
	}
	book, _ := records.Open("")
	s = New(func() Round { return &fakeRound{endOn: 1} }, WithBoard(book.Board("test")))
	update(t, s, press())
	if m, _ := s.Overlay(); m.Title() != "Game Over" || lines(m) != "Score: 7" {
		t.Fatalf("no table for a round without a result:\n%s", lines(m))
	}
}

func TestTheNameEntryFitsTheScreen(t *testing.T) {
	s, _ := newRankedSession(t, 99999999)
	update(t, s, press(), typed(strings.Repeat("W", records.MaxName)))
	screen := render.NewBufferScreen(80, 24)
	s.Render(nil, screen)
	if !strings.Contains(screen.Snapshot(), "Name: "+strings.Repeat("W", records.MaxName)) {
		t.Fatalf("the name must show whole:\n%s", screen.Snapshot())
	}
}

// failingBoard takes every result and never saves it.
type failingBoard struct{}

func (failingBoard) Top() []records.Entry           { return nil }
func (failingBoard) Qualifies(records.Entry) bool   { return true }
func (failingBoard) Add(records.Entry) (int, error) { return 1, errors.New("disk full") }

func TestAFailedSaveIsToldAtTheEnd(t *testing.T) {
	s := New(func() Round { return &rankedRound{fakeRound: fakeRound{endOn: 1}, score: 5} }, WithBoard(failingBoard{}))
	update(t, s, press(), press(input.ActionSelect))
	if m, _ := s.Overlay(); m.Title() != "Game Over" || !strings.Contains(lines(m), "could not save") {
		t.Fatalf("the end must say the table was not saved:\n%s", lines(m))
	}
}
