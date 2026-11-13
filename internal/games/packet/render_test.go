package packet

import (
	"strings"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// drawn renders g on an 80x24 screen and returns the playfield rows (inside
// the box) and the HUD row.
func drawn(g *Game) (rows []string, hud string) {
	s := render.NewBufferScreen(80, 24)
	Draw(s, g)
	all := s.Rows()
	f := render.CenterFrame(s, g.Width(), playRows)
	for row := 0; row < playRows; row++ {
		x, y := f.Cell(0, row)
		rows = append(rows, string(all[y][x:x+g.Width()]))
	}
	return rows, string(all[f.Y])
}

func TestDrawLinksFirewallsTokensAndThePacket(t *testing.T) {
	g := NewGame(Config{Width: 12, PlayerX: 3, Link: 1,
		Firewalls: []Firewall{{Link: 2, X: 6, Length: 2}},
		Tokens:    []Token{{Link: 0, X: 9}},
	})
	rows, _ := drawn(g)
	want := []string{
		"            ",
		"         ◈  ",
		"            ",
		"──  ──  ──  ",
		"            ",
		" ━━▶        ",
		"            ",
		"──  ──  ──  ",
		"      ▓▓    ",
		"      ▓▓    ",
		"      ▓▓    ",
		"──  ──  ──  ",
		"            ",
		"            ",
		"            ",
	}
	if strings.Join(rows, "\n") != strings.Join(want, "\n") {
		t.Fatalf("playfield:\n%s\nwant:\n%s", strings.Join(rows, "\n"), strings.Join(want, "\n"))
	}
}

func TestDrawSeparatorsScrollWithTheWorld(t *testing.T) {
	g := NewGame(Config{Width: 12, PlayerX: 3})
	for g.Distance() < 1 {
		g.Step(Input{})
	}
	if rows, _ := drawn(g); rows[LinkRows] != "─  ──  ──  ─" {
		t.Fatalf("separator after one cell = %q, want the dashes one cell to the left", rows[LinkRows])
	}
}

func TestDrawDroppedPacketAndHUD(t *testing.T) {
	// Wide enough for the whole HUD, which the frame clips to its width.
	g := NewGame(Config{Width: 30, PlayerX: 3, Link: 0, Firewalls: []Firewall{{Link: 1, X: 3, Length: 1}}})
	g.taken = 1
	g.Step(Input{Switch: 1})
	rows, hud := drawn(g)
	if !g.Dropped() || !strings.Contains(rows[linkRow(1)+1], string(glyphDropped)) {
		t.Fatalf("a dropped packet must show %q:\n%s", glyphDropped, strings.Join(rows, "\n"))
	}
	if !strings.Contains(hud, "Score: 50") || !strings.Contains(hud, "Speed x1.0") {
		t.Fatalf("HUD %q missing score or speed", hud)
	}
}
