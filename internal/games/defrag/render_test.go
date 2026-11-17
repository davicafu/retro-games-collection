package defrag

import (
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/polyomino"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// drawn renders g on an 80x24 screen and returns the playfield rows, their
// colours and the HUD row.
func drawn(g *Game) (rows []string, colors [][]render.Color, hud string) {
	s := render.NewBufferScreen(80, 24)
	Draw(s, g, true)
	all, cs := s.Rows(), s.Colors()
	f := render.CenterFrame(s, playfieldCols, Height)
	for row := 0; row < Height; row++ {
		x, y := f.Cell(0, row)
		rows = append(rows, string(all[y][x:x+playfieldCols]))
		colors = append(colors, cs[y][x:x+playfieldCols])
	}
	return rows, colors, string(all[f.Y])
}

// cellAt is the two characters drawn for disk cell p and their colour.
func cellAt(rows []string, colors [][]render.Color, p grid.Point) (string, render.Color) {
	return string([]rune(rows[p.Y])[2*p.X : 2*p.X+2]), colors[p.Y][2*p.X]
}

func TestDrawTheFragmentItsShadowTheDiskAndTheNext(t *testing.T) {
	g := newTestGame(Square, Tee)
	g.disk.Place(polyomino.Shape{Size: 1, Cells: []grid.Point{{}}}, pt(0, Height-1), int(Zed)+1)
	rows, colors, hud := drawn(g)
	for _, c := range []struct {
		p     grid.Point
		want  string
		color render.Color
	}{
		{pt(4, 0), spriteFragment, kindColors[Square]},
		{pt(5, 1), spriteFragment, kindColors[Square]},
		{pt(4, Height-1), spriteShadow, colorShadow},
		{pt(0, Height-1), spriteLocked, kindColors[Zed]},
		{pt(2, 5), spriteEmpty, colorEmpty},
	} {
		if got, color := cellAt(rows, colors, c.p); got != c.want || color != c.color {
			t.Fatalf("cell %v = %q in %v, want %q in %v\n%s", c.p, got, color, c.want, c.color, strings.Join(rows, "\n"))
		}
	}
	if r := []rune(rows[3])[2*Width]; r != '│' {
		t.Fatalf("the disk's right side must be drawn, got %q", r)
	}
	panel := strings.Join(rowsFrom(rows[:5], 2*Width+1), "\n")
	if !strings.Contains(panel, "NEXT") || strings.Count(panel, spriteFragment) != 4 {
		t.Fatalf("the next fragment must be shown beside the disk:\n%s", panel)
	}
	for _, want := range []string{"LEVEL", "SECTORS"} {
		if !strings.Contains(strings.Join(rowsFrom(rows, 2*Width+1), "\n"), want) {
			t.Fatalf("the side panel must show %q", want)
		}
	}
	g.score, g.level, g.sectors = 1234567, 12, 115
	rows, _, hud = drawn(g)
	side := strings.Join(rowsFrom(rows, 2*Width+1), "\n")
	if !strings.Contains(hud, "Score: 1234567") || !strings.Contains(side, "12") || !strings.Contains(side, "115") {
		t.Fatalf("HUD %q and panel\n%s\nmust show a large score, the level and the sectors", hud, side)
	}
}

func TestDrawFullRowsFlashing(t *testing.T) {
	g := newTestGame(Bar, Square)
	fill(g, Height-1, 3, 4, 5, 6)
	g.Step(Input{HardDrop: true})
	on, colors, _ := drawn(g)
	if got, color := cellAt(on, colors, pt(0, Height-1)); got != spriteFlash || color != colorFlash {
		t.Fatalf("a clearing row shows %q in %v, want the flash", got, color)
	}
	steps(g, FlashPeriod, Input{})
	off, _, _ := drawn(g)
	if got, _ := cellAt(off, colors, pt(0, Height-1)); got != spriteEmpty {
		t.Fatalf("a clearing row must blink, shows %q", got)
	}
}

// rowsFrom cuts every row from column col on.
func rowsFrom(rows []string, col int) []string {
	var out []string
	for _, r := range rows {
		out = append(out, string([]rune(r)[col:]))
	}
	return out
}

func TestGuidesOffHideTheDotsAndTheShadow(t *testing.T) {
	g := newTestGame(Square, Tee)
	s := render.NewBufferScreen(80, 24)
	Draw(s, g, false)
	all := s.Rows()
	f := render.CenterFrame(s, playfieldCols, Height)
	cell := func(p grid.Point) string {
		x, y := f.Cell(2*p.X, p.Y)
		return string(all[y][x : x+2])
	}
	if got := cell(pt(2, 5)); got != "  " {
		t.Fatalf("an empty cell shows %q without guides, want blank", got)
	}
	if got := cell(pt(4, Height-1)); got != "  " {
		t.Fatalf("the shadow shows %q without guides, want blank", got)
	}
	if got := cell(pt(4, 0)); got != spriteFragment {
		t.Fatalf("the fragment must still show, got %q", got)
	}
}
