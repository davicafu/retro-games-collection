package runner

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

// testLayout is the layout Draw uses on the 80x24 test screen for a 40-wide playfield.
var testLayout = newLayout(80, 24, 40)

// cellAt returns the rune drawn for playfield column x at height h above the ground.
func cellAt(rows [][]rune, x, h int) rune {
	p := testLayout.toScreen(x, h)
	return rows[p.Y][p.X]
}

func TestDrawShowsScoreAndSpeedMultiplier(t *testing.T) {
	g := newTestGame()
	for g.Score() < 20 {
		g.Step(ActionNone)
	}

	hud := string(drawRows(g)[testLayout.originY])

	if !strings.Contains(hud, "Score: 20") {
		t.Fatalf("HUD %q missing score", hud)
	}
	if !strings.Contains(hud, "Speed x1.1") { // 40 / 35 = 1.14
		t.Fatalf("HUD %q missing speed multiplier", hud)
	}
}

func TestDrawStandingPlayerUsesTwoCells(t *testing.T) {
	rows := drawRows(newTestGame())

	if got := cellAt(rows, testPlayerX, 0); got != glyphBody {
		t.Fatalf("feet cell = %q, want %q", got, glyphBody)
	}
	if got := cellAt(rows, testPlayerX, 1); got != glyphHead {
		t.Fatalf("head cell = %q, want %q", got, glyphHead)
	}
}

func TestDrawDuckingPlayerUsesOneCell(t *testing.T) {
	g := newTestGame()
	g.Step(ActionDuck)
	rows := drawRows(g)

	if got := cellAt(rows, testPlayerX, 0); got != glyphDuck {
		t.Fatalf("duck cell = %q, want %q", got, glyphDuck)
	}
	if got := cellAt(rows, testPlayerX, 1); got != ' ' {
		t.Fatalf("cell above ducking player = %q, want empty", got)
	}
}

func TestDrawJumpingPlayerFollowsHeight(t *testing.T) {
	g := newTestGame()
	g.Step(ActionJump)
	g.Step(ActionNone) // height 2
	rows := drawRows(g)

	if got := cellAt(rows, testPlayerX, 2); got != glyphBody {
		t.Fatalf("body at height 2 = %q, want %q", got, glyphBody)
	}
	if got := cellAt(rows, testPlayerX, 3); got != glyphHead {
		t.Fatalf("head at height 3 = %q, want %q", got, glyphHead)
	}
	if got := cellAt(rows, testPlayerX, 0); got != ' ' {
		t.Fatalf("ground cell under jumping player = %q, want empty", got)
	}
}

func TestDrawObstaclesByKind(t *testing.T) {
	g := newTestGame(
		Obstacle{X: 10, Width: 1, Tall: 2, Kind: KindGround},
		Obstacle{X: 20, Width: 2, Kind: KindFlying},
	)
	rows := drawRows(g)

	for h := 0; h < 2; h++ {
		if got := cellAt(rows, 10, h); got != glyphGround {
			t.Fatalf("ground obstacle at height %d = %q, want %q", h, got, glyphGround)
		}
	}
	for x := 20; x < 22; x++ {
		if got := cellAt(rows, x, 1); got != glyphFlying {
			t.Fatalf("flying obstacle at x=%d = %q, want %q", x, got, glyphFlying)
		}
		if got := cellAt(rows, x, 0); got != ' ' {
			t.Fatalf("under flying obstacle at x=%d = %q, want empty", x, got)
		}
	}
}

func TestDrawClipsObstaclesOutsideThePlayfield(t *testing.T) {
	g := newTestGame(Obstacle{X: 39, Width: 3, Kind: KindFlying}) // columns 39..41, width 40
	rows := drawRows(g)

	if got := cellAt(rows, 39, 1); got != glyphFlying {
		t.Fatalf("visible part = %q, want %q", got, glyphFlying)
	}
	if got := cellAt(rows, 40, 1); got != '│' {
		t.Fatalf("right border = %q, want it untouched", got)
	}
}

func TestDrawGroundLine(t *testing.T) {
	rows := drawRows(newTestGame())
	ground := testLayout.toScreen(0, -1)

	if got := string(rows[ground.Y][ground.X : ground.X+40]); got != strings.Repeat(string(glyphGroundLine), 40) {
		t.Fatalf("ground row = %q", got)
	}
}

func TestDrawCentersPlayfieldSoOverlaysSitOnIt(t *testing.T) {
	// 80x24 screen, 40-wide playfield: box is 42x8, plus the HUD row above it.
	l := newLayout(80, 24, 40)
	if l.originX != 19 || l.originY != 7 {
		t.Fatalf("origin = (%d,%d), want (19,7)", l.originX, l.originY)
	}
}

func TestDrawPinsPlayfieldToTopLeftOnSmallScreens(t *testing.T) {
	l := newLayout(30, 5, 40)
	if l.originX != 0 || l.originY != 0 {
		t.Fatalf("origin = (%d,%d), want (0,0) when the screen is smaller than the playfield", l.originX, l.originY)
	}
}
