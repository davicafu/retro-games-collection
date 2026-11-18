package hack

import (
	"fmt"
	"strings"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// Cols and Rows are the playfield's size in characters.
const (
	Cols = 72
	Rows = 18

	traceBarCells = 10
	// CursorBlink is how many ticks the prompt's cursor stays shown or hidden.
	CursorBlink = 30

	spriteIntruder = "<>"
	spritePort     = "[]"
	spriteSentry   = "◆◆"
	spriteWall     = "▓▓"
	spriteFloor    = "· "
	spriteCursor   = "▼"
)

// Colours: the terminal is green with red errors and yellow successes;
// the map shows owned machines green, reachable ones yellow and the rest
// red, with the selection cyan.
const (
	colorOutput    = render.ColorGreen
	colorCommand   = render.ColorDefault
	colorError     = render.ColorRed
	colorSuccess   = render.ColorYellow
	colorOwned     = render.ColorGreen
	colorReachable = render.ColorYellow
	colorLocked    = render.ColorRed
	colorSelected  = render.ColorCyan
	colorLink      = render.ColorBlue
)

var lineColors = map[LineKind]render.Color{LineOutput: colorOutput, LineCommand: colorCommand, LineError: colorError, LineSuccess: colorSuccess}

// Draw renders the HUD and the terminal, the map or the breach under way.
func Draw(screen render.Screen, g *Game) {
	f := render.CenterFrame(screen, Cols, Rows)
	s := int(g.Elapsed().Seconds())
	filled := g.Trace() * traceBarCells / 100
	f.Draw(screen, fmt.Sprintf("Nivel %d/%d  %s  Rastreo %s%s %d%%  Pistas %d  %02d:%02d",
		g.Level()+1, len(levels), g.Node(), strings.Repeat("█", filled), strings.Repeat("░", traceBarCells-filled), g.Trace(), g.Hints(), s/60, s%60))
	put := func(col, row int, text string, color render.Color) {
		if row < 0 || row >= Rows {
			return
		}
		x, y := f.Cell(col, row)
		render.WriteColoredText(screen, x, y, text, color)
	}
	switch g.Mode() {
	case ModeTerminal:
		drawTerminal(put, g)
	case ModeMap:
		drawMap(put, g)
	case ModeBreach:
		drawBreach(put, g.Breach())
	}
}

type putFunc func(col, row int, text string, color render.Color)

// wrap cuts text into rows of at most Cols characters, between words
// where it can.
func wrap(text string) []string {
	r := []rune(text)
	if len(r) == 0 {
		return []string{""}
	}
	var out []string
	for len(r) > Cols {
		cut, next := Cols, Cols
		for i := Cols; i > 0; i-- {
			if r[i] == ' ' {
				cut, next = i, i+1
				break
			}
		}
		out = append(out, string(r[:cut]))
		r = r[next:]
	}
	return append(out, string(r))
}

func drawTerminal(put putFunc, g *Game) {
	type row struct {
		text  string
		color render.Color
	}
	var rows []row
	for _, l := range g.Lines() {
		for _, w := range wrap(l.Text) {
			rows = append(rows, row{w, lineColors[l.Kind]})
		}
	}
	rows = rows[max(len(rows)-(Rows-1), 0):]
	for i, r := range rows {
		put(0, Rows-1-len(rows)+i, r.text, r.color)
	}
	prompt := []rune(g.Prompt() + g.Buffer())
	if g.Ticks()/CursorBlink%2 == 0 {
		prompt = append(prompt, '█')
	}
	prompt = prompt[max(len(prompt)-Cols, 0):]
	put(0, Rows-1, string(prompt), colorCommand)
}

func drawMap(put putFunc, g *Game) {
	label := func(n Node) string { return "[" + n.ID + "]" }
	centre := func(n Node) grid.Point { return grid.Point{X: n.Pos.X + len(label(n))/2, Y: n.Pos.Y} }
	for _, n := range g.Nodes() {
		for _, id := range n.Links {
			a, b := centre(n), centre(*nodeByID(g, id))
			steps := max(abs(b.X-a.X), abs(b.Y-a.Y))
			for i := 1; i < steps; i++ {
				put(a.X+(b.X-a.X)*i/steps, a.Y+(b.Y-a.Y)*i/steps, "·", colorLink)
			}
		}
	}
	for _, n := range g.Nodes() {
		color := colorLocked
		switch {
		case n.ID == g.Selected():
			color = colorSelected
		case g.Owned(n.ID):
			color = colorOwned
		case g.reachable(n.ID):
			color = colorReachable
		}
		text := label(n)
		if g.Watched(n.ID) {
			text += "!"
		}
		if n.ID == g.Node() {
			text += "@"
		}
		put(n.Pos.X, n.Pos.Y, text, color)
		put(n.Pos.X, n.Pos.Y+1, g.Status(n.ID), color)
	}
	put(0, Rows-1, "[flechas] elegir  [enter] conectar  [backspace] volver  @ tú  ! vigilada", colorCommand)
}

func nodeByID(g *Game, id string) *Node {
	return g.nodes[id]
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func drawBreach(put putFunc, b Breach) {
	header := func(title, goal, keys string) {
		put(0, 0, "BRECHA: "+title+"   "+goal, colorError)
		left := fmt.Sprintf("%2ds", (b.TicksLeft()+59)/60)
		put(Cols-len(left), 0, left, colorSuccess)
		put(0, Rows-1, keys, colorCommand)
	}
	switch b := b.(type) {
	case *ice:
		header("ICE", "llega al puerto [] sin que te pille un centinela", "[flechas] moverte")
		x0, y0 := (Cols-2*ICEWidth)/2, 4
		for y := 0; y < ICEHeight; y++ {
			for x := 0; x < ICEWidth; x++ {
				p := grid.Point{X: x, Y: y}
				sprite, color := spriteFloor, colorLink
				switch {
				case p == b.Port():
					sprite, color = spritePort, colorOwned
				case b.Wall(p):
					sprite, color = spriteWall, colorLink
				}
				put(x0+2*x, y0+y, sprite, color)
			}
		}
		for _, s := range b.Sentries() {
			put(x0+2*s.X, y0+s.Y, spriteSentry, colorError)
		}
		put(x0+2*b.Player().X, y0+b.Player().Y, spriteIntruder, colorSelected)
	case *syncBreach:
		header("SYNC", "fija la señal dentro de la ventana, 3 veces", "[espacio] fijar")
		x0, y := (Cols-2*SyncBar)/2, 8
		for i := 0; i < SyncBar; i++ {
			sprite, color := "──", colorLink
			if i >= b.Window() && i < b.Window()+b.Width() {
				sprite, color = "██", colorOwned
			}
			put(x0+2*i, y, sprite, color)
		}
		put(x0+2*b.Cursor(), y-1, spriteCursor, colorSuccess)
		put(x0, y+3, "fijadas "+marks(b.Hits(), b.Need(), "●", "○")+"   falladas "+marks(b.Misses(), SyncMisses, "✕", "○"), colorCommand)
	case *gates:
		header("GATES", "haz que cada salida sea la que pide el core", "[←→] elegir  [↑↓] poner  [espacio] cambiar")
		sw, gs, out, want := b.Switches(), b.Gates(), b.Outputs(), b.Target()
		// Switches spread over the playfield: 12 columns apart for 5, fewer
		// for more; each gate sits between the two switches it reads.
		gap := min(12, (Cols-8)/len(sw))
		left := (Cols - gap*(len(sw)-1) - 3) / 2
		for i := range sw {
			x := left + gap*i
			color := colorCommand
			if i == b.Cursor() {
				color = colorSelected
				put(x+1, 6, "▲", colorSelected)
			}
			put(x, 5, "["+bit(sw[i])+"]", color)
		}
		for i := range gs {
			x := left + gap*i + gap/2
			match := colorError
			if out[i] == want[i] {
				match = colorOwned
			}
			put(x, 8, gs[i].String(), colorCommand)
			put(x, 10, "sale "+bit(out[i]), match)
			put(x, 11, "pide "+bit(want[i]), colorSuccess)
		}
	}
}

func bit(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// marks is n filled marks out of total.
func marks(n, total int, on, off string) string {
	return strings.Repeat(on, n) + strings.Repeat(off, total-n)
}
