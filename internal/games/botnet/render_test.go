package botnet

import (
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// testFrame is the frame Draw uses on the 80x24 test screen for the test playfield.
var testFrame = render.CenterFrame(render.NewBufferScreen(80, 24), testWidth, testHeight)

func drawRows(g *Game) [][]rune {
	s := render.NewBufferScreen(80, 24)
	Draw(s, g)
	return s.Rows()
}

func cellAt(rows [][]rune, col, row int) rune {
	x, y := testFrame.Cell(col, row)
	return rows[y][x]
}

func textAt(rows [][]rune, col, row, n int) string {
	x, y := testFrame.Cell(col, row)
	return string(rows[y][x : x+n])
}

func TestDrawHUDShowsWaveScoreAndLives(t *testing.T) {
	g := newTestGame([]Scanner{lone}, nil)
	g.wave, g.score, g.lives = 3, 120, 2
	hud := string(drawRows(g)[testFrame.Y])
	for _, want := range []string{"Wave 3", "Score: 120", strings.Repeat(string(glyphLife), 2)} {
		if !strings.Contains(hud, want) {
			t.Fatalf("HUD %q missing %q", hud, want)
		}
	}
	if strings.Contains(hud, strings.Repeat(string(glyphLife), 3)) {
		t.Fatalf("HUD %q shows more lives than left", hud)
	}
}

func TestDrawServerOnTheBottomRow(t *testing.T) {
	g := newTestGame([]Scanner{lone}, nil)
	x, w := g.Server()
	if got := textAt(drawRows(g), x, serverRow, w); got != spriteServer {
		t.Fatalf("server = %q, want %q", got, spriteServer)
	}
}

func TestDrawScannersByKindWithTwoFrames(t *testing.T) {
	scanners := []Scanner{{X: 2, Row: 2, Kind: KindHigh}, {X: 10, Row: 4, Kind: KindMid}, {X: 20, Row: 6, Kind: KindLow}}
	for frame := 0; frame < 2; frame++ {
		g := newTestGame(scanners, nil)
		g.steps = frame
		rows := drawRows(g)
		for _, s := range scanners {
			if got, want := textAt(rows, s.X, s.Row, ScannerWidth), scannerSprites[s.Kind][frame]; got != want {
				t.Fatalf("frame %d kind %d: scanner = %q, want %q", frame, s.Kind, got, want)
			}
		}
	}
	if scannerSprites[KindLow][0] == scannerSprites[KindLow][1] {
		t.Fatalf("the two animation frames must differ")
	}
}

func TestDrawPacketAndShotsInTheHalfOfTheCellTheyAreIn(t *testing.T) {
	g := newTestGame([]Scanner{lone}, nil)
	g.packet, g.packetCol, g.packetY = true, 10, 8*SubCells
	g.shots = []Shot{{Col: 20, Y: 5*SubCells + SubCells/2}, {Col: 22, Y: 5*SubCells + SubCells/2 - 1}}
	rows := drawRows(g)
	if got := cellAt(rows, 10, 8); got != glyphPacketHigh {
		t.Fatalf("packet = %q, want %q", got, glyphPacketHigh)
	}
	g.packetY += SubCells / 2
	if got := cellAt(drawRows(g), 10, 8); got != glyphPacketLow {
		t.Fatalf("packet = %q, want %q", got, glyphPacketLow)
	}
	if got := cellAt(rows, 20, 5); got != glyphShotLow {
		t.Fatalf("shot = %q, want %q", got, glyphShotLow)
	}
	if got := cellAt(rows, 22, 5); got != glyphShotHigh {
		t.Fatalf("shot = %q, want %q", got, glyphShotHigh)
	}
}

func TestDrawCoverByHP(t *testing.T) {
	g := newTestGame([]Scanner{lone}, []Proxy{{X: 5, Row: 14, HP: 2}, {X: 6, Row: 14, HP: 1}})
	rows := drawRows(g)
	if got := cellAt(rows, 5, 14); got != glyphProxyFull {
		t.Fatalf("full proxy = %q, want %q", got, glyphProxyFull)
	}
	if got := cellAt(rows, 6, 14); got != glyphProxyWorn {
		t.Fatalf("worn proxy = %q, want %q", got, glyphProxyWorn)
	}
}

func TestDrawWaveBannerOnlyAtTheStartOfAWave(t *testing.T) {
	g := newTestGame([]Scanner{lone}, nil)
	g.wave = 2
	banner := func() bool {
		for _, r := range drawRows(g) {
			if strings.Contains(string(r), "WAVE 2") {
				return true
			}
		}
		return false
	}
	if !banner() {
		t.Fatalf("a new wave must show its banner")
	}
	for i := 0; i < BannerTicks; i++ {
		g.Step(Input{})
	}
	if banner() {
		t.Fatalf("the banner must go after %d ticks", BannerTicks)
	}
}
