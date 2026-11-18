// Package hack is "Hack the System", the hybrid game: a simulated terminal
// to explore a fictional network and read its story, a map of its
// machines, and short arcade breaches that guard them, under a trace meter.
package hack

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// TickPeriod is the real time one Step stands for; every per-tick constant
// is tuned to it.
const TickPeriod = time.Second / 60

// The trace rises by these points out of 100: a failed or unknown command,
// a wrong password or key, a lost breach.
const (
	TraceFailed   = 4
	TracePassword = 10
	TraceHoneypot = 30
	// While on a watched machine the trace rises by TraceWatched every
	// Difficulty.WatchTicks, until the level's IDS machine is yours.
	TraceWatched = 1
	maxLines     = 200
)

// The final score starts at ScoreStart and loses HintCost for every hint
// taken and TraceCost for every point of trace, never going below 0.
const (
	ScoreStart = 1000
	HintCost   = 100
	TraceCost  = 5
)

// Mode is what the screen shows and the keys drive.
type Mode int

const (
	ModeTerminal Mode = iota
	ModeMap
	ModeBreach
)

// LineKind is how a terminal line reads.
type LineKind int

const (
	LineOutput LineKind = iota
	LineCommand
	LineError
	LineSuccess
)

// Line is a terminal line.
type Line struct {
	Text string
	Kind LineKind
}

// Rand is the randomness the breaches need; *math/rand.Rand satisfies it.
type Rand interface {
	Intn(n int) int
}

// Input is the player's intent for one tick: typed text and editing keys
// for the terminal (History -1 is up, 1 down), and moves and Act (space)
// for the map and the breaches.
type Input struct {
	Text      []rune
	Enter     bool
	Backspace bool
	History   int
	Move      grid.Point
	Act       bool
}

type Config struct {
	Rand Rand
	// Resume, when set and possible, continues a saved run instead of
	// starting at level 1.
	Resume *Checkpoint
}

// Checkpoint is a run saved between levels: the level to play next, the
// score of the finished ones and the time played.
type Checkpoint struct {
	Level     int `json:"level"`
	Completed int `json:"completed"`
	Ticks     int `json:"ticks"`
}

// valid reports whether the checkpoint could come from a run: a level after
// the first that exists, and no negative score or time.
func (cp Checkpoint) valid() bool {
	return cp.Level >= 1 && cp.Level < len(levels) && cp.Completed >= 0 && cp.Ticks >= 0
}

// Game is the headless Hack simulation, advanced one tick at a time.
type Game struct {
	newBreach func(BreachKind) Breach

	level      int
	completed  int             // the score of the finished levels
	cleared    bool            // the level's secret is read; enter opens the next
	attempts   map[string]int  // wrong passwords per machine
	locked     map[string]bool // machines locked by too many wrong passwords
	watchClock int
	nodes      map[string]*Node
	owned      map[string]bool
	decrypted  map[string]bool // "node/file"
	node       string

	mode     Mode
	lines    []Line
	buffer   []rune
	history  []string
	recall   int
	password string // the node asking for a password, if any
	breach   Breach
	target   string // the node being breached
	selected string // the node selected on the map

	trace    int
	hints    int
	revealed []int // hints shown per stage of the level
	ticks    int
	won      bool
	over     bool
}

func NewGame(cfg Config) *Game {
	g := &Game{}
	g.newBreach = func(k BreachKind) Breach {
		d := difficulties[g.level]
		switch k {
		case BreachICE:
			return newICE(cfg.Rand, d)
		case BreachSync:
			return newSync(cfg.Rand, d)
		}
		return newGates(cfg.Rand, d)
	}
	g.loadLevel(0)
	if cp := cfg.Resume; cp != nil && cp.valid() {
		g.loadLevel(cp.Level)
		g.completed, g.ticks = cp.Completed, cp.Ticks
	}
	return g
}

// Checkpoint is the run to save once a level is cleared: the next level,
// with the score and time so far. A level under way has none.
func (g *Game) Checkpoint() (Checkpoint, bool) {
	if !g.cleared {
		return Checkpoint{}, false
	}
	return Checkpoint{Level: g.level + 1, Completed: g.completed, Ticks: g.ticks}, true
}

// loadLevel starts level n afresh at its gateway: a clear screen, no
// machine but the gateway, and no trace or hints.
func (g *Game) loadLevel(n int) {
	l := levels[n]
	g.level, g.cleared = n, false
	g.nodes, g.owned, g.decrypted = map[string]*Node{}, map[string]bool{"gate": true}, map[string]bool{}
	g.attempts, g.locked, g.watchClock = map[string]int{}, map[string]bool{}, 0
	for i := range l.Nodes {
		node := l.Nodes[i]
		g.nodes[node.ID] = &node
	}
	g.node, g.mode, g.password, g.breach = "gate", ModeTerminal, "", nil
	g.trace, g.hints, g.revealed = 0, 0, make([]int, len(l.Hints))
	g.lines = nil
	g.say(LineSuccess, fmt.Sprintf("NIVEL %d // %s", n+1, l.Name))
	g.say(LineOutput, "Escribe `help` para ver los comandos y `cat readme.txt` para el encargo.")
}

// watch raises the trace while the player stands on a watched machine and
// the level's IDS is not theirs.
func (g *Game) watch() {
	if g.cleared || !g.Watched(g.node) {
		g.watchClock = 0
		return
	}
	if g.watchClock++; g.watchClock >= difficulties[g.level].WatchTicks {
		g.watchClock = 0
		g.raise(TraceWatched)
	}
}

// Watched reports whether standing on id raises the trace now.
func (g *Game) Watched(id string) bool {
	n, ok := g.nodes[id]
	ids := levels[g.level].IDS
	return ok && n.Watched && (ids == "" || !g.owned[ids])
}

// finish closes the level whose secret was just read: the last one wins,
// any other adds its score and waits for enter to open the next.
func (g *Game) finish() {
	if g.level == len(levels)-1 {
		g.won, g.over = true, true
		return
	}
	points := g.levelScore()
	g.completed += points
	g.cleared = true
	g.say(LineSuccess, fmt.Sprintf("--- NIVEL %d SUPERADO (+%d puntos) ---", g.level+1, points))
	g.say(LineSuccess, "Pulsa enter para conectar con el siguiente objetivo.")
}

// Step advances the simulation by one tick. A finished round does not
// advance.
func (g *Game) Step(in Input) {
	if g.over {
		return
	}
	g.ticks++
	g.watch()
	if g.over {
		return
	}
	switch g.mode {
	case ModeTerminal:
		g.edit(in)
	case ModeMap:
		g.steerMap(in)
	case ModeBreach:
		g.breach.Step(in)
		switch g.breach.Outcome() {
		case Won:
			g.mode = ModeTerminal
			g.say(LineSuccess, "brecha superada: "+g.target+" es tuyo")
			g.enter(g.target)
		case Lost:
			g.mode = ModeTerminal
			g.say(LineError, "brecha fallida: conexión perdida")
			g.raise(difficulties[g.level].BreachTrace)
		}
	}
}

func (g *Game) edit(in Input) {
	g.buffer = append(g.buffer, in.Text...)
	if in.Backspace && len(g.buffer) > 0 {
		g.buffer = g.buffer[:len(g.buffer)-1]
	}
	if in.History != 0 && len(g.history) > 0 {
		g.recall = min(max(g.recall+in.History, 0), len(g.history))
		g.buffer = nil
		if g.recall < len(g.history) {
			g.buffer = []rune(g.history[g.recall])
		}
	}
	if !in.Enter {
		return
	}
	if g.cleared {
		g.buffer = nil
		g.loadLevel(g.level + 1)
		return
	}
	line := strings.TrimSpace(string(g.buffer))
	prompt := g.Prompt()
	g.buffer = nil
	if g.password != "" {
		g.say(LineCommand, prompt+strings.Repeat("*", len([]rune(line))))
		g.checkPassword(line)
		return
	}
	g.say(LineCommand, prompt+line)
	if line == "" {
		return
	}
	g.history = append(g.history, line)
	g.recall = len(g.history)
	g.run(strings.Fields(line))
}

func (g *Game) run(args []string) {
	switch cmd := args[0]; {
	case cmd == "help":
		for _, l := range []string{
			"help                       esta lista",
			"ls                         ficheros de esta máquina",
			"cat <fichero>              muestra un fichero",
			"scan                       máquinas a las que llega esta",
			"connect <máquina>          ve a una máquina; si está protegida, brecha",
			"decrypt <fichero> <clave>  descifra un fichero",
			"map                        mapa de la red (flechas, enter, backspace)",
			"hint                       una pista (resta puntos)",
			"clear                      limpia la pantalla",
		} {
			g.say(LineOutput, l)
		}
	case cmd == "ls":
		for _, f := range g.nodes[g.node].Files {
			name := f.Name
			if f.Encrypted && !g.decrypted[g.node+"/"+f.Name] {
				name += "   (cifrado)"
			}
			g.say(LineOutput, name)
		}
	case cmd == "cat" && len(args) == 2:
		g.cat(args[1])
	case cmd == "scan":
		for _, id := range g.nodes[g.node].Links {
			line := fmt.Sprintf("%-6s %-24s %s", id, g.nodes[id].Title, g.status(id))
			if g.Watched(id) {
				line += "  vigilada"
			}
			g.say(LineOutput, line)
		}
	case cmd == "connect" && len(args) == 2:
		g.connect(args[1])
	case cmd == "decrypt" && len(args) == 3:
		g.decrypt(args[1], args[2])
	case cmd == "hint":
		g.hint()
	case cmd == "map":
		g.mode, g.selected = ModeMap, g.node
	case cmd == "clear":
		g.lines = nil
	case slices.Contains([]string{"cat", "connect", "decrypt"}, cmd):
		g.fail("uso: mira `help`")
	default:
		g.fail("comando no encontrado: " + cmd)
	}
}

// status is how a node stands, as scan shows it.
func (g *Game) status(id string) string {
	n := g.nodes[id]
	switch {
	case g.owned[id]:
		return "[tuyo]"
	case g.locked[id]:
		return "[forzar " + breachNames[n.LockBreach] + "]"
	case n.Requires != "" && !g.owned[n.Requires]:
		return "[bloqueado: " + n.Requires + "]"
	case n.Password != "":
		return "[contraseña]"
	case n.NeedsKey != "":
		return "[clave + " + breachNames[n.Breach] + "]"
	}
	return "[" + breachNames[n.Breach] + "]"
}

var breachNames = map[BreachKind]string{NoBreach: "abierto", BreachICE: "ICE", BreachSync: "SYNC", BreachGates: "GATES"}

func (g *Game) file(name string) (File, bool) {
	for _, f := range g.nodes[g.node].Files {
		if f.Name == name {
			return f, true
		}
	}
	return File{}, false
}

func (g *Game) cat(name string) {
	f, ok := g.file(name)
	if !ok {
		g.fail("cat: " + name + ": no existe")
		return
	}
	if f.Encrypted && !g.decrypted[g.node+"/"+name] {
		g.say(LineOutput, "▓▒░▓▓░▒▒▓░░▒▓▒░▓▒▓░░▒▓▓▒░▒")
		g.say(LineOutput, "(cifrado: decrypt <fichero> <clave>)")
		return
	}
	for _, l := range f.Body {
		g.say(LineOutput, l)
	}
	if f.Final {
		g.finish()
	}
}

// reachable reports whether an owned machine links to id.
func (g *Game) reachable(id string) bool {
	for o := range g.owned {
		if slices.Contains(g.nodes[o].Links, id) {
			return true
		}
	}
	return false
}

func (g *Game) connect(id string) {
	n, ok := g.nodes[id]
	switch {
	case !ok:
		g.fail("connect: " + id + ": host desconocido")
	case id == g.node:
		g.say(LineOutput, "ya estás en "+id)
	case g.owned[id]:
		g.enter(id)
	case !g.reachable(id):
		g.fail("connect: " + id + ": sin ruta desde tus máquinas")
	case g.locked[id]:
		g.mode, g.target = ModeBreach, id
		g.breach = g.newBreach(n.LockBreach)
	case n.Requires != "" && !g.owned[n.Requires]:
		g.fail("connect: " + id + ": bloqueado por el firewall de " + n.Requires)
	case n.Honeypot:
		g.say(LineError, "connect: "+id+": ¡era un honeypot! te han marcado")
		g.raise(TraceHoneypot)
	case n.Password != "":
		g.password = id
	case n.NeedsKey != "" && !g.decrypted[g.keyHolder(n.NeedsKey)+"/"+n.NeedsKey]:
		g.fail("acceso denegado: " + id + " pide su clave de admin")
	case n.Breach != NoBreach:
		g.mode, g.target = ModeBreach, id
		g.breach = g.newBreach(n.Breach)
	default:
		g.enter(id)
	}
}

// keyHolder is the node that holds a file.
func (g *Game) keyHolder(name string) string {
	for _, n := range levels[g.level].Nodes {
		for _, f := range n.Files {
			if f.Name == name {
				return n.ID
			}
		}
	}
	return ""
}

func (g *Game) checkPassword(line string) {
	id := g.password
	g.password = ""
	if n := g.nodes[id]; line != n.Password {
		g.say(LineError, "acceso denegado")
		if g.attempts[id]++; n.Lockout > 0 && g.attempts[id] >= n.Lockout {
			g.locked[id] = true
			g.say(LineError, "demasiados intentos: "+id+" bloqueada. Solo queda forzarla.")
		}
		g.raise(TracePassword)
		return
	}
	g.enter(id)
}

// enter owns id and moves there.
func (g *Game) enter(id string) {
	g.owned[id] = true
	g.node = id
	g.say(LineSuccess, "conectado a "+id+" ("+g.nodes[id].Title+")")
}

func (g *Game) decrypt(name, key string) {
	f, ok := g.file(name)
	switch {
	case !ok:
		g.fail("decrypt: " + name + ": no existe")
	case !f.Encrypted:
		g.fail("decrypt: " + name + ": no está cifrado")
	case key != f.Key:
		g.say(LineError, "decrypt: clave incorrecta")
		g.raise(TracePassword)
	default:
		g.decrypted[g.node+"/"+name] = true
		for _, l := range f.Body {
			g.say(LineSuccess, l)
		}
		if f.Final {
			g.finish()
		}
	}
}

// steerMap moves the selection to the nearest machine in the pressed
// direction; enter connects to it and backspace returns to the terminal.
func (g *Game) steerMap(in Input) {
	switch {
	case in.Backspace:
		g.mode = ModeTerminal
	case in.Enter:
		g.mode = ModeTerminal
		g.say(LineCommand, g.Prompt()+"connect "+g.selected)
		g.connect(g.selected)
	case in.Move != (grid.Point{}):
		from := g.nodes[g.selected].Pos
		best, bestScore := g.selected, -1
		for _, n := range levels[g.level].Nodes {
			d := grid.Point{X: n.Pos.X - from.X, Y: n.Pos.Y - from.Y}
			along := d.X*in.Move.X + d.Y*in.Move.Y
			if along <= 0 {
				continue
			}
			across := d.X*in.Move.Y + d.Y*in.Move.X
			if across < 0 {
				across = -across
			}
			if score := along + 2*across; bestScore < 0 || score < bestScore {
				best, bestScore = n.ID, score
			}
		}
		g.selected = best
	}
}

func (g *Game) say(kind LineKind, text string) {
	g.lines = append(g.lines, Line{Text: text, Kind: kind})
	if len(g.lines) > maxLines {
		g.lines = g.lines[len(g.lines)-maxLines:]
	}
}

// stage is how far the run has got on this level: the last of its hints
// whose condition holds.
func (g *Game) stage() int {
	st := 0
	for i, h := range levels[g.level].Hints {
		if h.When == nil || h.When(g) {
			st = i
		}
	}
	return st
}

// hint shows the stage's next hint and charges for it; once both are
// shown it repeats the explicit one for free.
func (g *Game) hint() {
	st := g.stage()
	h := levels[g.level].Hints[st]
	text := []string{h.Vague, h.Explicit}
	if g.revealed[st] < len(text) {
		g.revealed[st]++
		g.hints++
		g.say(LineSuccess, fmt.Sprintf("pista (-%d puntos):", HintCost))
	}
	g.say(LineSuccess, text[g.revealed[st]-1])
}

// fail reports a failed command and raises the trace for it.
func (g *Game) fail(text string) {
	g.say(LineError, text)
	g.raise(TraceFailed)
}

func (g *Game) raise(points int) {
	g.trace = min(g.trace+points, 100)
	if g.trace == 100 {
		g.say(LineError, "RASTREO COMPLETO. Saben dónde estás.")
		g.over = true
	}
}

// Prompt is the terminal prompt: the shell's, or a password request.
func (g *Game) Prompt() string {
	if g.password != "" {
		return "contraseña de " + g.password + ": "
	}
	return "ghost@" + g.node + ":~$ "
}

// Buffer is the line being typed, masked while asking for a password.
func (g *Game) Buffer() string {
	if g.password != "" {
		return strings.Repeat("*", len(g.buffer))
	}
	return string(g.buffer)
}

// Text is the whole terminal, one line per line.
func (g *Game) Text() string {
	var b strings.Builder
	for _, l := range g.lines {
		b.WriteString(l.Text)
		b.WriteByte('\n')
	}
	return b.String()
}

// TakingText reports whether typed keys are text: at the terminal.
func (g *Game) TakingText() bool { return g.mode == ModeTerminal && !g.over }

func (g *Game) Lines() []Line           { return slices.Clone(g.lines) }
func (g *Game) Mode() Mode              { return g.mode }
func (g *Game) Node() string            { return g.node }
func (g *Game) Owned(id string) bool    { return g.owned[id] }
func (g *Game) Status(id string) string { return g.status(id) }
func (g *Game) Selected() string        { return g.selected }
func (g *Game) Breach() Breach          { return g.breach }
func (g *Game) Trace() int              { return g.trace }
func (g *Game) Hints() int              { return g.hints }

// Score is the finished levels' score plus, while the level is still
// being played, what this one would score now.
func (g *Game) Score() int {
	if g.cleared || g.over && !g.won {
		return g.completed
	}
	return g.completed + g.levelScore()
}

// levelScore is this level's score: ScoreStart less the hints and the
// trace, never below 0.
func (g *Game) levelScore() int { return max(ScoreStart-HintCost*g.hints-TraceCost*g.trace, 0) }

// Level is the level being played, from 0.
func (g *Game) Level() int { return g.level }

// Nodes are the level's machines.
func (g *Game) Nodes() []Node { return slices.Clone(levels[g.level].Nodes) }
func (g *Game) Won() bool     { return g.won }
func (g *Game) Over() bool    { return g.over }

// Elapsed is the time played, exact: TickPeriod itself is rounded down to
// a nanosecond, which adds up over long runs.
func (g *Game) Elapsed() time.Duration {
	return time.Duration(g.ticks) * time.Second / time.Duration(time.Second/TickPeriod)
}
func (g *Game) Ticks() int { return g.ticks }
