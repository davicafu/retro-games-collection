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
	if _, _, ok := find(rows, "comando no encontrado: nope"); !ok {
		t.Fatalf("earlier output must show above the prompt")
	}
	_, erow, _ := find(rows, "comando no encontrado")
	if colors[erow][0] != colorError {
		t.Fatalf("errors must be drawn in their colour")
	}
	_ = col
	for _, want := range []string{"Nivel 1/5", "gate", "Rastreo", "4%", "Pistas 0"} {
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
		b := newICE(&scriptedRand{}, easy)
		b.walls = nil
		b.sentries = nil
		return b
	}
	run(g, "connect dev")
	rows, _, _ := drawn(g)
	board := strings.Join(rows, "\n")
	if !strings.Contains(board, spriteIntruder) || !strings.Contains(board, spritePort) || !strings.Contains(board, "BRECHA: ICE") {
		t.Fatalf("the ICE breach must show the intruder and the port:\n%s", board)
	}
}

func TestDrawTheSyncAndGatesBreaches(t *testing.T) {
	g := newTestGame(Won)
	g.newBreach = func(BreachKind) Breach { return newSync(&scriptedRand{}, easy) }
	run(g, "connect mail")
	rows, _, _ := drawn(g)
	if _, _, ok := find(rows, "SYNC"); !ok || !strings.Contains(strings.Join(rows, ""), spriteCursor) {
		t.Fatalf("the sync breach must show its bar and cursor:\n%s", strings.Join(rows, "\n"))
	}
	g = newTestGame(Won)
	g.newBreach = func(BreachKind) Breach { return newGates(&scriptedRand{}, easy) }
	run(g, "connect mail")
	rows, _, _ = drawn(g)
	board := strings.Join(rows, "\n")
	if !strings.Contains(board, "GATES") || !strings.Contains(board, "[0]") || !strings.Contains(board, "pide") {
		t.Fatalf("the gates breach must show switches, gates and the target:\n%s", board)
	}
}

func TestEveryViewFitsThePlayfield(t *testing.T) {
	g := newTestGame(Won)
	type box struct {
		id     string
		row    int
		x0, x1 int // columns [x0, x1)
	}
	for lvl := range levels {
		g.loadLevel(lvl)
		var boxes []box
		for _, n := range g.Nodes() {
			status := len([]rune(g.Status(n.ID)))
			boxes = append(boxes, box{n.ID, n.Pos.Y, n.Pos.X, n.Pos.X + len(n.ID) + 3}, box{n.ID, n.Pos.Y + 1, n.Pos.X, n.Pos.X + max(status, 6)})
			if n.Pos.Y+1 >= Rows-1 || n.Pos.Y < 0 {
				t.Fatalf("level %d: %s at row %d leaves the map", lvl+1, n.ID, n.Pos.Y)
			}
		}
		for i, a := range boxes {
			if a.x1 > Cols {
				t.Fatalf("level %d: %s's label ends at column %d, past %d", lvl+1, a.id, a.x1, Cols)
			}
			for _, b := range boxes[i+1:] {
				if a.id != b.id && a.row == b.row && a.x0 < b.x1+1 && b.x0 < a.x1+1 {
					t.Fatalf("level %d: the labels of %s and %s touch on row %d", lvl+1, a.id, b.id, a.row)
				}
			}
		}
	}
	for lvl, d := range difficulties {
		for _, b := range []Breach{newICE(&scriptedRand{}, d), newSync(&scriptedRand{}, d), newGates(&scriptedRand{}, d)} {
			g.mode, g.breach = ModeBreach, b
			rows, _, _ := drawn(g)
			if !strings.HasSuffix(strings.TrimRight(rows[0], " "), "s") || strings.Count(rows[0], "   ") < 2 {
				t.Fatalf("level %d %v: the header must keep its goal clear of the timer: %q", lvl+1, b.Kind(), rows[0])
			}
			if gb, ok := b.(*gates); ok {
				board := strings.Join(rows, "\n")
				if strings.Count(board, "[0]") != len(gb.Switches()) || strings.Count(board, "pide 0")+strings.Count(board, "pide 1") != len(gb.Gates()) {
					t.Fatalf("level %d: every switch and gate must show:\n%s", lvl+1, board)
				}
			}
		}
	}
}

func TestTheHUDFitsTheWorstCase(t *testing.T) {
	g := newTestGame(Won)
	at(g, 2, "backup")
	g.trace, g.hints, g.ticks = 100, 10, 60*(99*60+59)
	_, _, hud := drawn(g)
	for _, want := range []string{"Nivel 3/5", "backup", "100%", "Pistas 10", "99:59"} {
		if !strings.Contains(hud, want) {
			t.Fatalf("HUD %q missing %q", hud, want)
		}
	}
}

func TestLongLinesWrapBetweenWords(t *testing.T) {
	word := strings.Repeat("a", 10)
	line := strings.TrimSpace(strings.Repeat(word+" ", 8)) // 87 columns
	got := wrap(line)
	if len(got) != 2 || got[0] != strings.TrimSpace(strings.Repeat(word+" ", 6)) || got[1] != word+" "+word {
		t.Fatalf("wrap = %q, want the break at the last space that fits", got)
	}
	if got := wrap(strings.Repeat("x", Cols+5)); len(got) != 2 || len([]rune(got[0])) != Cols {
		t.Fatalf("a word longer than a row must still be cut, got %q", got)
	}
}

func TestTheMapAndScanMarkWatchedMachines(t *testing.T) {
	g := newTestGame(Won)
	at(g, 3, "gate")
	run(g, "scan")
	if out := output(g); !strings.Contains(out, "vigilada") || strings.Count(out, "vigilada") != 2 {
		t.Fatalf("scan must mark puerto and correo as watched:\n%s", out)
	}
	run(g, "map")
	rows, _, _ := drawn(g)
	board := strings.Join(rows, "\n")
	if !strings.Contains(board, "[puerto]!") || strings.Contains(board, "[gate]!") || !strings.Contains(board, "! vigilada") {
		t.Fatalf("watched machines must carry a ! on the map, with a legend:\n%s", board)
	}
	g.owned["ids"] = true
	rows, _, _ = drawn(g)
	if strings.Contains(strings.Join(rows, "\n"), "[puerto]!") {
		t.Fatalf("with the ids yours nothing is watched")
	}
}
