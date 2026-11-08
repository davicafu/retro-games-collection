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

// testFrame is the frame Draw uses on the 80x24 test screen for a 40-wide playfield.
var testFrame = render.CenterFrame(render.NewBufferScreen(80, 24), 40, skyHeight)

// cellAt returns the rune drawn for playfield column x at height h above the ground.
func cellAt(rows [][]rune, x, h int) rune {
	sx, sy := testFrame.Cell(x, skyHeight-1-h)
	return rows[sy][sx]
}

func TestDrawShowsScoreAndSpeedMultiplier(t *testing.T) {
	g := newTestGame()
	for g.Score() < 40 {
		g.Step(ActionNone)
	}

	hud := string(drawRows(g)[testFrame.Y])

	if !strings.Contains(hud, "Score: 40") {
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
	gx, gy := testFrame.Cell(0, skyHeight)

	if got := string(rows[gy][gx : gx+40]); got != strings.Repeat(string(glyphGroundLine), 40) {
		t.Fatalf("ground row = %q", got)
	}
}
