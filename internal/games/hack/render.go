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
	f.Draw(screen, fmt.Sprintf("Node: %s   Trace: %s%s %d%%   Time: %02d:%02d",
		g.Node(), strings.Repeat("█", filled), strings.Repeat("░", traceBarCells-filled), g.Trace(), s/60, s%60))
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

// wrap cuts text into rows of at most Cols characters.
func wrap(text string) []string {
	r := []rune(text)
	if len(r) == 0 {
		return []string{""}
	}
	var out []string
	for len(r) > Cols {
		out = append(out, string(r[:Cols]))
		r = r[Cols:]
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
	for _, n := range network {
		for _, id := range n.Links {
			a, b := centre(n), centre(*nodeByID(id))
			steps := max(abs(b.X-a.X), abs(b.Y-a.Y))
			for i := 1; i < steps; i++ {
				put(a.X+(b.X-a.X)*i/steps, a.Y+(b.Y-a.Y)*i/steps, "·", colorLink)
			}
		}
	}
	for _, n := range network {
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
		if n.ID == g.Node() {
			text += "@"
		}
		put(n.Pos.X, n.Pos.Y, text, color)
		put(n.Pos.X, n.Pos.Y+1, g.Status(n.ID), color)
	}
	put(0, Rows-1, "[arrows] select  [enter] connect  [backspace] terminal   @ you are here", colorCommand)
}

func nodeByID(id string) *Node {
	for i := range network {
		if network[i].ID == id {
			return &network[i]
		}
	}
	return nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func drawBreach(put putFunc, b Breach) {
	header := func(title, goal, keys string) {
		put(0, 0, "BREACH: "+title+"   "+goal, colorError)
		left := fmt.Sprintf("%2ds", (b.TicksLeft()+59)/60)
		put(Cols-len(left), 0, left, colorSuccess)
		put(0, Rows-1, keys, colorCommand)
	}
	switch b := b.(type) {
	case *ice:
		header("ICE", "reach the port [] before a sentry reaches you", "[arrows] move")
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
		header("SYNC", "lock the signal inside the key window, 3 times", "[space] lock")
		x0, y := (Cols-2*SyncBar)/2, 8
		for i := 0; i < SyncBar; i++ {
			sprite, color := "──", colorLink
			if i >= b.Window() && i < b.Window()+b.Width() {
				sprite, color = "██", colorOwned
			}
			put(x0+2*i, y, sprite, color)
		}
		put(x0+2*b.Cursor(), y-1, spriteCursor, colorSuccess)
		put(x0, y+3, "locked "+marks(b.Hits(), SyncHits, "●", "○")+"   missed "+marks(b.Misses(), SyncMisses, "✕", "○"), colorCommand)
	case *gates:
		header("GATES", "make every output match what the core wants", "[←→] choose  [↑↓] set  [space] flip")
		sw, gs, out, want := b.Switches(), b.Gates(), b.Outputs(), b.Target()
		for i := range sw {
			x := 8 + 12*i
			color := colorCommand
			if i == b.Cursor() {
				color = colorSelected
				put(x+1, 6, "▲", colorSelected)
			}
			put(x, 5, "["+bit(sw[i])+"]", color)
		}
		for i := range gs {
			x := 13 + 12*i
			match := colorError
			if out[i] == want[i] {
				match = colorOwned
			}
			put(x, 8, gs[i].String(), colorCommand)
			put(x, 10, "out "+bit(out[i]), match)
			put(x, 11, "want "+bit(want[i]), colorSuccess)
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
