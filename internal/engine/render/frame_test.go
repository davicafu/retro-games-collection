package render

import (
	"strings"
	"testing"
)

func TestCenterFrameCentresHUDAndBoxOnTheScreen(t *testing.T) {
	s := NewBufferScreen(20, 9)
	f := CenterFrame(s, 6, 3) // frame: 8 wide, HUD + 5 box rows = 6 tall

	if f.X != 6 || f.Y != 1 {
		t.Fatalf("frame origin = (%d,%d), want (6,1)", f.X, f.Y)
	}
	if x, y := f.Cell(0, 0); x != 7 || y != 3 {
		t.Fatalf("cell (0,0) = (%d,%d), want (7,3) inside the box", x, y)
	}
	if x, y := f.Cell(5, 3); x != 12 || y != 6 {
		t.Fatalf("cell (5,3) = (%d,%d), want (12,6) on the bottom border", x, y)
	}
}

func TestCenterFrameNeverStartsOffScreen(t *testing.T) {
	f := CenterFrame(NewBufferScreen(4, 2), 10, 10)
	if f.X != 0 || f.Y != 0 {
		t.Fatalf("frame origin = (%d,%d), want (0,0) on a small screen", f.X, f.Y)
	}
}

func TestFrameDrawWritesThePaddedHUDAndTheBox(t *testing.T) {
	s := NewBufferScreen(10, 7)
	f := CenterFrame(s, 6, 3)
	WriteText(s, 2, 0, "stale!") // old HUD text must be overwritten

	f.Draw(s, "HUD")

	want := strings.Join([]string{
		"  HUD     ",
		" ┌──────┐ ",
		" │      │ ",
		" │      │ ",
		" │      │ ",
		" └──────┘ ",
		"          ",
	}, "\n")
	if got := s.Snapshot(); got != want {
		t.Fatalf("frame =\n%s\nwant\n%s", got, want)
	}
}
