package hack

import (
	"github.com/davidbarreiro/cli-games-collection/internal/engine/glide"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// Outcome is how a breach stands.
type Outcome int

const (
	Running Outcome = iota
	Won
	Lost
)

// Breach is a short timed arcade challenge guarding a node.
type Breach interface {
	Step(in Input)
	Outcome() Outcome
	TicksLeft() int
	Kind() BreachKind
}

// breachTicks is each kind's time limit: 20 s, 20 s and 30 s.
var breachTicks = map[BreachKind]int{BreachICE: 1200, BreachSync: 1200, BreachGates: 1800}

// timer counts a breach's time down and holds its outcome.
type timer struct {
	left    int
	outcome Outcome
}

// tick spends a tick; running out of time loses.
func (c *timer) tick() {
	if c.outcome != Running {
		return
	}
	if c.left--; c.left <= 0 {
		c.outcome = Lost
	}
}

func (c *timer) Outcome() Outcome { return c.outcome }
func (c *timer) TicksLeft() int   { return c.left }

// ICE: reach the port across a grid with walls while sentries chase.
const (
	ICEWidth  = 21
	ICEHeight = 9
	// The intruder moves at about 8 cells/s while a key is held; a tap
	// moves one cell. Sentries move a cell every SentryTicks (4 cells/s).
	ICESpeed       = 13
	ICEGlide       = 7
	SentryTicks    = 15
	Sentries       = 3
	SentryDistance = 8
	ICEWalls       = 16
)

type ice struct {
	timer
	walls    map[grid.Point]bool
	player   grid.Point
	port     grid.Point
	sentries []grid.Point
	walk     glide.Glide
	facing   grid.Point
	sentryAt int
}

func newICE(rng Rand) *ice {
	b := &ice{
		timer:  timer{left: breachTicks[BreachICE]},
		player: grid.Point{X: 0, Y: ICEHeight / 2},
		port:   grid.Point{X: ICEWidth - 1, Y: ICEHeight / 2},
		walk:   glide.Glide{Speed: ICESpeed, Ticks: ICEGlide},
	}
	for {
		b.walls = map[grid.Point]bool{}
		for len(b.walls) < ICEWalls {
			b.walls[grid.Point{X: 2 + rng.Intn(ICEWidth-4), Y: rng.Intn(ICEHeight)}] = true
		}
		if grid.Distances(ICEWidth, ICEHeight, b.port, b.open).At(b.player) != grid.Unreachable {
			break
		}
	}
	for len(b.sentries) < Sentries {
		p := grid.Point{X: rng.Intn(ICEWidth), Y: rng.Intn(ICEHeight)}
		if b.open(p) && p != b.port && manhattan(p, b.player) >= SentryDistance {
			b.sentries = append(b.sentries, p)
		}
	}
	return b
}

func (b *ice) open(p grid.Point) bool {
	return p.X >= 0 && p.X < ICEWidth && p.Y >= 0 && p.Y < ICEHeight && !b.walls[p]
}

func (b *ice) Step(in Input) {
	if b.outcome != Running {
		return
	}
	if in.Move != (grid.Point{}) {
		if in.Move != b.facing {
			b.walk.Stop()
			b.facing = in.Move
		}
		b.walk.Press(1)
	}
	if b.walk.Step() != 0 {
		if next := b.player.Add(b.facing); b.open(next) {
			b.player = next
		}
	}
	if b.caught() {
		return
	}
	if b.player == b.port {
		b.outcome = Won
		return
	}
	if b.sentryAt++; b.sentryAt >= SentryTicks {
		b.sentryAt = 0
		chase := grid.Distances(ICEWidth, ICEHeight, b.player, b.open)
		for i, s := range b.sentries {
			best := s
			for _, d := range grid.Neighbours {
				// Unreachable cells (walls, off the grid) read as -1.
				if n := s.Add(d); chase.At(n) != grid.Unreachable && chase.At(n) < chase.At(best) {
					best = n
				}
			}
			b.sentries[i] = best
		}
	}
	if b.caught() {
		return
	}
	b.tick()
}

func (b *ice) caught() bool {
	for _, s := range b.sentries {
		if s == b.player {
			b.outcome = Lost
			return true
		}
	}
	return false
}

func (b *ice) Kind() BreachKind       { return BreachICE }
func (b *ice) Player() grid.Point     { return b.player }
func (b *ice) Port() grid.Point       { return b.port }
func (b *ice) Wall(p grid.Point) bool { return b.walls[p] }
func (b *ice) Sentries() []grid.Point { return append([]grid.Point(nil), b.sentries...) }

func manhattan(a, b grid.Point) int {
	dx, dy := a.X-b.X, a.Y-b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// Sync: stop a sweeping cursor inside the key window SyncHits times before
// missing SyncMisses times. The window narrows after every hit.
const (
	SyncBar    = 30
	SyncWindow = 5
	SyncHits   = 3
	SyncMisses = 3
	// SyncTicks is how many ticks the cursor takes per cell: 20 cells/s.
	SyncTicks = 3
)

type syncBreach struct {
	timer
	rng      Rand
	cursor   int
	dir      int
	window   int
	width    int
	hits     int
	misses   int
	cursorAt int
}

func newSync(rng Rand) *syncBreach {
	b := &syncBreach{timer: timer{left: breachTicks[BreachSync]}, rng: rng, dir: 1, width: SyncWindow}
	b.place()
	return b
}

// place puts the window somewhere new along the bar.
func (b *syncBreach) place() { b.window = b.rng.Intn(SyncBar - b.width + 1) }

func (b *syncBreach) Step(in Input) {
	if b.outcome != Running {
		return
	}
	if in.Act {
		if b.cursor >= b.window && b.cursor < b.window+b.width {
			b.hits++
			b.width = max(b.width-1, 1)
			b.place()
		} else {
			b.misses++
		}
		switch {
		case b.hits >= SyncHits:
			b.outcome = Won
			return
		case b.misses >= SyncMisses:
			b.outcome = Lost
			return
		}
	}
	if b.cursorAt++; b.cursorAt >= SyncTicks {
		b.cursorAt = 0
		b.cursor += b.dir
		if b.cursor <= 0 || b.cursor >= SyncBar-1 {
			b.dir = -b.dir
		}
	}
	b.tick()
}

func (b *syncBreach) Kind() BreachKind { return BreachSync }
func (b *syncBreach) Cursor() int      { return b.cursor }
func (b *syncBreach) Window() int      { return b.window }
func (b *syncBreach) Width() int       { return b.width }
func (b *syncBreach) Hits() int        { return b.hits }
func (b *syncBreach) Misses() int      { return b.misses }

// Gates: set the input switches so that every gate's output matches the
// target. Each gate reads two neighbouring switches.
const Switches = 5

// Gate is a logic gate of two inputs.
type Gate int

const (
	GateAnd Gate = iota
	GateOr
	GateXor
	GateNand
	gateKinds
)

var gateNames = [...]string{GateAnd: "AND", GateOr: "OR", GateXor: "XOR", GateNand: "NAND"}

func (g Gate) String() string { return gateNames[g] }

// Eval is the gate's output for inputs a and b.
func (g Gate) Eval(a, b bool) bool {
	switch g {
	case GateAnd:
		return a && b
	case GateOr:
		return a || b
	case GateXor:
		return a != b
	}
	return !(a && b)
}

type gates struct {
	timer
	switches [Switches]bool
	gates    [Switches - 1]Gate
	target   [Switches - 1]bool
	cursor   int
}

// newGates deals random gates and a target that some setting reaches but
// the starting one, all off, does not.
func newGates(rng Rand) *gates {
	b := &gates{timer: timer{left: breachTicks[BreachGates]}}
	for {
		for i := range b.gates {
			b.gates[i] = Gate(rng.Intn(int(gateKinds)))
		}
		var secret [Switches]bool
		for i := range secret {
			secret[i] = rng.Intn(2) == 1
		}
		b.target = b.outputsFor(secret)
		if !b.Solved() {
			return b
		}
	}
}

func (b *gates) outputsFor(sw [Switches]bool) [Switches - 1]bool {
	var out [Switches - 1]bool
	for i, g := range b.gates {
		out[i] = g.Eval(sw[i], sw[i+1])
	}
	return out
}

func (b *gates) Solved() bool { return b.outputsFor(b.switches) == b.target }

func (b *gates) Step(in Input) {
	if b.outcome != Running {
		return
	}
	switch in.Move {
	case grid.Point{X: -1}:
		b.cursor = max(b.cursor-1, 0)
	case grid.Point{X: 1}:
		b.cursor = min(b.cursor+1, Switches-1)
	case grid.Point{Y: -1}:
		b.switches[b.cursor] = true
	case grid.Point{Y: 1}:
		b.switches[b.cursor] = false
	}
	if in.Act {
		b.switches[b.cursor] = !b.switches[b.cursor]
	}
	if b.Solved() {
		b.outcome = Won
		return
	}
	b.tick()
}

func (b *gates) Kind() BreachKind            { return BreachGates }
func (b *gates) Cursor() int                 { return b.cursor }
func (b *gates) Switches() [Switches]bool    { return b.switches }
func (b *gates) Gates() [Switches - 1]Gate   { return b.gates }
func (b *gates) Target() [Switches - 1]bool  { return b.target }
func (b *gates) Outputs() [Switches - 1]bool { return b.outputsFor(b.switches) }
