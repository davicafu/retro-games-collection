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

func TestDrawHUDShowsLevelScoreLivesAndWide(t *testing.T) {
	// Real playfield width: the test one is too narrow for the full HUD.
	g := NewGame(Config{Width: 40, Height: testHeight, Levels: Campaign})
	g.level = 9
	g.lives = MaxLives
	g.score = 9990
	g.wideTicks = 10
	h := hud(drawRows(g))
	for _, want := range []string{"Level 10/10", "Score: 9990", strings.Repeat(string(glyphLife), MaxLives), "WIDE"} {
		if !strings.Contains(h, want) {
			t.Fatalf("HUD %q missing %q", h, want)
		}
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
	cases := []struct {
		brick Brick
		want  string
	}{
		{Brick{X: 0, Y: 2, HP: 3}, "▐██▌"},
		{Brick{X: 4, Y: 2, HP: 2}, "▐▓▓▌"},
		{Brick{X: 8, Y: 2, HP: 1}, "▐░░▌"},
		{Brick{X: 12, Y: 2, Solid: true}, "▐╳╳▌"},
	}
	g := newTestGame()
	for _, c := range cases {
		g.bricks = append(g.bricks, c.brick)
	}
	rows := drawRows(g)
	for _, c := range cases {
		x := c.brick.X
		got := string([]rune{cellAt(rows, x, 2), cellAt(rows, x+1, 2), cellAt(rows, x+2, 2), cellAt(rows, x+3, 2)})
		if got != c.want {
			t.Fatalf("block %+v = %q, want %q", c.brick, got, c.want)
		}
	}
}

func TestDrawLevelBannerAndLaunchHintOnlyBeforeLaunch(t *testing.T) {
	g := NewGame(Config{Width: testWidth, Height: testHeight, Levels: Campaign})
	g.level = 2
	if snap := string(drawRowsFlat(g)); !strings.Contains(snap, "LEVEL 3") || !strings.Contains(snap, launchHint) {
		t.Fatalf("banner or launch hint missing before launch:\n%s", snap)
	}
	g.Step(Input{Launch: true})
	if snap := string(drawRowsFlat(g)); strings.Contains(snap, "LEVEL 3") || strings.Contains(snap, launchHint) {
		t.Fatalf("banner and hint must disappear once the virus is launched")
	}
}

func drawRowsFlat(g *Game) []rune {
	var out []rune
	for _, r := range drawRows(g) {
		out = append(out, r...)
		out = append(out, '\n')
	}
	return out
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
