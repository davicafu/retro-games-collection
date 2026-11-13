package leak

import (
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// drawn renders g on an 80x24 screen and returns the playfield rows (inside
// the box) and the HUD row.
func drawn(g *Game) (rows [][]rune, hud string) {
	s := render.NewBufferScreen(80, 24)
	Draw(s, g)
	all := s.Rows()
	f := render.CenterFrame(s, 2*(g.Width()-2), g.Height()-2)
	for row := 0; row < g.Height()-2; row++ {
		x, y := f.Cell(0, row)
		rows = append(rows, all[y][x:x+2*(g.Width()-2)])
	}
	return rows, string(all[f.Y])
}

// chars returns n characters of a row from column from.
func chars(row []rune, from, n int) string { return string(row[from : from+n]) }

func joined(rows [][]rune) string {
	var lines []string
	for _, r := range rows {
		lines = append(lines, string(r))
	}
	return strings.Join(lines, "\n")
}

var small = []string{
	"#######",
	"#L.o###",
	"#.##..#",
	"#.....#",
	"#######",
}

func TestDrawMapTiles(t *testing.T) {
	g := newTestGame(small...)
	g.started = true // no READY banner
	rows, _ := drawn(g)
	// A wall connects only to walls that are drawn: (5,1) sits under a
	// block's blank inside, so it is a plain line into the right edge.
	want := []string{
		"◖◗· ◆ └───",
		"· ─── · · ",
		"· · · · · ",
	}
	for i := range want {
		if string(rows[i]) != want[i] {
			t.Fatalf("row %d = %q, want %q\nall:\n%s", i, string(rows[i]), want[i], joined(rows))
		}
	}
}

func TestDrawFillsSolidBlocksAndShowsThePoolGate(t *testing.T) {
	rows, _ := drawn(newTestGame(RAM...))
	// Map row 7 holds the left solid block (columns 1-3 blank inside, its
	// right edge at column 3) and row 8 the pool's top wall with its gate.
	if got := chars(rows[6], 0, 6); got != "    │ " {
		t.Fatalf("inside of the left block = %q, want blank with its edge", got)
	}
	// Glyphs sit on the left character of their cell, so a wall that ends
	// leaves the right character blank.
	if got := chars(rows[7], 12, 10); got != "┌───╌╌──┐ " {
		t.Fatalf("pool top = %q, want %q", got, "┌───╌╌──┐ ")
	}
}

func TestDrawActorsInHalfCellSteps(t *testing.T) {
	g := newTestGame(small...)
	g.leak = Actor{Cell: pt(1, 3), Dir: Right, Progress: CellUnits/2 - 1}
	rows, _ := drawn(g)
	if got := chars(rows[2], 0, 2); got != spriteLeak {
		t.Fatalf("before half a cell the leak stays on its cell: %q", string(rows[2]))
	}
	g.leak.Progress = CellUnits / 2
	rows, _ = drawn(g)
	if got := chars(rows[2], 1, 2); got != spriteLeak {
		t.Fatalf("from half a cell the leak is one character ahead: %q", string(rows[2]))
	}
	g.leak = Actor{Cell: pt(1, 2), Dir: Down, Progress: CellUnits / 2}
	rows, _ = drawn(g)
	if got := chars(rows[2], 0, 2); got != spriteLeak {
		t.Fatalf("from half a cell down the leak is drawn on the next row: %q", string(rows[2]))
	}
}

func TestDrawCollectorsByKindAndWhileFleeing(t *testing.T) {
	g := newTestGame(small...)
	for i, k := range []Kind{KindTracer, KindInterceptor, KindWanderer} {
		addCollector(g, k, pt(1+2*i, 3), Right)
	}
	rows, _ := drawn(g)
	for i, k := range []Kind{KindTracer, KindInterceptor, KindWanderer} {
		if got := chars(rows[2], 4*i, 2); got != spriteCollector[k] {
			t.Fatalf("kind %d = %q, want %q", k, got, spriteCollector[k])
		}
	}

	g.collectors[0].fleeing = true
	g.fright = BlinkTicks + 1
	if rows, _ = drawn(g); chars(rows[2], 0, 2) != spriteFleeing {
		t.Fatalf("fleeing collector = %q, want %q", chars(rows[2], 0, 2), spriteFleeing)
	}
	seen := map[string]bool{}
	for g.fright = BlinkTicks; g.fright > 0; g.fright-- {
		rows, _ = drawn(g)
		seen[chars(rows[2], 0, 2)] = true
	}
	if !seen[spriteFleeing] || !seen[spriteWearing] {
		t.Fatalf("a malloc about to wear off must blink between %q and %q, saw %v", spriteFleeing, spriteWearing, seen)
	}
}

func TestDrawHUDShowsLevelScoreAndLives(t *testing.T) {
	g := newTestGame(RAM...)
	g.level, g.score, g.lives = 4, 1230, 2
	_, hud := drawn(g)
	for _, want := range []string{"Level 4", "Score: 1230", strings.Repeat(string(glyphLife), 2)} {
		if !strings.Contains(hud, want) {
			t.Fatalf("HUD %q missing %q", hud, want)
		}
	}
	if strings.Contains(hud, strings.Repeat(string(glyphLife), 3)) {
		t.Fatalf("HUD %q shows more lives than left", hud)
	}
}

func TestDrawReadyUntilTheFirstMove(t *testing.T) {
	g := newTestGame(RAM...)
	if rows, _ := drawn(g); !strings.Contains(joined(rows), bannerReady) {
		t.Fatalf("%q missing before the first move", bannerReady)
	}
	g.Step(Input{Dir: Left})
	if rows, _ := drawn(g); strings.Contains(joined(rows), bannerReady) {
		t.Fatalf("%q still shown after the first move", bannerReady)
	}
}

func TestDrawColoursTheLeakYellowAndEachCollectorKindApart(t *testing.T) {
	g := newTestGame(small...)
	kinds := []Kind{KindTracer, KindInterceptor, KindWanderer}
	for i, k := range kinds {
		addCollector(g, k, pt(1+2*i, 3), Right)
	}
	s := render.NewBufferScreen(80, 24)
	f := render.CenterFrame(s, 2*(g.Width()-2), g.Height()-2)
	colourOf := func(col, row int) [2]render.Color {
		Draw(s, g)
		x, y := f.Cell(col, row)
		c := s.Colors()
		return [2]render.Color{c[y][x], c[y][x+1]}
	}

	if got := colourOf(0, 0); got != [2]render.Color{render.ColorYellow, render.ColorYellow} {
		t.Fatalf("leak colours = %v, want yellow", got)
	}
	seen := map[render.Color]bool{render.ColorYellow: true}
	for i, k := range kinds {
		got := colourOf(4*i, 2)
		if got[0] != got[1] || got[0] == render.ColorDefault || seen[got[0]] {
			t.Fatalf("kind %d colours = %v, want one colour of its own, not the leak's or another kind's", k, got)
		}
		seen[got[0]] = true
	}

	g.collectors[0].fleeing = true
	g.fright = BlinkTicks + 1
	fleeing := colourOf(0, 2)
	if fleeing[0] == render.ColorDefault || seen[fleeing[0]] {
		t.Fatalf("fleeing colours = %v, want a colour of their own", fleeing)
	}
	if got := colourOf(2, 0); got[0] != render.ColorDefault {
		t.Fatalf("the map keeps the default colour, got %v", got)
	}
}
