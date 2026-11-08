package breach

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

var testLayout = newLayout(80, 24, testWidth, testHeight)

func cellAt(rows [][]rune, col, row int) rune {
	p := testLayout.toScreen(col, row)
	return rows[p.Y][p.X]
}

func hud(rows [][]rune) string { return string(rows[testLayout.originY]) }

func TestDrawHUDShowsScoreLivesAndLaunchHint(t *testing.T) {
	// Real playfield width: the test one is too narrow for the full HUD.
	g := NewGame(Config{Width: 40, Height: testHeight, Bricks: []Brick{farBrick}})
	h := hud(drawRows(g))
	if !strings.Contains(h, "Score: 0") || !strings.Contains(h, strings.Repeat(string(glyphLife), StartLives)) || !strings.Contains(h, launchHint) {
		t.Fatalf("HUD %q, want score, %d lives and launch hint", h, StartLives)
	}

	g.Step(Input{Launch: true})
	g.wideTicks = 10
	h = hud(drawRows(g))
	if strings.Contains(h, launchHint) || !strings.Contains(h, "WIDE") {
		t.Fatalf("HUD in play %q, want no launch hint and the WIDE marker", h)
	}
}

func TestDrawVirusAndInjector(t *testing.T) {
	g := newTestGame(farBrick)
	rows := drawRows(g)
	col, row := g.BallCell()
	if got := cellAt(rows, col, row); got != glyphVirus {
		t.Fatalf("virus cell = %q, want %q", got, glyphVirus)
	}
	x, w := g.Paddle()
	for c := x; c < x+w; c++ {
		if got := cellAt(rows, c, paddleRow); got != glyphInjector {
			t.Fatalf("injector column %d = %q, want %q", c, got, glyphInjector)
		}
	}
	if got := cellAt(rows, x+w, paddleRow); got != ' ' {
		t.Fatalf("cell right of the injector = %q, want empty", got)
	}
}

func TestDrawBlocksShowTheirStrength(t *testing.T) {
	rows := drawRows(newTestGame(Brick{X: 4, Y: 2, HP: 2}, Brick{X: 8, Y: 2, HP: 1}))
	got := string([]rune{cellAt(rows, 4, 2), cellAt(rows, 5, 2), cellAt(rows, 6, 2), cellAt(rows, 7, 2)})
	if got != "▐██▌" {
		t.Fatalf("hardened block = %q, want %q", got, "▐██▌")
	}
	got = string([]rune{cellAt(rows, 8, 2), cellAt(rows, 9, 2), cellAt(rows, 10, 2), cellAt(rows, 11, 2)})
	if got != "▐▓▓▌" {
		t.Fatalf("damaged block = %q, want %q", got, "▐▓▓▌")
	}
}

func TestDrawDrops(t *testing.T) {
	g := newTestGame(farBrick)
	g.drops = []Drop{{Col: 3, Y: 550, Kind: PowerUpWide}, {Col: 12, Y: 750, Kind: PowerUpLife}}
	rows := drawRows(g)
	if got := cellAt(rows, 3, 5); got != glyphWide {
		t.Fatalf("wide drop = %q, want %q", got, glyphWide)
	}
	if got := cellAt(rows, 12, 7); got != glyphLife {
		t.Fatalf("life drop = %q, want %q", got, glyphLife)
	}
}

func TestLayoutCentersTheFrame(t *testing.T) {
	l := newLayout(80, 24, 40, 16)
	if l.originX != (80-42)/2 || l.originY != (24-19)/2 {
		t.Fatalf("layout = %+v, want centred", l)
	}
	if small := newLayout(10, 5, 40, 16); small.originX != 0 || small.originY != 0 {
		t.Fatalf("layout on a small screen = %+v, want clamped to 0", small)
	}
}
