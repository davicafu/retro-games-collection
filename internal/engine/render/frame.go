package render

// Frame is a playfield of Width×Height cells inside a bordered box, with a
// HUD row right above the box, centred on the screen. Overlays centre on the
// screen too, so they sit on top of the playfield.
type Frame struct {
	X, Y          int // left border column and HUD row
	Width, Height int // playfield size, borders excluded
}

// CenterFrame centres a frame for a width×height playfield on the screen,
// never starting off-screen.
func CenterFrame(screen Screen, width, height int) Frame {
	sw, sh := screen.Size()
	frameWidth := width + 2
	frameHeight := 1 + height + 2 // HUD, box with its borders
	return Frame{
		X:      max(0, (sw-frameWidth)/2),
		Y:      max(0, (sh-frameHeight)/2),
		Width:  width,
		Height: height,
	}
}

// Cell maps a playfield column and row to screen coordinates. Row Height is
// the box's bottom border.
func (f Frame) Cell(col, row int) (x, y int) {
	return f.X + 1 + col, f.Y + 2 + row
}

// Draw writes the HUD, padded to the playfield width, and the box.
func (f Frame) Draw(screen Screen, hud string) {
	WriteText(screen, f.X+1, f.Y, PadRight(hud, f.Width))
	DrawBox(screen, f.X, f.Y+1, f.Width, f.Height)
}
