package shell

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// ─── ChannelInput ──────────────────────────────────────────────────────────────

func TestChannelInput_SendAndPoll(t *testing.T) {
	ci := NewChannelInput(4)
	ci.Send(input.ActionUp)
	ci.Send(input.ActionDown)

	fi := ci.Poll()
	if len(fi.Actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(fi.Actions))
	}
	if fi.Actions[0] != input.ActionUp {
		t.Fatalf("expected first action up, got %q", fi.Actions[0])
	}
	if fi.Actions[1] != input.ActionDown {
		t.Fatalf("expected second action down, got %q", fi.Actions[1])
	}
}

func TestChannelInput_PollEmpty(t *testing.T) {
	ci := NewChannelInput(4)
	fi := ci.Poll()
	if len(fi.Actions) != 0 {
		t.Fatalf("expected 0 actions from empty channel, got %d", len(fi.Actions))
	}
}

func TestChannelInput_NonBlockingWhenFull(t *testing.T) {
	ci := NewChannelInput(2)
	ci.Send(input.ActionUp)
	ci.Send(input.ActionDown)
	// Channel is full now; this send should drop
	ci.Send(input.ActionLeft)

	fi := ci.Poll()
	if len(fi.Actions) != 2 {
		t.Fatalf("expected 2 actions (drop when full), got %d", len(fi.Actions))
	}
}

func TestChannelInput_DrainsAllOnPoll(t *testing.T) {
	ci := NewChannelInput(4)
	ci.Send(input.ActionQuit)
	ci.Poll() // drain

	// After drain, channel should be empty
	fi := ci.Poll()
	if len(fi.Actions) != 0 {
		t.Fatalf("expected 0 after drain, got %d", len(fi.Actions))
	}
}

// ─── mapKeyToAction ─────────────────────────────────────────────────────────

func TestMapKeyToAction(t *testing.T) {
	tests := []struct {
		key  tea.KeyType
		want input.Action
	}{
		{tea.KeyUp, input.ActionUp},
		{tea.KeyDown, input.ActionDown},
		{tea.KeyLeft, input.ActionLeft},
		{tea.KeyRight, input.ActionRight},
		{tea.KeyEnter, input.Action("")},
		{tea.KeyEscape, input.Action("")},
		{tea.KeyCtrlC, input.Action("")},
		{0, input.Action("")}, // unknown
	}

	for _, tt := range tests {
		got := mapKeyToAction(tt.key)
		if got != tt.want {
			t.Errorf("mapKeyToAction(%v): expected %q, got %q", tt.key, tt.want, got)
		}
	}
}

// ─── Model ──────────────────────────────────────────────────────────────────

func TestNew_CreatesMenuStateWithDefaults(t *testing.T) {
	m := New()
	if m.state != stateMenu {
		t.Fatalf("expected stateMenu, got %v", m.state)
	}
	if len(m.games) != 11 {
		t.Fatalf("expected 11 games, got %d", len(m.games))
	}
	if m.games[0].Name != "Snake" {
		t.Fatalf("expected game name 'Snake', got %q", m.games[0].Name)
	}
	if m.games[1].Name != "Sokoban" {
		t.Fatalf("expected game name 'Sokoban', got %q", m.games[1].Name)
	}
	if m.games[2].Name != "Runner" {
		t.Fatalf("expected game name 'Runner', got %q", m.games[2].Name)
	}
	if m.games[3].Name != "Drone" {
		t.Fatalf("expected game name 'Drone', got %q", m.games[3].Name)
	}
	if m.games[4].Name != "Breach" {
		t.Fatalf("expected game name 'Breach', got %q", m.games[4].Name)
	}
	if m.games[5].Name != "Botnet" {
		t.Fatalf("expected game name 'Botnet', got %q", m.games[5].Name)
	}
	if m.games[6].Name != "Leak" {
		t.Fatalf("expected game name 'Leak', got %q", m.games[6].Name)
	}
	if m.games[7].Name != "Packet" {
		t.Fatalf("expected game name 'Packet', got %q", m.games[7].Name)
	}
	if m.games[8].Name != "Debris" {
		t.Fatalf("expected game name 'Debris', got %q", m.games[8].Name)
	}
	if m.games[9].Name != "Rack" {
		t.Fatalf("expected game name 'Rack', got %q", m.games[9].Name)
	}
	if m.games[10].Name != "Defrag" {
		t.Fatalf("expected game name 'Defrag', got %q", m.games[10].Name)
	}
	if m.cursor != 0 {
		t.Fatalf("expected cursor 0, got %d", m.cursor)
	}
	if m.width != 80 || m.height != 24 {
		t.Fatalf("expected default size 80x24, got %dx%d", m.width, m.height)
	}
}

func TestNewWithGame_KnownGameSetsPlayingState(t *testing.T) {
	m := NewWithGame("snake")
	if m.state != statePlaying {
		t.Fatalf("expected statePlaying, got %v", m.state)
	}
	if m.gameCtrl == nil {
		t.Fatal("expected gameCtrl to be set")
	}
	if m.chInput == nil {
		t.Fatal("expected chInput to be set")
	}
}

func TestNewWithGame_RunnerSetsPlayingState(t *testing.T) {
	m := NewWithGame("runner")
	if m.state != statePlaying {
		t.Fatalf("expected statePlaying, got %v", m.state)
	}
	if m.games[m.cursor].Name != "Runner" {
		t.Fatalf("expected cursor on Runner, got %q", m.games[m.cursor].Name)
	}
	m.gameCtrl.cancel()
}

func TestNewWithGame_DroneSetsPlayingState(t *testing.T) {
	m := NewWithGame("drone")
	if m.state != statePlaying {
		t.Fatalf("expected statePlaying, got %v", m.state)
	}
	if m.games[m.cursor].Name != "Drone" {
		t.Fatalf("expected cursor on Drone, got %q", m.games[m.cursor].Name)
	}
	m.gameCtrl.cancel()
}

func TestNewWithGame_BreachSetsPlayingState(t *testing.T) {
	m := NewWithGame("breach")
	if m.state != statePlaying {
		t.Fatalf("expected statePlaying, got %v", m.state)
	}
	if m.games[m.cursor].Name != "Breach" {
		t.Fatalf("expected cursor on Breach, got %q", m.games[m.cursor].Name)
	}
	m.gameCtrl.cancel()
}

func TestNewWithGame_BotnetSetsPlayingState(t *testing.T) {
	m := NewWithGame("botnet")
	if m.state != statePlaying {
		t.Fatalf("expected statePlaying, got %v", m.state)
	}
	if m.games[m.cursor].Name != "Botnet" {
		t.Fatalf("expected cursor on Botnet, got %q", m.games[m.cursor].Name)
	}
	m.gameCtrl.cancel()
}

func TestNewWithGame_LeakSetsPlayingState(t *testing.T) {
	m := NewWithGame("leak")
	if m.state != statePlaying {
		t.Fatalf("expected statePlaying, got %v", m.state)
	}
	if m.games[m.cursor].Name != "Leak" {
		t.Fatalf("expected cursor on Leak, got %q", m.games[m.cursor].Name)
	}
	m.gameCtrl.cancel()
}

func TestNewWithGame_PacketSetsPlayingState(t *testing.T) {
	m := NewWithGame("packet")
	if m.state != statePlaying {
		t.Fatalf("expected statePlaying, got %v", m.state)
	}
	if m.games[m.cursor].Name != "Packet" {
		t.Fatalf("expected cursor on Packet, got %q", m.games[m.cursor].Name)
	}
	m.gameCtrl.cancel()
}

func TestNewWithGame_DebrisSetsPlayingState(t *testing.T) {
	m := NewWithGame("debris")
	if m.state != statePlaying {
		t.Fatalf("expected statePlaying, got %v", m.state)
	}
	if m.games[m.cursor].Name != "Debris" {
		t.Fatalf("expected cursor on Debris, got %q", m.games[m.cursor].Name)
	}
	m.gameCtrl.cancel()
}

func TestNewWithGame_RackSetsPlayingState(t *testing.T) {
	m := NewWithGame("rack")
	if m.state != statePlaying {
		t.Fatalf("expected statePlaying, got %v", m.state)
	}
	if m.games[m.cursor].Name != "Rack" {
		t.Fatalf("expected cursor on Rack, got %q", m.games[m.cursor].Name)
	}
	m.gameCtrl.cancel()
}

func TestNewWithGame_DefragSetsPlayingState(t *testing.T) {
	m := NewWithGame("defrag")
	if m.state != statePlaying {
		t.Fatalf("expected statePlaying, got %v", m.state)
	}
	if m.games[m.cursor].Name != "Defrag" {
		t.Fatalf("expected cursor on Defrag, got %q", m.games[m.cursor].Name)
	}
	m.gameCtrl.cancel()
}

func TestMenuViewNamesTheGameEachOneIsLike(t *testing.T) {
	view := New().menuView()
	for _, want := range []string{
		"Runner · Chrome Dino-like",
		"Drone · Flappy Bird-like",
		"Breach · Breakout-like",
		"Botnet · Space Invaders-like",
		"Leak · Pac-Man-like",
		"Packet · Subway Surfers-like",
		"Debris · Asteroids-like",
		"Rack · Donkey Kong-like",
		"Defrag · Tetris-like",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("menu missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Snake ·") || strings.Contains(view, "Sokoban ·") {
		t.Fatalf("Snake and Sokoban already name their genre and need no reference")
	}
}

func TestMenuViewFitsDefaultTerminalWithAllGames(t *testing.T) {
	m := New()
	if lines := strings.Count(m.menuView(), "\n") + 1; lines > m.height {
		t.Fatalf("menu has %d lines, terminal is %d", lines, m.height)
	}
}

func TestMenuViewFitsTheWholeRoadmap(t *testing.T) {
	m := New()
	for len(m.games) < 12 {
		m.games = append(m.games, gameEntry{Name: "Next", Description: "Planned"})
	}
	if lines := strings.Count(m.menuView(), "\n") + 1; lines > m.height {
		t.Fatalf("menu with 12 games has %d lines, terminal is %d", lines, m.height)
	}
}

func TestMenuViewDescribesOnlyTheSelectedGame(t *testing.T) {
	m := New()
	for i, selected := range m.games {
		m.cursor = i
		view := m.menuView()
		for _, g := range m.games {
			if got := strings.Contains(view, g.Description); got != (g.Name == selected.Name) {
				t.Fatalf("cursor on %s: description of %s shown = %v", selected.Name, g.Name, got)
			}
			if !strings.Contains(view, g.Name) {
				t.Fatalf("cursor on %s: %s missing from the list", selected.Name, g.Name)
			}
		}
	}
}

func TestNewWithGame_UnknownGameLeavesMenuState(t *testing.T) {
	m := NewWithGame("unknown")
	if m.state != stateMenu {
		t.Fatalf("expected stateMenu for unknown game, got %v", m.state)
	}
}

func TestNewWithGame_CaseInsensitive(t *testing.T) {
	m := NewWithGame("SNAKE")
	if m.state != statePlaying {
		t.Fatalf("expected statePlaying for 'SNAKE', got %v", m.state)
	}
}

func TestInit_ReturnsNilWhenMenu(t *testing.T) {
	m := New()
	cmd := m.Init()
	if cmd != nil {
		t.Fatal("expected nil cmd for menu state")
	}
}

func TestInit_ReturnsCmdWhenPlaying(t *testing.T) {
	m := NewWithGame("snake")
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected non-nil cmd for playing state")
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := New()
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model := m2.(Model)
	if model.width != 100 {
		t.Fatalf("expected width 100, got %d", model.width)
	}
	if model.height != 30 {
		t.Fatalf("expected height 30, got %d", model.height)
	}
}

// ─── Menu keyboard ──────────────────────────────────────────────────────────

func TestUpdate_MenuKeyEnter(t *testing.T) {
	m := New()
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := m2.(Model)
	if model.state != statePlaying {
		t.Fatalf("expected statePlaying after Enter, got %v", model.state)
	}
	if model.gameCtrl == nil {
		t.Fatal("expected gameCtrl after starting game")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd (awaitGameEvent) after Enter")
	}
}

func TestUpdate_MenuCtrlCQuits(t *testing.T) {
	m := New()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit cmd for Ctrl+C")
	}
}

func TestUpdate_MenuEscapeQuits(t *testing.T) {
	m := New()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected quit cmd for Escape")
	}
}

func TestUpdate_MenuQQuits(t *testing.T) {
	m := New()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit cmd for 'q' in menu")
	}
}

func TestUpdate_MenuQUpperQuits(t *testing.T) {
	m := New()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Q'}})
	if cmd == nil {
		t.Fatal("expected quit cmd for 'Q'")
	}
}

// ─── Playing state machine ──────────────────────────────────────────────────

func TestUpdate_FrameMsgInPlaying(t *testing.T) {
	m := NewWithGame("snake")
	m2, cmd := m.Update(FrameMsg{Frame: "frame-content"})
	model := m2.(Model)
	if model.frame != "frame-content" {
		t.Fatalf("expected frame to be set, got %q", model.frame)
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd (awaitGameEvent) after FrameMsg")
	}
}

func TestUpdate_FrameMsgInMenuIgnored(t *testing.T) {
	m := New()
	m2, _ := m.Update(FrameMsg{Frame: "should-be-ignored"})
	if m2.(Model).frame != "" {
		t.Fatal("expected frame not to be set when in menu state")
	}
}

func TestUpdate_GameOverMsgReturnsToMenu(t *testing.T) {
	m := NewWithGame("snake")
	m2, _ := m.Update(GameOverMsg{Err: nil})
	model := m2.(Model)
	if model.state != stateMenu {
		t.Fatalf("expected stateMenu after game over, got %v", model.state)
	}
	if model.gameCtrl != nil {
		t.Fatal("expected gameCtrl to be nil after game over")
	}
}

func TestAwaitGameEventReturnsGameOverWhenDoneReadyAndFrameClosed(t *testing.T) {
	frameCh := make(chan string)
	close(frameCh)
	doneCh := make(chan error, 1)
	doneCh <- nil

	m := New()
	m.state = statePlaying
	m.gameCtrl = &gameController{
		cancel:  func() {},
		frameCh: frameCh,
		doneCh:  doneCh,
	}

	msg := m.awaitGameEvent()()
	if msg == nil {
		t.Fatal("expected GameOverMsg when done channel is ready and frame channel is closed")
	}
	if _, ok := msg.(GameOverMsg); !ok {
		t.Fatalf("expected GameOverMsg, got %T", msg)
	}
}

// ─── Playing key routing ────────────────────────────────────────────────────

func TestUpdate_PlayingArrowKeys(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(4)

	cases := []struct {
		key  tea.KeyType
		want input.Action
	}{
		{tea.KeyUp, input.ActionUp},
		{tea.KeyDown, input.ActionDown},
		{tea.KeyLeft, input.ActionLeft},
		{tea.KeyRight, input.ActionRight},
	}

	for _, tc := range cases {
		updated, _ := m.Update(tea.KeyMsg{Type: tc.key})
		m = updated.(Model)
	}

	fi := m.chInput.Poll()
	if len(fi.Actions) != 4 {
		t.Fatalf("expected 4 actions, got %d", len(fi.Actions))
	}
	for i, tc := range cases {
		if fi.Actions[i] != tc.want {
			t.Errorf("action %d: expected %q, got %q", i, tc.want, fi.Actions[i])
		}
	}
}

func TestUpdate_PlayingWASD(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(4)

	cases := []struct {
		r    rune
		want input.Action
	}{
		{'w', input.ActionUp},
		{'a', input.ActionLeft},
		{'s', input.ActionDown},
		{'d', input.ActionRight},
	}

	for _, tc := range cases {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tc.r}})
		m = updated.(Model)
	}

	fi := m.chInput.Poll()
	if len(fi.Actions) != 4 {
		t.Fatalf("expected 4 actions, got %d", len(fi.Actions))
	}
	for i, tc := range cases {
		if fi.Actions[i] != tc.want {
			t.Errorf("action %d: expected %q, got %q", i, tc.want, fi.Actions[i])
		}
	}
}

func TestUpdate_PlayingFSendsFire(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(4)

	for _, r := range []rune{'f', 'F'} {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}
	fi := m.chInput.Poll()
	if len(fi.Actions) != 2 || fi.Actions[0] != input.ActionFire || fi.Actions[1] != input.ActionFire {
		t.Fatalf("expected ActionFire for 'f' and 'F', got %v", fi.Actions)
	}
}

func TestUpdate_PlayingGSendsGuides(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(4)

	for _, r := range []rune{'g', 'G'} {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}
	fi := m.chInput.Poll()
	if len(fi.Actions) != 2 || fi.Actions[0] != input.ActionGuides || fi.Actions[1] != input.ActionGuides {
		t.Fatalf("expected ActionGuides for 'g' and 'G', got %v", fi.Actions)
	}
}

func TestUpdate_PlayingWASDUpper(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(4)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'W'}})
	m = updated.(Model)
	fi := m.chInput.Poll()
	if len(fi.Actions) != 1 || fi.Actions[0] != input.ActionUp {
		t.Fatalf("expected ActionUp for 'W', got %v", fi.Actions)
	}
}

func TestUpdate_PlayingSendsTypedTextAlongsideActions(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(8)

	for _, msg := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'l', 's'}},
		{Type: tea.KeySpace, Runes: []rune{' '}},
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
		{Type: tea.KeyBackspace},
	} {
		updated, _ := m.Update(msg)
		m = updated.(Model)
	}
	fi := m.chInput.Poll()
	if string(fi.Text) != "ls q" {
		t.Fatalf("text = %q, want %q", string(fi.Text), "ls q")
	}
	want := []input.Action{input.ActionDown, input.ActionSelect, input.ActionQuit, input.ActionBackspace} // the s of "ls" is down
	if len(fi.Actions) != len(want) {
		t.Fatalf("actions = %v, want %v", fi.Actions, want)
	}
	for i := range want {
		if fi.Actions[i] != want[i] {
			t.Fatalf("actions = %v, want %v", fi.Actions, want)
		}
	}
}

func TestUpdate_PlayingEscapeSendsPause(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(4)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(Model)
	fi := m.chInput.Poll()
	if len(fi.Actions) != 1 || fi.Actions[0] != input.ActionPause {
		t.Fatalf("expected ActionPause for Escape in playing, got %v", fi.Actions)
	}
}

func TestUpdate_PlayingEnterSendsSelect(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(4)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	fi := m.chInput.Poll()
	if len(fi.Actions) != 1 || fi.Actions[0] != input.ActionSelect {
		t.Fatalf("expected ActionSelect for Enter, got %v", fi.Actions)
	}
}

func TestUpdate_PlayingPSendsPause(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(4)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = updated.(Model)
	fi := m.chInput.Poll()
	if len(fi.Actions) != 1 || fi.Actions[0] != input.ActionPause {
		t.Fatalf("expected ActionPause for 'p', got %v", fi.Actions)
	}
}

func TestUpdate_PlayingSpaceSendsSelect(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(4)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(Model)
	fi := m.chInput.Poll()
	if len(fi.Actions) != 1 || fi.Actions[0] != input.ActionSelect {
		t.Fatalf("expected ActionSelect for Space, got %v", fi.Actions)
	}
}

func TestUpdate_PlayingQSendsQuit(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(4)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(Model)
	fi := m.chInput.Poll()
	if len(fi.Actions) != 1 || fi.Actions[0] != input.ActionQuit {
		t.Fatalf("expected ActionQuit for 'q', got %v", fi.Actions)
	}
}

func TestUpdate_PlayingCtrlCSendsQuit(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.chInput = NewChannelInput(4)
	m.gameCtrl = &gameController{cancel: func() {}}

	upd, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	_ = upd.(Model)
	if cmd != nil {
		t.Fatal("expected nil cmd for Ctrl+C in playing")
	}
	fi := m.chInput.Poll()
	if len(fi.Actions) != 1 || fi.Actions[0] != input.ActionQuit {
		t.Fatalf("expected ActionQuit for Ctrl+C, got %v", fi.Actions)
	}
}

// ─── FrameBroadcaster ───────────────────────────────────────────────────────

func TestFrameBroadcaster_PresentSendsSnapshot(t *testing.T) {
	ch := make(chan string, 1)
	bs := render.NewBufferScreen(10, 5)
	fb := &frameBroadcaster{
		BufferScreen: bs,
		frameCh:      ch,
	}

	bs.SetCell(3, 2, render.Cell{Rune: 'X'})
	err := fb.Present()
	if err != nil {
		t.Fatalf("Present returned error: %v", err)
	}

	frame := <-ch
	if !strings.Contains(frame, "X") {
		t.Fatalf("expected frame to contain 'X', got %q", frame)
	}
}

func TestFrameBroadcaster_SendsColours(t *testing.T) {
	ch := make(chan string, 1)
	bs := render.NewBufferScreen(4, 1)
	fb := &frameBroadcaster{BufferScreen: bs, frameCh: ch}

	bs.SetCell(1, 0, render.Cell{Rune: 'X', Color: render.ColorYellow})
	if err := fb.Present(); err != nil {
		t.Fatalf("Present returned error: %v", err)
	}
	if frame := <-ch; !strings.Contains(frame, "\x1b[93mX\x1b[0m") {
		t.Fatalf("expected the yellow X with its colour codes, got %q", frame)
	}
}

func TestFrameBroadcaster_NonBlockingWhenFull(t *testing.T) {
	ch := make(chan string, 1)
	bs := render.NewBufferScreen(10, 2)
	fb := &frameBroadcaster{
		BufferScreen: bs,
		frameCh:      ch,
	}

	// Fill the channel
	ch <- "existing"

	// This should not block
	err := fb.Present()
	if err != nil {
		t.Fatalf("Present should not return error on full channel, got: %v", err)
	}
}

// ─── View rendering ─────────────────────────────────────────────────────────

func TestMenuView_ContainsGameContent(t *testing.T) {
	m := New()
	view := m.View()

	checks := []string{
		"CLI GAMES COLLECTION",
		"Snake",
		"Navigate, eat, survive",
		"navigate",
		"select",
		"quit",
		"╔", "╗", "╚", "╝",
	}

	for _, c := range checks {
		if !strings.Contains(view, c) {
			t.Errorf("menu view should contain %q", c)
		}
	}
}

func TestPlayingView_StartingPlaceholder(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.frame = ""

	view := m.View()
	if !strings.Contains(view, "Starting...") {
		t.Fatalf("expected 'Starting...' when no frame, got %q", view)
	}
}

func TestPlayingView_ShowsFrameAndOverlay(t *testing.T) {
	m := New()
	m.state = statePlaying
	m.frame = "game-content"

	view := m.View()
	if !strings.Contains(view, "game-content") {
		t.Fatalf("expected view to contain frame content, got %q", view)
	}
	if !strings.Contains(view, "ESC") {
		t.Fatalf("expected view to contain ESC hint, got %q", view)
	}
}

func TestEveryGamesControlsHintFitsTheTerminal(t *testing.T) {
	for _, g := range New().games {
		if line := "  [ESC] pause  [q] quit  " + g.Controls; len([]rune(line)) > 80 {
			t.Fatalf("%s controls hint is %d columns, wider than 80: %q", g.Name, len([]rune(line)), line)
		}
	}
}

func TestPlayingView_ShowsControlsOfTheRunningGame(t *testing.T) {
	cases := map[string][]string{
		"Snake":   {"[ESC] pause", "[q] quit", "[arrows/WASD] move"},
		"Sokoban": {"[ESC] pause", "[q] quit", "[arrows/WASD] move"},
		"Runner":  {"[ESC] pause", "[q] quit", "[↑/W/space] jump", "[↓/S] duck"},
		"Drone":   {"[ESC] pause", "[q] quit", "[↑/W/space] thrust"},
		"Breach":  {"[ESC] pause", "[q] quit", "[←→/AD] move", "[↑/W/space] launch"},
		"Botnet":  {"[ESC] pause", "[q] quit", "[←→/AD] move", "[↑/W/space] fire"},
		"Leak":    {"[ESC] pause", "[q] quit", "[arrows/WASD] move"},
		"Packet":  {"[ESC] pause", "[q] quit", "[↑↓/WS] switch link"},
		"Debris":  {"[ESC] pause", "[q] quit", "[←→/AD] turn", "[↑/W] thrust", "[↓/S] brake", "[space] fire"},
		"Rack":    {"[ESC] pause", "[q] quit", "[arrows/WASD] walk/climb", "[space] jump", "[f] throw"},
		"Defrag":  {"[ESC] pause", "[q] quit", "[←→↓/ASD] move", "[↑/W] rotate", "[space] slam", "[g] guides"},
	}
	for name, want := range cases {
		m := New()
		for i, g := range m.games {
			if g.Name == name {
				m.cursor = i
			}
		}
		m.state = statePlaying
		m.frame = "game-content"

		view := m.View()
		for _, hint := range want {
			if !strings.Contains(view, hint) {
				t.Fatalf("%s: view missing %q, got %q", name, hint, view)
			}
		}
		if name == "Runner" && strings.Contains(view, "move") {
			t.Fatalf("Runner hint must not say move, got %q", view)
		}
	}
}
