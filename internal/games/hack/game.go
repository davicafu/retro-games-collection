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
	TraceBreach   = 20
	maxLines      = 200
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
}

// Game is the headless Hack simulation, advanced one tick at a time.
type Game struct {
	newBreach func(BreachKind) Breach

	nodes     map[string]*Node
	owned     map[string]bool
	decrypted map[string]bool // "node/file"
	node      string

	mode     Mode
	lines    []Line
	buffer   []rune
	history  []string
	recall   int
	password string // the node asking for a password, if any
	breach   Breach
	target   string // the node being breached
	selected string // the node selected on the map

	trace int
	ticks int
	won   bool
	over  bool
}

func NewGame(cfg Config) *Game {
	g := &Game{
		nodes: map[string]*Node{}, owned: map[string]bool{"gate": true},
		decrypted: map[string]bool{}, node: "gate",
	}
	g.newBreach = func(k BreachKind) Breach {
		switch k {
		case BreachICE:
			return newICE(cfg.Rand)
		case BreachSync:
			return newSync(cfg.Rand)
		}
		return newGates(cfg.Rand)
	}
	for i := range network {
		n := network[i]
		g.nodes[n.ID] = &n
	}
	g.say(LineSuccess, "MERIDIAN GRID // contractor gateway")
	g.say(LineOutput, "Type `help` for the commands, `cat readme.txt` for the job.")
	return g
}

// Step advances the simulation by one tick. A finished round does not
// advance.
func (g *Game) Step(in Input) {
	if g.over {
		return
	}
	g.ticks++
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
			g.say(LineSuccess, "breach complete: "+g.target+" is yours")
			g.enter(g.target)
		case Lost:
			g.mode = ModeTerminal
			g.say(LineError, "breach failed: connection dropped")
			g.raise(TraceBreach)
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
			"help                    this list",
			"ls                      files on this machine",
			"cat <file>              print a file",
			"scan                    machines this one reaches",
			"connect <machine>       move to a machine, breaching it if guarded",
			"decrypt <file> <key>    decrypt a file",
			"map                     the network map (arrows, enter, backspace)",
			"clear                   clear the screen",
		} {
			g.say(LineOutput, l)
		}
	case cmd == "ls":
		for _, f := range g.nodes[g.node].Files {
			name := f.Name
			if f.Encrypted && !g.decrypted[g.node+"/"+f.Name] {
				name += "   (encrypted)"
			}
			g.say(LineOutput, name)
		}
	case cmd == "cat" && len(args) == 2:
		g.cat(args[1])
	case cmd == "scan":
		for _, id := range g.nodes[g.node].Links {
			g.say(LineOutput, fmt.Sprintf("%-6s %-20s %s", id, g.nodes[id].Title, g.status(id)))
		}
	case cmd == "connect" && len(args) == 2:
		g.connect(args[1])
	case cmd == "decrypt" && len(args) == 3:
		g.decrypt(args[1], args[2])
	case cmd == "map":
		g.mode, g.selected = ModeMap, g.node
	case cmd == "clear":
		g.lines = nil
	case slices.Contains([]string{"cat", "connect", "decrypt"}, cmd):
		g.fail("usage: see `help`")
	default:
		g.fail("command not found: " + cmd)
	}
}

// status is how a node stands, as scan shows it.
func (g *Game) status(id string) string {
	n := g.nodes[id]
	switch {
	case g.owned[id]:
		return "[owned]"
	case n.Password != "":
		return "[password]"
	case n.NeedsKey != "":
		return "[key + " + breachNames[n.Breach] + "]"
	}
	return "[" + breachNames[n.Breach] + "]"
}

var breachNames = map[BreachKind]string{NoBreach: "open", BreachICE: "ICE", BreachSync: "SYNC", BreachGates: "GATES"}

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
		g.fail("cat: " + name + ": no such file")
		return
	}
	if f.Encrypted && !g.decrypted[g.node+"/"+name] {
		g.say(LineOutput, "▓▒░▓▓░▒▒▓░░▒▓▒░▓▒▓░░▒▓▓▒░▒")
		g.say(LineOutput, "(encrypted: decrypt <file> <key>)")
		return
	}
	for _, l := range f.Body {
		g.say(LineOutput, l)
	}
	if f.Final {
		g.won, g.over = true, true
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
		g.fail("connect: " + id + ": unknown host")
	case id == g.node:
		g.say(LineOutput, "already on "+id)
	case g.owned[id]:
		g.enter(id)
	case !g.reachable(id):
		g.fail("connect: " + id + ": no route from your machines")
	case n.Password != "":
		g.password = id
	case n.NeedsKey != "" && !g.decrypted[g.keyHolder(n.NeedsKey)+"/"+n.NeedsKey]:
		g.fail("access denied: " + id + " asks for its admin key")
	case n.Breach != NoBreach:
		g.mode, g.target = ModeBreach, id
		g.breach = g.newBreach(n.Breach)
	default:
		g.enter(id)
	}
}

// keyHolder is the node that holds a file.
func (g *Game) keyHolder(name string) string {
	for _, n := range network {
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
	if line != g.nodes[id].Password {
		g.say(LineError, "access denied")
		g.raise(TracePassword)
		return
	}
	g.enter(id)
}

// enter owns id and moves there.
func (g *Game) enter(id string) {
	g.owned[id] = true
	g.node = id
	g.say(LineSuccess, "connected to "+id+" ("+g.nodes[id].Title+")")
}

func (g *Game) decrypt(name, key string) {
	f, ok := g.file(name)
	switch {
	case !ok:
		g.fail("decrypt: " + name + ": no such file")
	case !f.Encrypted:
		g.fail("decrypt: " + name + ": not encrypted")
	case key != f.Key:
		g.say(LineError, "decrypt: wrong key")
		g.raise(TracePassword)
	default:
		g.decrypted[g.node+"/"+name] = true
		for _, l := range f.Body {
			g.say(LineSuccess, l)
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
		for _, n := range network {
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

// fail reports a failed command and raises the trace for it.
func (g *Game) fail(text string) {
	g.say(LineError, text)
	g.raise(TraceFailed)
}

func (g *Game) raise(points int) {
	g.trace = min(g.trace+points, 100)
	if g.trace == 100 {
		g.say(LineError, "TRACE COMPLETE. They know where you are.")
		g.over = true
	}
}

// Prompt is the terminal prompt: the shell's, or a password request.
func (g *Game) Prompt() string {
	if g.password != "" {
		return "password for " + g.password + ": "
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
func (g *Game) Won() bool               { return g.won }
func (g *Game) Over() bool              { return g.over }
func (g *Game) Elapsed() time.Duration  { return time.Duration(g.ticks) * TickPeriod }
func (g *Game) Ticks() int              { return g.ticks }
