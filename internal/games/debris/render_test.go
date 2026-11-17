package debris

import (
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/motion"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// drawn renders g on an 80x24 screen and returns the playfield rows, their
// colours and the HUD row.
func drawn(g *Game) (rows []string, colors [][]render.Color, hud string) {
	s := render.NewBufferScreen(80, 24)
	Draw(s, g)
	all, cs := s.Rows(), s.Colors()
	f := render.CenterFrame(s, g.Cols(), g.Rows())
	for row := 0; row < g.Rows(); row++ {
		x, y := f.Cell(0, row)
		rows = append(rows, string(all[y][x:x+g.Cols()]))
		colors = append(colors, cs[y][x:x+g.Cols()])
	}
	return rows, colors, string(all[f.Y])
}

func at(rows []string, col, row int) rune { return []rune(rows[row])[col] }

func TestDrawTheDroneByHeadingAtItsCell(t *testing.T) {
	g := newTestGame(far(Small))
	g.drone.Pos = motion.Vec{X: 7*CellW + CellW/2, Y: 3*CellH + CellH - 1}
	for heading, want := range map[int]rune{0: '▲', motion.Headings / 8: '◥', motion.Headings / 4: '▶', motion.Headings / 2: '▼', motion.Headings - 1: '▲'} {
		g.drone.Heading = heading
		rows, colors, _ := drawn(g)
		if got := at(rows, 7, 3); got != want || colors[3][7] != colorDrone {
			t.Fatalf("heading %d drawn %q in colour %v, want %q in the drone's colour", heading, got, colors[3][7], want)
		}
	}
}

func TestDrawPiecesCentredAndWrappedAcrossTheEdges(t *testing.T) {
	g := newTestGame(Piece{Size: Large, Body: motion.Body{Pos: motion.Vec{X: 0, Y: 0}}})
	g.pieces = append(g.pieces, Piece{Size: Medium, Body: motion.Body{Pos: motion.Vec{X: 10 * CellW, Y: 5 * CellH}}},
		Piece{Size: Small, Body: motion.Body{Pos: motion.Vec{X: 20 * CellW, Y: 5 * CellH}}})
	rows, _, _ := drawn(g)
	// The large piece spans columns -2..1 and rows -1..0 around (0, 0).
	last, bottom := testCols-1, testRows-1
	for _, c := range []struct {
		col, row int
		want     rune
	}{{last - 1, bottom, '▟'}, {last, bottom, '█'}, {0, bottom, '█'}, {1, bottom, '▙'}, {last - 1, 0, '▜'}, {1, 0, '▛'}, {9, 5, '▓'}, {10, 5, '▓'}, {20, 5, '▪'}} {
		if got := at(rows, c.col, c.row); got != c.want {
			t.Fatalf("cell (%d,%d) = %q, want %q\n%s", c.col, c.row, got, c.want, strings.Join(rows, "\n"))
		}
	}
}

func TestDrawPiecesWhereTheyAreHit(t *testing.T) {
	// A large piece whose centre is near the bottom right corner of cell
	// (10, 5) is drawn over columns 9..12 and rows 5..6, around that centre.
	g := newTestGame(Piece{Size: Large, Body: motion.Body{Pos: motion.Vec{X: 10*CellW + CellW - 1, Y: 5*CellH + CellH - 1}}})
	rows, colors, _ := drawn(g)
	for row := 4; row <= 7; row++ {
		for col := 8; col <= 13; col++ {
			inside := row >= 5 && row <= 6 && col >= 9 && col <= 12
			if drawn := at(rows, col, row) != ' '; drawn != inside {
				t.Fatalf("cell (%d,%d) drawn %v, want %v\n%s", col, row, drawn, inside, strings.Join(rows, "\n"))
			}
			if inside && colors[row][col] != colorPiece {
				t.Fatalf("debris drawn in %v, want %v", colors[row][col], colorPiece)
			}
		}
	}
	if colorPiece == colorDrone || colorPiece == colorShot || colorPiece != render.ColorRed {
		t.Fatalf("debris must be red and apart from the drone and the shots")
	}
}

func TestDrawShotsAndHUD(t *testing.T) {
	g := newTestGame(far(Large))
	g.shots = []Shot{{Body: motion.Body{Pos: motion.Vec{X: 3 * CellW, Y: 2 * CellH}}, Left: 10}}
	g.score, g.lives, g.wave = 120, 2, 3
	rows, colors, hud := drawn(g)
	if at(rows, 3, 2) != glyphShot || colors[2][3] != colorShot {
		t.Fatalf("shot drawn %q in %v, want %q in the shots' colour", at(rows, 3, 2), colors[2][3], glyphShot)
	}
	for _, want := range []string{"Wave 3", "Score: 120", strings.Repeat(string(glyphLife), 2)} {
		if !strings.Contains(hud, want) {
			t.Fatalf("HUD %q missing %q", hud, want)
		}
	}
	if strings.Contains(hud, strings.Repeat(string(glyphLife), 3)) {
		t.Fatalf("HUD %q shows more lives than left", hud)
	}
}

func TestDrawAShieldedDroneBlinks(t *testing.T) {
	g := newTestGame(far(Large))
	col, row := testCols/2, testRows/2
	seen := map[rune]bool{}
	for g.shield = ShieldTicks; g.shield > 0; g.shield-- {
		rows, _, _ := drawn(g)
		seen[at(rows, col, row)] = true
	}
	if !seen['▲'] || !seen[' '] {
		t.Fatalf("a shielded drone must blink, saw %v", seen)
	}
	g.shield = 0
	if rows, _, _ := drawn(g); at(rows, col, row) != '▲' {
		t.Fatalf("an unshielded drone must always show")
	}
}

// TestDrawAnAimMarkAlongTheExactHeading: the drone's glyph only shows the
// nearest eighth of a turn, so a mark ahead of it shows where it really
// points and thrusts.
func TestDrawAnAimMarkAlongTheExactHeading(t *testing.T) {
	g := newTestGame(far(Small))
	g.drone.Pos = motion.Vec{X: 7*CellW + CellW/2, Y: 3*CellH + CellH - 1}
	for heading, want := range map[int][2]int{0: {7, 1}, 1: {8, 2}, motion.Headings - 2: {5, 2}} {
		g.drone.Heading = heading
		rows, colors, _ := drawn(g)
		if at(rows, 7, 3) != '▲' {
			t.Fatalf("heading %d: drone drawn %q, want ▲", heading, at(rows, 7, 3))
		}
		if got := at(rows, want[0], want[1]); got != glyphAim || colors[want[1]][want[0]] != colorDrone {
			t.Fatalf("heading %d: cell %v = %q, want the aim mark\n%s", heading, want, got, strings.Join(rows, "\n"))
		}
	}
}
