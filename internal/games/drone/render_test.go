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

// testFrame is the frame Draw uses on the 80x24 test screen for the test playfield.
var testFrame = render.CenterFrame(render.NewBufferScreen(80, 24), testWidth, testHeight)

func cellAt(rows [][]rune, x, row int) rune {
	sx, sy := testFrame.Cell(x, row)
	return rows[sy][sx]
}

func TestDrawHUDShowsTakeOffHintThenScore(t *testing.T) {
	g := newTestGame()
	if hud := string(drawRows(g)[testFrame.Y]); !strings.Contains(hud, "Score: 0") || !strings.Contains(hud, "to take off") {
		t.Fatalf("HUD before start %q, want score and take-off hint", hud)
	}

	g.Step(true)
	g.score = 7
	hud := string(drawRows(g)[testFrame.Y])
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
	a := Antenna{X: 20, GapTop: 3, Gap: 5}
	rows := drawRows(newTestGame(a))

	for x := a.X; x < a.X+AntennaWidth; x++ {
		for row := 0; row < testHeight; row++ {
			got := cellAt(rows, x, row)
			var want rune
			switch {
			case row == a.GapTop-1:
				want = glyphUpperTip
			case row == a.GapTop+a.Gap:
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
	rows := drawRows(newTestGame(Antenna{X: testWidth - 1, GapTop: 3, Gap: 5}, Antenna{X: -1, GapTop: 3, Gap: 5}))
	if got := cellAt(rows, testWidth, 0); got != '│' {
		t.Fatalf("right border = %q, antenna drawn over it", got)
	}
	if got := cellAt(rows, -1, 0); got != '│' {
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
