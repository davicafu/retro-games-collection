package shell

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/records"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/ui/highscore"
	"github.com/davidbarreiro/cli-games-collection/internal/games/botnet"
	"github.com/davidbarreiro/cli-games-collection/internal/games/breach"
	"github.com/davidbarreiro/cli-games-collection/internal/games/debris"
	"github.com/davidbarreiro/cli-games-collection/internal/games/defrag"
	"github.com/davidbarreiro/cli-games-collection/internal/games/drone"
	"github.com/davidbarreiro/cli-games-collection/internal/games/hack"
	"github.com/davidbarreiro/cli-games-collection/internal/games/leak"
	"github.com/davidbarreiro/cli-games-collection/internal/games/packet"
	"github.com/davidbarreiro/cli-games-collection/internal/games/rack"
	"github.com/davidbarreiro/cli-games-collection/internal/games/runner"
	"github.com/davidbarreiro/cli-games-collection/internal/games/snake"
	"github.com/davidbarreiro/cli-games-collection/internal/games/sokoban"
)

// FrameMsg is sent each time the game produces a new rendered frame.
type FrameMsg struct{ Frame string }

// GameOverMsg is sent when the game loop exits.
type GameOverMsg struct{ Err error }

type appState int

const (
	stateMenu appState = iota
	statePlaying
	// stateScores shows one game's high score table.
	stateScores
	// stateResume asks whether to continue a saved run or start a new one.
	stateResume
)

// gameEntry describes a playable game.
type gameEntry struct {
	Name string
	// Like names the classic the game is inspired by, shown next to its
	// name so the menu says what each game plays like; empty when the name
	// already says it.
	Like        string
	Description string
	Controls    string // game-specific keys shown under the frame while playing
	// Typing marks a game that takes typed text: q is a letter there, so
	// its hint offers ctrl+c to quit instead.
	Typing bool
	// Saved marks the game whose run is saved between plays.
	Saved   bool
	starter func(ctx context.Context, input input.Provider, screen render.Screen, l launch) error
}

// key names the game in the records files.
func (g gameEntry) key() string { return strings.ToLower(g.Name) }

// launch is what a game starts with: the terminal size, its high score
// table and, for a saved game, its slot and whether to continue it.
type launch struct {
	Width, Height int
	Board         records.Board
	Save          records.Slot
	Resume        bool
}

// Model is the root Bubble Tea model that orchestrates the entire app.
type Model struct {
	state appState

	// Menu
	cursor int
	games  []gameEntry

	// Playing
	gameCtrl *gameController
	frame    string
	chInput  *ChannelInput

	// Layout
	width  int
	height int

	// Records
	book   *records.Book
	saves  *records.Saves
	scores int // the game whose table is shown
	// newGame selects "New game" instead of "Continue" while asking.
	newGame bool
}

// Option configures the model.
type Option func(*Model)

// WithRecords keeps the high score tables in book and saved runs in saves;
// without it they live in memory only.
func WithRecords(book *records.Book, saves *records.Saves) Option {
	return func(m *Model) { m.book, m.saves = book, saves }
}

type gameController struct {
	cancel  context.CancelFunc
	frameCh chan string
	doneCh  chan error
}

// New creates the root model with available games.
func New(opts ...Option) Model {
	m := Model{
		state:  stateMenu,
		cursor: 0,
		games: []gameEntry{
			{Name: "Snake", Description: "Navigate, eat, survive", Controls: "[arrows/WASD] move", starter: runSnakeGame},
			{Name: "Sokoban", Description: "Push boxes to their targets", Controls: "[arrows/WASD] move", starter: runSokobanGame},
			{Name: "Runner", Like: "Chrome Dino", Description: "Jump, duck, keep running", Controls: "[↑/W/space] jump  [↓/S] duck", starter: runRunnerGame},
			{Name: "Drone", Like: "Flappy Bird", Description: "Fly between the antennas", Controls: "[↑/W/space] thrust", starter: runDroneGame},
			{Name: "Breach", Like: "Breakout", Description: "Infect your way through the firewall", Controls: "[←→/AD] move  [↑/W/space] launch", starter: runBreachGame},
			{Name: "Botnet", Like: "Space Invaders", Description: "Fire back at the antivirus waves", Controls: "[←→/AD] move  [↑/W/space] fire", starter: runBotnetGame},
			{Name: "Leak", Like: "Pac-Man", Description: "Leak memory, dodge the collectors", Controls: "[arrows/WASD] move", starter: runLeakGame},
			{Name: "Packet", Like: "Subway Surfers", Description: "Switch links, dodge the firewalls", Controls: "[↑↓/WS] switch link", starter: runPacketGame},
			{Name: "Debris", Like: "Asteroids", Description: "Clear the orbit of space debris", Controls: "[←→/AD] turn  [↑/W] thrust  [↓/S] brake  [space] fire", starter: runDebrisGame},
			{Name: "Rack", Like: "Donkey Kong", Description: "Climb the data center, grab the disks", Controls: "[arrows/WASD] walk/climb  [space] jump  [f] throw", starter: runRackGame},
			{Name: "Defrag", Like: "Tetris", Description: "Fit the fragments, free the sectors", Controls: "[←→↓/ASD] move  [↑/W] rotate  [space] slam  [g] guides", starter: runDefragGame},
			{Name: "Hack", Description: "Crack the network, read its secret", Controls: "[enter] run  [↑↓] history  [arrows/space] breach", Typing: true, Saved: true, starter: runHackGame},
		},
		width:  80,
		height: 24,
	}
	for _, o := range opts {
		o(&m)
	}
	if m.book == nil {
		m.book, _ = records.Open("")
	}
	if m.saves == nil {
		m.saves, _ = records.OpenSaves("")
	}
	return m
}

// NewWithGame creates a model that starts a specific game immediately,
// skipping the menu. Used by the play command.
func NewWithGame(name string, opts ...Option) Model {
	m := New(opts...)
	for i, g := range m.games {
		if strings.EqualFold(g.Name, name) {
			m.cursor = i
			m = m.choose(g)
			break
		}
	}
	return m
}

func (m Model) Init() tea.Cmd {
	if m.state == statePlaying {
		return m.awaitGameEvent()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateMenu:
			return m.updateMenu(msg)
		case statePlaying:
			return m.updatePlaying(msg)
		case stateScores:
			return m.updateScores(msg), nil
		case stateResume:
			return m.updateResume(msg)
		}

	case FrameMsg:
		if m.state == statePlaying {
			m.frame = msg.Frame
			return m, m.awaitGameEvent()
		}

	case GameOverMsg:
		m.state = stateMenu
		m.gameCtrl = nil
		m.frame = ""
		return m, nil
	}

	return m, nil
}

func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < len(m.games)-1 {
			m.cursor++
		}
	case tea.KeyEnter:
		m = m.choose(m.games[m.cursor])
		if m.state == statePlaying {
			return m, m.awaitGameEvent()
		}
		return m, nil
	case tea.KeyCtrlC, tea.KeyEscape:
		return m, tea.Quit
	}

	// Rune-based keys
	switch msg.String() {
	case "q", "Q":
		return m, tea.Quit
	case "r", "R":
		m.state, m.scores = stateScores, m.cursor
	}
	return m, nil
}

// choose starts the game, or first asks whether to continue its saved run.
func (m Model) choose(entry gameEntry) Model {
	if _, ok := m.savedLevel(entry); ok {
		m.state, m.newGame = stateResume, false
		return m
	}
	return m.startGame(entry)
}

// savedLevel is the level, from 1, a saved game's run continues at.
func (m Model) savedLevel(entry gameEntry) (int, bool) {
	if !entry.Saved {
		return 0, false
	}
	var cp hack.Checkpoint
	if ok, _ := m.saves.Slot(entry.key()).Load(&cp); !ok {
		return 0, false
	}
	return cp.Level + 1, true
}

// updateScores moves between games' tables with ←→; esc, q or r go back.
func (m Model) updateScores(msg tea.KeyMsg) Model {
	switch msg.Type {
	case tea.KeyLeft:
		m.scores = (m.scores + len(m.games) - 1) % len(m.games)
	case tea.KeyRight:
		m.scores = (m.scores + 1) % len(m.games)
	case tea.KeyEscape:
		m.state = stateMenu
	}
	switch msg.String() {
	case "q", "Q", "r", "R":
		m.state = stateMenu
	}
	return m
}

// updateResume picks Continue or New game with ↑↓ and enter; esc goes back.
func (m Model) updateResume(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		m.newGame = false
	case tea.KeyDown:
		m.newGame = true
	case tea.KeyEnter:
		m = m.launchGame(m.games[m.cursor], !m.newGame)
		return m, m.awaitGameEvent()
	case tea.KeyEscape:
		m.state = stateMenu
	}
	return m, nil
}

func (m Model) updatePlaying(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		m.chInput.Send(input.ActionQuit)
		if m.gameCtrl != nil {
			m.gameCtrl.cancel()
		}
		return m, nil
	case tea.KeyEnter:
		m.chInput.Send(input.ActionSelect)
		return m, nil
	case tea.KeyEscape:
		m.chInput.Send(input.ActionPause)
		return m, nil
	case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight:
		m.chInput.Send(mapKeyToAction(msg.Type))
		return m, nil
	}

	// Typed characters: each arrives as text and as the action it stands
	// for (input.RuneAction), so games keep WASD, Q, P and the rest.
	switch {
	case msg.Type == tea.KeyBackspace:
		m.chInput.Send(input.ActionBackspace)
	case (msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace) && !msg.Alt:
		runes := msg.Runes
		if msg.Type == tea.KeySpace && len(runes) == 0 {
			runes = []rune{' '}
		}
		for _, r := range runes {
			m.chInput.Type(r)
		}
	}

	return m, nil
}

// startGame starts a new game of entry.
func (m Model) startGame(entry gameEntry) Model { return m.launchGame(entry, false) }

// launchGame starts entry; resume continues its saved run.
func (m Model) launchGame(entry gameEntry, resume bool) Model {
	frameCh := make(chan string, 1)
	doneCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())

	m.state = statePlaying
	m.gameCtrl = &gameController{
		cancel:  cancel,
		frameCh: frameCh,
		doneCh:  doneCh,
	}
	m.chInput = NewChannelInput(4)

	screen := render.NewBufferScreen(m.width, m.height)
	provider := m.chInput

	go func() {
		defer close(frameCh)
		defer close(doneCh)

		broadcaster := &frameBroadcaster{
			BufferScreen: screen,
			frameCh:      frameCh,
		}

		l := launch{Width: m.width, Height: m.height, Board: m.book.Board(entry.key()), Resume: resume}
		if entry.Saved {
			l.Save = m.saves.Slot(entry.key())
		}
		err := entry.starter(ctx, provider, broadcaster, l)
		if err != nil {
			doneCh <- err
			return
		}
		doneCh <- nil
	}()

	return m
}

func (m Model) awaitGameEvent() tea.Cmd {
	return func() tea.Msg {
		if m.gameCtrl == nil {
			return nil
		}
		select {
		case err := <-m.gameCtrl.doneCh:
			return GameOverMsg{Err: err}
		default:
		}
		select {
		case frame, ok := <-m.gameCtrl.frameCh:
			if !ok {
				select {
				case err := <-m.gameCtrl.doneCh:
					return GameOverMsg{Err: err}
				default:
					return GameOverMsg{Err: nil}
				}
			}
			return FrameMsg{Frame: frame}
		case err := <-m.gameCtrl.doneCh:
			return GameOverMsg{Err: err}
		}
	}
}

func (m Model) View() string {
	switch m.state {
	case stateMenu:
		return m.menuView()
	case statePlaying:
		return m.playingView()
	case stateScores:
		return m.scoresView()
	case stateResume:
		return m.resumeView()
	}
	return ""
}

func (m Model) menuView() string {
	const width = 44
	const inner = width - 2 // 42 — espacio visible entre ║║
	const innerBox = 36     // 42 - 6 — espacio entre ││ del inner box, 2 de margen a cada lado
	const textArea = 34     // innerBox - 2 — texto entre los espacios │ ▸ text │

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("220")).
		Render("🎮  CLI GAMES COLLECTION")

	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", inner))

	hintsStyle := lipgloss.NewStyle().
		Width(inner).
		Foreground(lipgloss.Color("240"))
	hints := hintsStyle.Render("  ↑↓ navigate  ↵ select  r scores  q quit")

	var b strings.Builder
	b.Grow(width * 18)

	// Helper: escribe ║ + line(42) + ║\n  (line ya debe tener 42 cols visibles)
	boxLine := func(line42 string) {
		b.WriteString("║" + line42 + "║\n")
	}

	// ─── Outer box ──────────────────────────────────────────────
	b.WriteString("╔" + strings.Repeat("═", inner) + "╗\n")
	boxLine(lipgloss.NewStyle().Width(inner).Align(lipgloss.Center).Render(title))
	boxLine(sep)
	boxLine(strings.Repeat(" ", inner))

	// ─── Inner game list box ────────────────────────────────────
	boxLine("  ┌" + strings.Repeat("─", innerBox) + "┐  ")

	for i, g := range m.games {
		cur := "  "
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
		if i == m.cursor {
			cur = "▸ "
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
		}

		// "  │ ▸ Snake                        │  " (cols 0-1 margen, 2│, 3 , 4-37 texto, 38 , 39│, 40-41 margen)
		label := nameStyle.Render(g.Name)
		if g.Like != "" {
			label += lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(" · " + g.Like + "-like")
		}
		if _, ok := m.savedLevel(g); ok {
			label += lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(" · saved run")
		}
		nameText := lipgloss.NewStyle().Width(textArea).Render(cur + label)
		boxLine("  │ " + nameText + " │  ")
	}

	boxLine("  └" + strings.Repeat("─", innerBox) + "┘  ")

	// Only the selected game is described, below the list, so the menu fits
	// a 24-row terminal with every game of the roadmap.
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	desc := lipgloss.NewStyle().Width(inner).Render("    " + descStyle.Render(m.games[m.cursor].Description))
	boxLine(desc)

	// ─── Footer ─────────────────────────────────────────────────
	boxLine(strings.Repeat(" ", inner))
	boxLine(hints)
	b.WriteString("╚" + strings.Repeat("═", inner) + "╝")

	return b.String()
}

// panel draws a box like the menu's: a title, a separator, the lines and
// a hint row, every row inner columns wide.
func panel(title string, lines []string, hint string) string {
	const inner = 42
	row := func(text string) string {
		return "║" + lipgloss.NewStyle().Width(inner).Render(text) + "║\n"
	}
	var b strings.Builder
	b.WriteString("╔" + strings.Repeat("═", inner) + "╗\n")
	b.WriteString("║" + lipgloss.NewStyle().Width(inner).Align(lipgloss.Center).Render(
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")).Render(title)) + "║\n")
	b.WriteString("║" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("─", inner)) + "║\n")
	b.WriteString(row(""))
	for _, l := range lines {
		b.WriteString(row("  " + l))
	}
	b.WriteString(row(""))
	b.WriteString(row(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  " + hint)))
	b.WriteString("╚" + strings.Repeat("═", inner) + "╝")
	return b.String()
}

// scoresView is one game's high score table.
func (m Model) scoresView() string {
	g := m.games[m.scores]
	lines := highscore.Lines(m.book.Top(g.key()), 0, nil)[1:]
	return panel("HIGH SCORES · "+g.Name, lines, "←→ game  esc back")
}

// resumeView asks whether to continue the saved run or start a new one.
func (m Model) resumeView() string {
	g := m.games[m.cursor]
	level, _ := m.savedLevel(g)
	options := []string{fmt.Sprintf("Continue (level %d)", level), "New game"}
	selected := 0
	if m.newGame {
		selected = 1
	}
	for i := range options {
		cur := "  "
		if i == selected {
			cur = "▸ "
			options[i] = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46")).Render(options[i])
		}
		options[i] = cur + options[i]
	}
	return panel(g.Name+" · saved run", options, "↑↓ choose  ↵ start  esc back")
}

func (m Model) playingView() string {
	if m.frame == "" {
		return "Starting..."
	}

	overlay := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("  [ESC] pause  " + quitKey(m.games[m.cursor]) + "  " + m.games[m.cursor].Controls)

	return m.frame + "\n" + overlay
}

// frameBroadcaster wraps BufferScreen and sends snapshots, with their
// colours, after each Present.
type frameBroadcaster struct {
	*render.BufferScreen
	frameCh chan<- string
}

func (f *frameBroadcaster) Present() error {
	select {
	case f.frameCh <- f.StyledSnapshot():
	default:
	}
	return nil
}

func mapKeyToAction(kt tea.KeyType) input.Action {
	switch kt {
	case tea.KeyUp:
		return input.ActionUp
	case tea.KeyDown:
		return input.ActionDown
	case tea.KeyLeft:
		return input.ActionLeft
	case tea.KeyRight:
		return input.ActionRight
	default:
		return ""
	}
}

func runSnakeGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	ticker := clock.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	selector := snake.NewRandomSelector(rand.New(rand.NewSource(time.Now().UnixNano())))
	module := snake.NewModule(ticker, screen, provider, selector).WithBoard(l.Board)
	return module.Run(ctx)
}

func runSokobanGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	module := sokoban.NewModule(screen, provider).WithBoard(l.Board)
	return module.Run(ctx)
}

func runRunnerGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	ticker := clock.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	module := runner.NewModule(ticker, screen, provider, rng).WithBoard(l.Board)
	return module.Run(ctx)
}

func runDroneGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	ticker := clock.NewTicker(drone.TickPeriod)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	module := drone.NewModule(ticker, screen, provider, rng).WithBoard(l.Board)
	return module.Run(ctx)
}

func runBreachGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	ticker := clock.NewTicker(breach.TickPeriod)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	module := breach.NewModule(ticker, screen, provider, rng).WithBoard(l.Board)
	return module.Run(ctx)
}

func runBotnetGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	ticker := clock.NewTicker(botnet.TickPeriod)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	module := botnet.NewModule(ticker, screen, provider, rng).WithBoard(l.Board)
	return module.Run(ctx)
}

func runLeakGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	ticker := clock.NewTicker(leak.TickPeriod)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	module := leak.NewModule(ticker, screen, provider, rng).WithBoard(l.Board)
	return module.Run(ctx)
}

func runPacketGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	ticker := clock.NewTicker(packet.TickPeriod)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	module := packet.NewModule(ticker, screen, provider, rng).WithBoard(l.Board)
	return module.Run(ctx)
}

func runDebrisGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	ticker := clock.NewTicker(debris.TickPeriod)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	module := debris.NewModule(ticker, screen, provider, rng).WithBoard(l.Board)
	return module.Run(ctx)
}

func runRackGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	ticker := clock.NewTicker(rack.TickPeriod)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	module := rack.NewModule(ticker, screen, provider, rng).WithBoard(l.Board)
	return module.Run(ctx)
}

func runDefragGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	ticker := clock.NewTicker(defrag.TickPeriod)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	module := defrag.NewModule(ticker, screen, provider, rng).WithBoard(l.Board)
	return module.Run(ctx)
}

// quitKey is the hint for leaving a game: q, or ctrl+c where q is text.
func quitKey(g gameEntry) string {
	if g.Typing {
		return "[ctrl+c] quit"
	}
	return "[q] quit"
}

func runHackGame(ctx context.Context, provider input.Provider, screen render.Screen, l launch) error {
	ticker := clock.NewTicker(hack.TickPeriod)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	module := hack.NewModule(ticker, screen, provider, rng).WithBoard(l.Board).WithSave(l.Save, l.Resume)
	return module.Run(ctx)
}
