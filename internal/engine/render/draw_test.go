package render

import "testing"

func TestWriteTextPlacesOneRunePerCell(t *testing.T) {
	s := NewBufferScreen(12, 1)

	WriteText(s, 1, 0, "⏱ ↑↓ ok")

	if got, want := s.Snapshot(), " ⏱ ↑↓ ok    "; got != want {
		t.Fatalf("row = %q, want %q", got, want)
	}
}

func TestPadRightCountsRunes(t *testing.T) {
	cases := []struct{ in, want string }{
		{"↑ up", "↑ up  "}, // pad to 6 cells
		{"ab", "ab    "},   // ASCII unchanged
		{"←→←→←→←→", "←→←→←→"}, // truncate to 6 cells
		{"abcdef", "abcdef"}, // exact width
	}
	for _, c := range cases {
		if got := PadRight(c.in, 6); got != c.want {
			t.Fatalf("PadRight(%q, 6) = %q, want %q", c.in, got, c.want)
		}
	}
}
