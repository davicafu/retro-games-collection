package hack

import (
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// drawn renders g on an 80x24 screen and returns the playfield rows, their
// colours and the HUD row.
func drawn(g *Game) (rows []string, colors [][]render.Color, hud string) {
	s := render.NewBufferScreen(80, 24)
	Draw(s, g)
	all, cs := s.Rows(), s.Colors()
	f := render.CenterFrame(s, Cols, Rows)
	for row := 0; row < Rows; row++ {
		x, y := f.Cell(0, row)
		rows = append(rows, string(all[y][x:x+Cols]))
		colors = append(colors, cs[y][x:x+Cols])
	}
	return rows, colors, string(all[f.Y])
}

func find(rows []string, s string) (col, row int, ok bool) {
	for y, r := range rows {
		if i := strings.Index(r, s); i >= 0 {
			return len([]rune(r[:i])), y, true
		}
	}
	return 0, 0, false
}

func TestDrawTheTerminalWithItsPromptAndColours(t *testing.T) {
	g := newTestGame(Won)
	run(g, "nope")
	g.Step(Input{Text: []rune("ls")})
	rows, colors, hud := drawn(g)
	col, row, ok := find(rows, "ghost@gate:~$ ls")
	if !ok || row != Rows-1 {
		t.Fatalf("the prompt with the typed line must be the last row:\n%s", strings.Join(rows, "\n"))
	}
	if _, _, ok := find(rows, "command not found: nope"); !ok {
		t.Fatalf("earlier output must show above the prompt")
	}
	_, erow, _ := find(rows, "command not found")
	if colors[erow][0] != colorError {
		t.Fatalf("errors must be drawn in their colour")
	}
	_ = col
	for _, want := range []string{"gate", "Trace", "4%"} {
		if !strings.Contains(hud, want) {
			t.Fatalf("HUD %q missing %q", hud, want)
		}
	}
}

func TestDrawTheTerminalScrollsAndWrapsLongLines(t *testing.T) {
	g := newTestGame(Won)
	for i := 0; i < 30; i++ {
		run(g, "scan")
	}
	g.say(LineOutput, strings.Repeat("x", Cols+5))
	rows, _, _ := drawn(g)
	if !strings.HasPrefix(rows[Rows-3], strings.Repeat("x", Cols)) || !strings.HasPrefix(rows[Rows-2], "xxxxx ") {
		t.Fatalf("a long line must wrap onto the next row, newest at the bottom:\n%s", strings.Join(rows, "\n"))
	}
}

func TestDrawTheMapWithNodesAndTheSelection(t *testing.T) {
	g := newTestGame(Won)
	run(g, "map")
	g.Step(Input{Move: right})
	rows, colors, _ := drawn(g)
	for _, id := range []string{"gate", "mail", "dev", "vault", "core"} {
		if _, _, ok := find(rows, id); !ok {
			t.Fatalf("the map must show %s:\n%s", id, strings.Join(rows, "\n"))
		}
	}
	c, r, _ := find(rows, "mail")
	if colors[r][c] != colorSelected {
		t.Fatalf("the selected node must stand out")
	}
	c, r, _ = find(rows, "gate")
	if colors[r][c] != colorOwned {
		t.Fatalf("owned nodes must show as owned")
	}
}

func TestDrawTheICEBreach(t *testing.T) {
	g := newTestGame(Won)
	g.newBreach = func(BreachKind) Breach {
		b := newICE(&scriptedRand{})
		b.walls = nil
		b.sentries = nil
		return b
	}
	run(g, "connect dev")
	rows, _, _ := drawn(g)
	board := strings.Join(rows, "\n")
	if !strings.Contains(board, spriteIntruder) || !strings.Contains(board, spritePort) || !strings.Contains(board, "ICE") {
		t.Fatalf("the ICE breach must show the intruder and the port:\n%s", board)
	}
}

func TestDrawTheSyncAndGatesBreaches(t *testing.T) {
	g := newTestGame(Won)
	g.newBreach = func(BreachKind) Breach { return newSync(&scriptedRand{}) }
	run(g, "connect mail")
	rows, _, _ := drawn(g)
	if _, _, ok := find(rows, "SYNC"); !ok || !strings.Contains(strings.Join(rows, ""), spriteCursor) {
		t.Fatalf("the sync breach must show its bar and cursor:\n%s", strings.Join(rows, "\n"))
	}
	g = newTestGame(Won)
	g.newBreach = func(BreachKind) Breach { return newGates(&scriptedRand{}) }
	run(g, "connect mail")
	rows, _, _ = drawn(g)
	board := strings.Join(rows, "\n")
	if !strings.Contains(board, "GATES") || !strings.Contains(board, "[0]") || !strings.Contains(board, "want") {
		t.Fatalf("the gates breach must show switches, gates and the target:\n%s", board)
	}
}

func TestEveryViewFitsThePlayfield(t *testing.T) {
	g := newTestGame(Won)
	for _, n := range network {
		if end := n.Pos.X + len([]rune(g.Status(n.ID))); end > Cols || n.Pos.X+len(n.ID)+3 > Cols {
			t.Fatalf("%s's map label ends at column %d, past %d", n.ID, end, Cols)
		}
	}
	for _, b := range []Breach{newICE(&scriptedRand{}), newSync(&scriptedRand{}), newGates(&scriptedRand{})} {
		g.mode, g.breach = ModeBreach, b
		rows, _, _ := drawn(g)
		if !strings.HasSuffix(strings.TrimRight(rows[0], " "), "s") || strings.Count(rows[0], "   ") < 2 {
			t.Fatalf("%v: the header must keep its goal clear of the timer: %q", b.Kind(), rows[0])
		}
	}
}
