package overlay

import (
	"github.com/davidbarreiro/cli-games-collection/internal/engine/render"
	"github.com/mattn/go-runewidth"
)

const (
	overlayMinTextWidth       = 18
	overlayPaddingX           = 1
	overlayGapAfterTitle      = 1
	overlayGapBetweenSections = 1
	overlayGapBeforeOptions   = 1
)

func Draw(screen render.Screen, model Model) {
	if screen == nil {
		return
	}

	contentLines := flattenSections(model.Sections())
	optionLines := buildOptionLines(model)
	textWidth := maxTextWidth(model.Title(), contentLines, optionLines)
	if textWidth < overlayMinTextWidth {
		textWidth = overlayMinTextWidth
	}

	interiorWidth := textWidth + (overlayPaddingX * 2)
	interiorHeight := 1 + overlayGapAfterTitle + len(contentLines)
	if len(optionLines) > 0 {
		interiorHeight += overlayGapBeforeOptions + len(optionLines)
	}

	sw, sh := screen.Size()
	if sw < 2 || sh < 2 {
		return
	}
	if maxWidth := sw - 2; interiorWidth > maxWidth {
		interiorWidth = maxWidth
	}
	if maxHeight := sh - 2; interiorHeight > maxHeight {
		interiorHeight = maxHeight
	}
	if interiorWidth < 1 || interiorHeight < 1 {
		return
	}

	boxX := (sw - (interiorWidth + 2)) / 2
	boxY := (sh - (interiorHeight + 2)) / 2
	if boxX < 0 {
		boxX = 0
	}
	if boxY < 0 {
		boxY = 0
	}

	render.FillRect(screen, boxX+1, boxY+1, interiorWidth, interiorHeight, ' ')
	render.DrawBox(screen, boxX, boxY, interiorWidth, interiorHeight)

	titleWidth := titleLayoutWidth(model.Title())
	titleX := boxX + 1
	if titleWidth < interiorWidth {
		titleX = boxX + 1 + (interiorWidth-titleWidth)/2
	}
	writeTitleBounded(screen, titleX, boxY+1, model.Title(), (boxX+1)+interiorWidth)

	row := boxY + 2 + overlayGapAfterTitle
	textX := boxX + 1 + overlayPaddingX
	maxTextX := (boxX + 1) + interiorWidth - overlayPaddingX

	for _, line := range contentLines {
		if row > boxY+interiorHeight {
			break
		}
		writeLineBounded(screen, textX, row, line, maxTextX)
		row++
	}

	if len(optionLines) > 0 {
		row += overlayGapBeforeOptions
		for _, line := range optionLines {
			if row > boxY+interiorHeight {
				break
			}
			writeLineBounded(screen, textX, row, line, maxTextX)
			row++
		}
	}
}

func flattenSections(sections []Section) []string {
	lines := make([]string, 0)
	wroteSection := false
	for _, section := range sections {
		if len(section.Lines) == 0 {
			continue
		}
		if wroteSection {
			for gap := 0; gap < overlayGapBetweenSections; gap++ {
				lines = append(lines, "")
			}
		}
		lines = append(lines, section.Lines...)
		wroteSection = true
	}
	return lines
}

func buildOptionLines(model Model) []string {
	options := model.Options()
	lines := make([]string, 0, len(options))
	for index, option := range options {
		marker := "  "
		if index == model.SelectedIndex() {
			marker = "▸ "
		}
		lines = append(lines, marker+option.Label)
	}
	return lines
}

func maxTextWidth(title string, contentLines, optionLines []string) int {
	width := titleLayoutWidth(title)
	for _, line := range contentLines {
		if lineWidth := runeLen(line); lineWidth > width {
			width = lineWidth
		}
	}
	for _, line := range optionLines {
		if lineWidth := runeLen(line); lineWidth > width {
			width = lineWidth
		}
	}
	return width
}

func runeLen(s string) int {
	return runewidth.StringWidth(s)
}

func titleLayoutWidth(s string) int {
	width := runeLen(s)
	for _, r := range []rune(s) {
		if runewidth.RuneWidth(r) > 1 {
			width--
		}
	}
	return width
}

func titleRuneAdvance(r rune) int {
	runeWidth := runewidth.RuneWidth(r)
	switch {
	case runeWidth <= 0:
		return 0
	case runeWidth > 1:
		return 1
	default:
		return runeWidth
	}
}

func writeLineBounded(screen render.Screen, x, y int, text string, maxX int) {
	column := x
	for _, r := range text {
		width := runewidth.RuneWidth(r)
		if width <= 0 {
			continue
		}
		if column+width > maxX {
			break
		}

		render.SetCell(screen, column, y, r)
		column += width
	}
}

func writeTitleBounded(screen render.Screen, x, y int, text string, maxX int) {
	column := x
	for _, r := range text {
		width := titleRuneAdvance(r)
		if width <= 0 {
			continue
		}
		if column+width > maxX {
			break
		}

		render.SetCell(screen, column, y, r)
		column += width
	}
}
