package drone

import (
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

func drawRows(g *Game) [][]rune {
	screen := render.NewBufferScreen(80, 24)
	Draw(screen, g)
	return screen.Rows()
}

// testLayout is the layout Draw uses on the 80x24 test screen for the test playfield.
var testLayout = newLayout(80, 24, testWidth, testHeight)

func cellAt(rows [][]rune, x, row int) rune {
	p := testLayout.toScreen(x, row)
	return rows[p.Y][p.X]
}

func TestDrawHUDShowsTakeOffHintThenScore(t *testing.T) {
	g := newTestGame()
	if hud := string(drawRows(g)[testLayout.originY]); !strings.Contains(hud, "Score: 0") || !strings.Contains(hud, "to take off") {
		t.Fatalf("HUD before start %q, want score and take-off hint", hud)
	}

	g.Step(true)
	g.score = 7
	hud := string(drawRows(g)[testLayout.originY])
	if !strings.Contains(hud, "Score: 7") || strings.Contains(hud, "take off") {
		t.Fatalf("HUD in flight %q, want score without the hint", hud)
	}
}

func TestDrawPlacesTheDroneOnItsRow(t *testing.T) {
	g := newTestGame()
	rows := drawRows(g)
	if got := cellAt(rows, testDroneX, g.Row()); got != glyphDrone {
		t.Fatalf("drone cell = %q, want %q", got, glyphDrone)
	}
}

func TestDrawAntennaLeavesTheGapOpenWithTipsAtItsEdges(t *testing.T) {
	a := Antenna{X: 20, GapTop: 3}
	rows := drawRows(newTestGame(a))

	for x := a.X; x < a.X+AntennaWidth; x++ {
		for row := 0; row < testHeight; row++ {
			got := cellAt(rows, x, row)
			var want rune
			switch {
			case row == a.GapTop-1:
				want = glyphUpperTip
			case row == a.GapTop+GapSize:
				want = glyphLowerTip
			case a.blocks(row):
				want = glyphMast
			default:
				want = ' '
			}
			if got != want {
				t.Fatalf("antenna column %d row %d = %q, want %q", x, row, got, want)
			}
		}
	}
}

func TestDrawClipsAntennasToThePlayfield(t *testing.T) {
	rows := drawRows(newTestGame(Antenna{X: testWidth - 1, GapTop: 3}, Antenna{X: -1, GapTop: 3}))
	right := testLayout.toScreen(testWidth, 0)
	if got := rows[right.Y][right.X]; got != '│' {
		t.Fatalf("right border = %q, antenna drawn over it", got)
	}
	left := testLayout.toScreen(-1, 0)
	if got := rows[left.Y][left.X]; got != '│' {
		t.Fatalf("left border = %q, antenna drawn over it", got)
	}
}

func TestDrawGroundLineBelowTheSky(t *testing.T) {
	rows := drawRows(newTestGame())
	for x := 0; x < testWidth; x++ {
		if got := cellAt(rows, x, testHeight); got != glyphGround {
			t.Fatalf("ground column %d = %q, want %q", x, got, glyphGround)
		}
	}
}

func TestLayoutCentersTheFrame(t *testing.T) {
	l := newLayout(80, 24, 48, 14)
	frameWidth, frameHeight := 48+2, 1+1+14+1
	if l.originX != (80-frameWidth)/2 || l.originY != (24-frameHeight)/2 {
		t.Fatalf("layout = %+v, want centered", l)
	}
	if small := newLayout(10, 5, 48, 14); small.originX != 0 || small.originY != 0 {
		t.Fatalf("layout on a small screen = %+v, want clamped to 0", small)
	}
}
