package shell

import (
	"context"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/clock"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
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
)

// gameEntry describes a playable game.
type gameEntry struct {
	Name        string
	Description string
	starter     func(ctx context.Context, input input.Provider, screen render.Screen, width, height int) error
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
}

type gameController struct {
	cancel  context.CancelFunc
	frameCh chan string
	doneCh  chan error
}

// New creates the root model with available games.
func New() Model {
	return Model{
		state:  stateMenu,
		cursor: 0,
		games: []gameEntry{
			{Name: "Snake", Description: "Navigate, eat, survive", starter: runSnakeGame},
			{Name: "Sokoban", Description: "Push boxes to their targets", starter: runSokobanGame},
			{Name: "Runner", Description: "Jump, duck, keep running", starter: runRunnerGame},
		},
		width:  80,
		height: 24,
	}
}

// NewWithGame creates a model that starts a specific game immediately,
// skipping the menu. Used by the play command.
func NewWithGame(name string) Model {
	m := New()
	for i, g := range m.games {
		if strings.EqualFold(g.Name, name) {
			m = m.startGame(g)
			m.cursor = i
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
		entry := m.games[m.cursor]
		m = m.startGame(entry)
		return m, m.awaitGameEvent()
	case tea.KeyCtrlC, tea.KeyEscape:
		return m, tea.Quit
	}

	// Rune-based keys
	switch msg.String() {
	case "q", "Q":
		return m, tea.Quit
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

	// Rune-based keys
	switch msg.String() {
	case "w", "W":
		m.chInput.Send(input.ActionUp)
	case "a", "A":
		m.chInput.Send(input.ActionLeft)
	case "s", "S":
		m.chInput.Send(input.ActionDown)
	case "d", "D":
		m.chInput.Send(input.ActionRight)
	case "q", "Q":
		m.chInput.Send(input.ActionQuit)
	case " ":
		m.chInput.Send(input.ActionSelect)
	case "p", "P":
		m.chInput.Send(input.ActionPause)
	case "u", "U":
		m.chInput.Send(input.ActionUndo)
	case "?", "h", "H":
		m.chInput.Send(input.ActionHelp)
	}

	return m, nil
}

func (m Model) startGame(entry gameEntry) Model {
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

		err := entry.starter(ctx, provider, broadcaster, m.width, m.height)
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
	}
	return ""
}

func (m Model) menuView() string {
	const width = 44
	const inner = width - 2 // 42 — espacio visible entre ║║
	const innerBox = 38     // 42 - 4 — espacio entre ││ del inner box
	const textArea = 36     // innerBox - 2 — texto entre los espacios │ ▸ text │

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
	hints := hintsStyle.Render("  ↑↓ navigate  ↵ select  q quit")

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
	boxLine("  ┌" + strings.Repeat("─", innerBox) + "┐")

	for i, g := range m.games {
		cur := "  "
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
		if i == m.cursor {
			cur = "▸ "
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
		}

		// "  │ ▸ Snake                          │" (cols 0-1 margen, 2│, 3 , 4-39 texto, 40 , 41│)
		nameText := lipgloss.NewStyle().Width(textArea).Render(cur + nameStyle.Render(g.Name))
		boxLine("  │ " + nameText + " │")

		// "  │      Navigate, eat, survive       │"
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
		descText := lipgloss.NewStyle().Width(textArea).Render("    " + descStyle.Render(g.Description))
		boxLine("  │ " + descText + " │")

		if i < len(m.games)-1 {
			boxLine("  │" + strings.Repeat(" ", innerBox) + "│")
		}
	}

	boxLine("  └" + strings.Repeat("─", innerBox) + "┘")

	// ─── Footer ─────────────────────────────────────────────────
	boxLine(strings.Repeat(" ", inner))
	boxLine(hints)
	b.WriteString("╚" + strings.Repeat("═", inner) + "╝")

	return b.String()
}

func (m Model) playingView() string {
	if m.frame == "" {
		return "Starting..."
	}

	overlay := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("  [ESC] pause  [q] quit  [arrows/WASD] move")

	return m.frame + "\n" + overlay
}

// frameBroadcaster wraps BufferScreen and sends snapshots after each Present.
type frameBroadcaster struct {
	*render.BufferScreen
	frameCh chan<- string
}

func (f *frameBroadcaster) Present() error {
	select {
	case f.frameCh <- f.Snapshot():
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

func runSnakeGame(ctx context.Context, provider input.Provider, screen render.Screen, width, height int) error {
	ticker := clock.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	selector := snake.NewRandomSelector(rand.New(rand.NewSource(time.Now().UnixNano())))
	module := snake.NewModule(ticker, screen, provider, selector)
	return module.Run(ctx)
}

func runSokobanGame(ctx context.Context, provider input.Provider, screen render.Screen, width, height int) error {
	module := sokoban.NewModule(screen, provider)
	return module.Run(ctx)
}

func runRunnerGame(ctx context.Context, provider input.Provider, screen render.Screen, width, height int) error {
	ticker := clock.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	module := runner.NewModule(ticker, screen, provider, rng)
	return module.Run(ctx)
}
