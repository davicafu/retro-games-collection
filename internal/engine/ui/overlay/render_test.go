package overlay

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/mattn/go-runewidth"
)

func TestDrawCentersTitleInsideOverlayFrame(t *testing.T) {
	screen := render.NewBufferScreen(40, 20)
	model := NewInfo("Paused", []Section{Text("Press P to resume")})

	Draw(screen, model)

	rows := screen.Rows()
	left, top := findFirstRune(rows, '┌')
	if left < 0 || top < 0 {
		t.Fatalf("expected overlay frame top-left corner to be rendered")
	}

	right, ok := findFirstRuneOnRow(rows[top], '┐')
	if !ok {
		t.Fatalf("expected overlay frame top-right corner to be rendered")
	}

	interiorWidth := right - left - 1
	titleRunes := []rune(model.Title())
	expectedX := left + 1 + (interiorWidth-runewidth.StringWidth(model.Title()))/2
	titleY := top + 1

	for i, r := range titleRunes {
		if got := rows[titleY][expectedX+i]; got != r {
			t.Fatalf("expected title rune %q at (%d,%d), got %q", r, expectedX+i, titleY, got)
		}
	}
}

func TestDrawRendersSelectedOptionMarker(t *testing.T) {
	screen := render.NewBufferScreen(40, 20)
	model := NewAction(
		"Paused",
		[]Section{Text("Choose an option")},
		[]Option{
			{ID: "resume", Label: "Resume", Enabled: true},
			{ID: "restart", Label: "Restart", Enabled: true},
			{ID: "quit", Label: "Quit", Enabled: true},
		},
	)

	model.Apply(IntentMoveDown)
	Draw(screen, model)

	frame := rowsToStrings(screen.Rows())
	full := strings.Join(frame, "\n")
	if !strings.Contains(full, "▸ Restart") {
		t.Fatalf("expected selected option marker for restart option")
	}
	if !strings.Contains(full, "  Resume") {
		t.Fatalf("expected unselected option to render without marker")
	}
	if got := countRune(full, '▸'); got != 1 {
		t.Fatalf("expected exactly one selected marker, got %d", got)
	}
}

func TestDrawCentersWideTitleAndClampsSmallScreens(t *testing.T) {
	screen := render.NewBufferScreen(24, 6)
	model := NewInfo("🚀 Go", []Section{Text("Hi")})

	Draw(screen, model)

	rows := screen.Rows()
	left, top := findFirstRune(rows, '┌')
	if top != 0 {
		t.Fatalf("expected overlay to clamp to the top edge on small screens, got top %d", top)
	}

	right, ok := findFirstRuneOnRow(rows[top], '┐')
	if !ok {
		t.Fatalf("expected overlay top-right corner on clamped screen")
	}

	interiorWidth := right - left - 1
	expectedX := left + 1 + (interiorWidth-titleLayoutWidth(model.Title()))/2
	titleY := top + 1
	if got := rows[titleY][expectedX]; got != '🚀' {
		t.Fatalf("expected wide title to start with rocket at (%d,%d), got %q", expectedX, titleY, got)
	}
	if got := rows[titleY][expectedX+2]; got != 'G' {
		t.Fatalf("expected wide title alignment to place G at (%d,%d), got %q", expectedX+2, titleY, got)
	}
}

func TestDrawCentersEmojiWrappedTitleWithoutExtraGap(t *testing.T) {
	screen := render.NewBufferScreen(36, 8)
	model := NewInfo("🎉 FIN DEL JUEGO!🎉", []Section{Text("Done")})

	Draw(screen, model)

	rows := rowsToStrings(screen.Rows())
	titleRow := findRowContaining(rows, "FIN DEL JUEGO!")
	if titleRow < 0 {
		t.Fatalf("expected to find title row for %q", model.Title())
	}
	if strings.Contains(rows[titleRow], "🎉  FIN") || strings.Contains(rows[titleRow], "! 🎉") {
		t.Fatalf("expected compact emoji title spacing, got row %q", rows[titleRow])
	}
	leftBorder := strings.Index(rows[titleRow], "│")
	rightBorder := strings.LastIndex(rows[titleRow], "│")
	content := rows[titleRow][leftBorder+len("│") : rightBorder]
	titleStart := strings.Index(content, model.Title())
	if titleStart < 0 {
		t.Fatalf("expected title content inside row %q", rows[titleRow])
	}
	leftPad := runewidth.StringWidth(content[:titleStart])
	rightPad := runewidth.StringWidth(content[titleStart+len(model.Title()):])
	if leftPad > rightPad {
		t.Fatalf("expected extra centering space not to accumulate on right side, got left=%d right=%d row=%q", leftPad, rightPad, rows[titleRow])
	}
}

func TestDrawUsesCompactTitleWidthForOverlayBounds(t *testing.T) {
	screen := render.NewBufferScreen(40, 10)
	model := NewInfo("🎉 TITULO MUY LARGO🎉", []Section{Text("Done")})

	Draw(screen, model)

	rows := screen.Rows()
	left, top := findFirstRune(rows, '┌')
	right, ok := findFirstRuneOnRow(rows[top], '┐')
	if !ok {
		t.Fatalf("expected overlay frame top-right corner")
	}

	interiorWidth := right - left - 1
	expected := titleLayoutWidth(model.Title()) + (overlayPaddingX * 2)
	if interiorWidth != expected {
		t.Fatalf("expected compact title-based interior width %d, got %d", expected, interiorWidth)
	}
}

func TestDrawSkipsZeroWidthRunesWithoutCollapsingFollowingText(t *testing.T) {
	screen := render.NewBufferScreen(24, 8)
	model := NewInfo("e\u0301X", []Section{Text("Hi")})

	Draw(screen, model)

	rows := screen.Rows()
	left, top := findFirstRune(rows, '┌')
	right, ok := findFirstRuneOnRow(rows[top], '┐')
	if !ok {
		t.Fatalf("expected overlay frame top-right corner")
	}

	interiorWidth := right - left - 1
	expectedX := left + 1 + (interiorWidth-runewidth.StringWidth(model.Title()))/2
	titleY := top + 1
	if got := rows[titleY][expectedX]; got != 'e' {
		t.Fatalf("expected title to start with e at (%d,%d), got %q", expectedX, titleY, got)
	}
	if got := rows[titleY][expectedX+1]; got != 'X' {
		t.Fatalf("expected X to remain visible after zero-width rune at (%d,%d), got %q", expectedX+1, titleY, got)
	}
}

func TestDrawSkipsEmptySectionsWithoutAddingExtraGaps(t *testing.T) {
	screen := render.NewBufferScreen(30, 12)
	model := NewInfo("Help", []Section{
		Text("Line one"),
		{},
		Text("Line two"),
	})

	Draw(screen, model)

	rows := rowsToStrings(screen.Rows())
	lineOneY := findRowContaining(rows, "Line one")
	lineTwoY := findRowContaining(rows, "Line two")
	if lineOneY < 0 || lineTwoY < 0 {
		t.Fatalf("expected both content lines to be rendered")
	}
	if got := lineTwoY - lineOneY; got != 2 {
		t.Fatalf("expected exactly one blank row between non-empty sections, got row distance %d", got)
	}
}

func TestDrawClampsOversizedOverlayToScreenBounds(t *testing.T) {
	screen := render.NewBufferScreen(10, 4)
	model := NewAction(
		"Very long title",
		[]Section{Text("This content is longer than the screen")},
		[]Option{{ID: "resume", Label: "Resume", Enabled: true}},
	)

	Draw(screen, model)

	rows := screen.Rows()
	if got := rows[0][0]; got != '┌' {
		t.Fatalf("expected clamped frame to start at top-left corner, got %q", got)
	}
	if got := rows[0][9]; got != '┐' {
		t.Fatalf("expected clamped frame to end at top-right corner, got %q", got)
	}
	if got := rows[3][0]; got != '└' {
		t.Fatalf("expected clamped frame to reach bottom-left corner, got %q", got)
	}
	if got := rows[3][9]; got != '┘' {
		t.Fatalf("expected clamped frame to reach bottom-right corner, got %q", got)
	}
}

func findFirstRune(rows [][]rune, target rune) (x int, y int) {
	for rowIndex := range rows {
		if col, ok := findFirstRuneOnRow(rows[rowIndex], target); ok {
			return col, rowIndex
		}
	}
	return -1, -1
}

func findFirstRuneOnRow(row []rune, target rune) (int, bool) {
	for col, r := range row {
		if r == target {
			return col, true
		}
	}
	return -1, false
}

func rowsToStrings(rows [][]rune) []string {
	out := make([]string, len(rows))
	for i := range rows {
		out[i] = string(rows[i])
	}
	return out
}

func countRune(s string, target rune) int {
	count := 0
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		if r == target {
			count++
		}
	}
	return count
}

func findRowContaining(rows []string, fragment string) int {
	for i, row := range rows {
		if strings.Contains(row, fragment) {
			return i
		}
	}
	return -1
}
