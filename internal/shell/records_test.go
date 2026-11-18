package shell

import (
	"context"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/records"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/games/hack"
)

func stores(t *testing.T) (*records.Book, *records.Saves) {
	t.Helper()
	book, _ := records.Open("")
	saves, _ := records.OpenSaves("")
	return book, saves
}

func runeKey(r rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func send(m Model, msgs ...tea.Msg) Model {
	for _, msg := range msgs {
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	return m
}

var ansi = regexp.MustCompile("\x1b\\[[0-9;]*m")

func plain(view string) string { return ansi.ReplaceAllString(view, "") }

// capture replaces every starter with one that reports how it was launched.
func capture(m Model) (Model, chan launch) {
	got := make(chan launch, 1)
	for i := range m.games {
		m.games[i].starter = func(_ context.Context, _ input.Provider, _ render.Screen, l launch) error {
			got <- l
			return nil
		}
	}
	return m, got
}

func cursorOn(m Model, name string) Model {
	for i, g := range m.games {
		if g.Name == name {
			m.cursor = i
		}
	}
	return m
}

func TestRShowsTheHighScoresGameByGame(t *testing.T) {
	book, saves := stores(t)
	book.Add("snake", records.Entry{Name: "ana", Score: 50})
	m := send(New(WithRecords(book, saves)), runeKey('r'))
	view := plain(m.View())
	if m.state != stateScores || !strings.Contains(view, "Snake") || !strings.Contains(view, "1. ana") {
		t.Fatalf("r opens the selected game's high scores:\n%s", view)
	}
	m = send(m, tea.KeyMsg{Type: tea.KeyRight})
	if view := plain(m.View()); !strings.Contains(view, "Sokoban") || !strings.Contains(view, "none yet") {
		t.Fatalf("right shows the next game's table:\n%s", view)
	}
	m = send(m, tea.KeyMsg{Type: tea.KeyLeft}, tea.KeyMsg{Type: tea.KeyLeft})
	if view := plain(m.View()); !strings.Contains(view, "Hack") {
		t.Fatalf("left from the first game goes round to the last:\n%s", view)
	}
	for _, back := range []tea.KeyMsg{{Type: tea.KeyEscape}, runeKey('q'), runeKey('r')} {
		m := send(m, back)
		if m.state != stateMenu {
			t.Fatalf("%v goes back to the menu", back)
		}
		if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown}); cmd != nil {
			t.Fatalf("the menu still works")
		}
	}
}

func TestTheMenuOffersTheHighScores(t *testing.T) {
	if !strings.Contains(plain(New().menuView()), "r scores") {
		t.Fatalf("the menu hints must name the high scores key")
	}
}

func TestTheHighScoresFitTheTerminal(t *testing.T) {
	book, saves := stores(t)
	for i := 0; i < records.Size; i++ {
		book.Add("sokoban", records.Entry{Name: strings.Repeat("W", records.MaxName), Score: 99999999, Seconds: 3599})
	}
	m := send(cursorOn(New(WithRecords(book, saves)), "Sokoban"), runeKey('r'))
	view := plain(m.View())
	if lines := strings.Split(view, "\n"); len(lines) > m.height || !strings.Contains(view, "59:59") {
		t.Fatalf("%d lines:\n%s", len(lines), view)
	}
	for _, line := range strings.Split(view, "\n") {
		if w := len([]rune(line)); w != len([]rune(strings.Split(view, "\n")[0])) {
			t.Fatalf("every row of the box has the same width:\n%s", view)
		}
	}
}

func TestEveryGameIsLaunchedWithItsOwnBoard(t *testing.T) {
	book, saves := stores(t)
	m, got := capture(New(WithRecords(book, saves)))
	for i, g := range m.games {
		m.cursor = i
		m.startGame(g)
		l := <-got
		l.Board.Add(records.Entry{Name: g.Name, Score: 1})
		if top := book.Top(strings.ToLower(g.Name)); len(top) != 1 || top[0].Name != g.Name {
			t.Fatalf("%s must get its own table", g.Name)
		}
	}
}

func TestHackWithASavedRunAsksToContinueOrStartAgain(t *testing.T) {
	book, saves := stores(t)
	saves.Slot("hack").Store(hack.Checkpoint{Level: 2, Completed: 1500, Ticks: 60})
	m, got := capture(New(WithRecords(book, saves)))
	m = cursorOn(m, "Hack")
	if view := plain(m.menuView()); !strings.Contains(view, "Hack · saved run") {
		t.Fatalf("the menu tells that Hack has a saved run:\n%s", view)
	}
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	view := plain(m.View())
	if m.state != stateResume || !strings.Contains(view, "Continue (level 3)") || !strings.Contains(view, "New game") {
		t.Fatalf("enter on Hack with a save asks first:\n%s", view)
	}
	if send(m, tea.KeyMsg{Type: tea.KeyEscape}).state != stateMenu {
		t.Fatalf("esc goes back to the menu")
	}
	send(m, tea.KeyMsg{Type: tea.KeyEnter})
	if l := <-got; !l.Resume || l.Save != saves.Slot("hack") {
		t.Fatalf("continue resumes the saved run: %+v", l)
	}
	send(m, tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter})
	if l := <-got; l.Resume || l.Save != saves.Slot("hack") {
		t.Fatalf("new game starts over, still saving: %+v", l)
	}
}

func TestHackWithoutASaveStartsAtOnce(t *testing.T) {
	book, saves := stores(t)
	m, got := capture(New(WithRecords(book, saves)))
	m = send(cursorOn(m, "Hack"), tea.KeyMsg{Type: tea.KeyEnter})
	if l := <-got; m.state != statePlaying || l.Resume || l.Save != saves.Slot("hack") {
		t.Fatalf("with no save Hack starts a new game: %+v", l)
	}
	if strings.Contains(plain(m.menuView()), "saved run") {
		t.Fatalf("no save, no mark")
	}
}

func TestOnlyHackIsSaved(t *testing.T) {
	book, saves := stores(t)
	m, got := capture(New(WithRecords(book, saves)))
	for i, g := range m.games {
		m.cursor = i
		m.startGame(g)
		if l := <-got; (l.Save != records.Slot{}) != (g.Name == "Hack") {
			t.Fatalf("%s: save slot %+v", g.Name, l.Save)
		}
	}
}

func TestPlayHackWithASavedRunAsksFirst(t *testing.T) {
	book, saves := stores(t)
	saves.Slot("hack").Store(hack.Checkpoint{Level: 1})
	if m := NewWithGame("hack", WithRecords(book, saves)); m.state != stateResume {
		t.Fatalf("play hack with a save asks to continue, state %v", m.state)
	}
}
