package packet

import (
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

const (
	// LinkRows is how many rows each link is drawn on; the packet and the
	// tokens sit on the middle one.
	LinkRows = 3

	glyphFirewall  = '▓'
	glyphToken     = '◈'
	glyphPacket    = '▶'
	glyphTrail     = '━'
	glyphDropped   = '×'
	glyphSeparator = '─'

	// separatorDash and separatorPeriod shape the dashed separators between
	// links: dashes of separatorDash cells every separatorPeriod cells,
	// scrolling with the world.
	separatorDash   = 2
	separatorPeriod = 4
)

// playRows is the playfield height: the links and the separators between them.
const playRows = Links*LinkRows + Links - 1

// linkRow is the playfield row of a link's first row.
func linkRow(link int) int { return link * (LinkRows + 1) }

// Draw renders the HUD, the links, firewalls, tokens and the packet.
func Draw(screen render.Screen, g *Game) {
	f := render.CenterFrame(screen, g.Width(), playRows)
	f.Draw(screen, fmt.Sprintf("Score: %d   Speed x%d.%d", g.Score(), g.Speed()*10/BaseSpeed/10, g.Speed()*10/BaseSpeed%10))
	set := func(x, row int, glyph rune) {
		if x >= 0 && x < g.Width() {
			sx, sy := f.Cell(x, row)
			render.SetCell(screen, sx, sy, glyph)
		}
	}
	for link := 1; link < Links; link++ {
		for x := 0; x < g.Width(); x++ {
			if (x+g.Distance())%separatorPeriod < separatorDash {
				set(x, linkRow(link)-1, glyphSeparator)
			}
		}
	}
	for _, w := range g.Firewalls() {
		for x := w.X; x < w.X+w.Length; x++ {
			for r := 0; r < LinkRows; r++ {
				set(x, linkRow(w.Link)+r, glyphFirewall)
			}
		}
	}
	for _, t := range g.Tokens() {
		set(t.X, linkRow(t.Link)+LinkRows/2, glyphToken)
	}
	row := linkRow(g.Link()) + LinkRows/2
	if g.Dropped() {
		set(g.PlayerX(), row, glyphDropped)
		return
	}
	set(g.PlayerX()-2, row, glyphTrail)
	set(g.PlayerX()-1, row, glyphTrail)
	set(g.PlayerX(), row, glyphPacket)
}
