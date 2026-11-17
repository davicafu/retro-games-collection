package rack

import (
	"fmt"
	"strings"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
)

// Every cell is drawn 2 characters wide so the level looks square.
const (
	spriteFloor      = "══"
	spriteLadder     = "╟╢"
	spriteDisk       = "◘◘"
	spriteDoorClosed = "▐▌"
	spriteDoorOpen   = "▯▯"
	spriteSpool      = "◎◎"
	spriteCable      = "≈≈"
	spriteBot        = "◄►"
	spriteTechnician = "▟▙"
	spriteToolbox    = "▤▤"
	spriteWrench     = "××"

	glyphLife = '♥'
)

// Colours: what can hurt the technician is red or magenta, what it is
// after is yellow, and the door turns green once it opens.
const (
	colorFloor      = render.ColorBlue
	colorLadder     = render.ColorWhite
	colorDisk       = render.ColorYellow
	colorDoorClosed = render.ColorWhite
	colorDoorOpen   = render.ColorGreen
	colorSpool      = render.ColorRed
	colorCable      = render.ColorRed
	colorBot        = render.ColorMagenta
	colorTechnician = render.ColorCyan
	colorToolbox    = render.ColorGreen
	colorWrench     = render.ColorCyan
)

// Draw renders the HUD, the level, the cables, the bots, the wrenches in
// flight and the technician.
func Draw(screen render.Screen, g *Game) {
	f := render.CenterFrame(screen, 2*g.Width(), g.Height())
	f.Draw(screen, fmt.Sprintf("Level %d   Score: %d   Disks: %d   Wrenches: %d   %s", g.Level(), g.Score(), g.DisksLeft(), g.WrenchesLeft(), strings.Repeat(string(glyphLife), g.Lives())))
	set := func(p grid.Point, sprite string, color render.Color) {
		x, y := f.Cell(2*p.X, p.Y)
		render.WriteColoredText(screen, x, y, sprite, color)
	}
	for y := 0; y < g.Height(); y++ {
		for x := 0; x < g.Width(); x++ {
			p := grid.Point{X: x, Y: y}
			switch g.Cell(p) {
			case CellFloor:
				set(p, spriteFloor, colorFloor)
			case CellLadder:
				set(p, spriteLadder, colorLadder)
			case CellDisk:
				set(p, spriteDisk, colorDisk)
			case CellToolbox:
				set(p, spriteToolbox, colorToolbox)
			case CellSpool:
				set(p, spriteSpool, colorSpool)
			case CellExit:
				if g.DoorOpen() {
					set(p, spriteDoorOpen, colorDoorOpen)
				} else {
					set(p, spriteDoorClosed, colorDoorClosed)
				}
			}
		}
	}
	for _, c := range g.Cables() {
		set(c.Pos, spriteCable, colorCable)
	}
	for _, b := range g.Bots() {
		set(b.Pos, spriteBot, colorBot)
	}
	for _, w := range g.Thrown() {
		set(w.Pos, spriteWrench, colorWrench)
	}
	set(g.Player().Pos, spriteTechnician, colorTechnician)
}
