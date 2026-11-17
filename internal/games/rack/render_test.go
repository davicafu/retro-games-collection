package rack

import (
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// drawn renders g on an 80x24 screen and returns the playfield rows, their
// colours and the HUD row.
func drawn(g *Game) (rows []string, colors [][]render.Color, hud string) {
	s := render.NewBufferScreen(80, 24)
	Draw(s, g)
	all, cs := s.Rows(), s.Colors()
	f := render.CenterFrame(s, 2*g.Width(), g.Height())
	for row := 0; row < g.Height(); row++ {
		x, y := f.Cell(0, row)
		rows = append(rows, string(all[y][x:x+2*g.Width()]))
		colors = append(colors, cs[y][x:x+2*g.Width()])
	}
	return rows, colors, string(all[f.Y])
}

// sprite is the two characters drawn for cell (x, y) and their colour.
func sprite(rows []string, colors [][]render.Color, x, y int) (string, render.Color) {
	return string([]rune(rows[y])[2*x : 2*x+2]), colors[y][2*x]
}

func TestDrawTheLevelTwoCharactersPerCell(t *testing.T) {
	g := newTestGame(room...)
	rows, colors, hud := drawn(g)
	for _, c := range []struct {
		x, y  int
		want  string
		color render.Color
	}{
		{0, 2, spriteFloor, colorFloor},
		{5, 3, spriteLadder, colorLadder},
		{8, 4, spriteDisk, colorDisk},
		{3, 1, spriteDoorClosed, colorDoorClosed},
		{1, 4, spriteTechnician, colorTechnician},
		{9, 4, "  ", render.ColorDefault},
	} {
		if got, color := sprite(rows, colors, c.x, c.y); got != c.want || color != c.color {
			t.Fatalf("cell (%d,%d) = %q in %v, want %q in %v\n%s", c.x, c.y, got, color, c.want, c.color, strings.Join(rows, "\n"))
		}
	}
	full := NewGame(Config{Levels: Levels, Rand: &scriptedRand{}})
	full.wrenches = 2
	_, _, hud = drawn(full)
	for _, want := range []string{"Level 1", "Score: 0", "Disks: 6", "Wrenches: 2", "♥♥♥"} {
		if !strings.Contains(hud, want) {
			t.Fatalf("HUD %q missing %q", hud, want)
		}
	}
}

func TestDrawTheOpenDoorCablesAndBots(t *testing.T) {
	g := newTestGame(room...)
	g.disks = nil
	g.cables = []Cable{{Pos: pt(2, 1), Dir: 1}}
	g.bots = []Bot{{Pos: pt(7, 4), Dir: 1}}
	g.tools = map[grid.Point]bool{pt(4, 4): true}
	g.thrown = []Wrench{{Pos: pt(6, 1), Dir: 1}}
	rows, colors, _ := drawn(g)
	for _, c := range []struct {
		x, y  int
		want  string
		color render.Color
	}{
		{3, 1, spriteDoorOpen, colorDoorOpen},
		{2, 1, spriteCable, colorCable},
		{7, 4, spriteBot, colorBot},
		{4, 4, spriteToolbox, colorToolbox},
		{6, 1, spriteWrench, colorWrench},
	} {
		if got, color := sprite(rows, colors, c.x, c.y); got != c.want || color != c.color {
			t.Fatalf("cell (%d,%d) = %q in %v, want %q in %v", c.x, c.y, got, color, c.want, c.color)
		}
	}
}
