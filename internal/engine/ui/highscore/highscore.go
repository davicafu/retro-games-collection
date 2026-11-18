// Package highscore is the high score UI shared by every game: the prompt
// that asks for a name when a result enters its table, and the table
// itself, as overlay content.
package highscore

import (
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/records"
	"github.com/davidbarreiro/cli-games-collection/internal/engine/ui/overlay"
)

// Board is a game's high score table (records.Board).
type Board interface {
	Top() []records.Entry
	Qualifies(records.Entry) bool
	Add(records.Entry) (int, error)
}

// Title titles the overlay that asks for a name.
const Title = "NEW HIGH SCORE"

// Prompt asks for the name of a result that enters a table.
type Prompt struct {
	entry records.Entry
	name  []rune
}

// NewPrompt asks for e's name.
func NewPrompt(e records.Entry) *Prompt { return &Prompt{entry: e} }

// Update edits the name with the frame's typed text and backspace, and
// reports done on enter or esc. Typed characters are only text here: q
// does not quit and space does not confirm.
func (p *Prompt) Update(frame input.FrameInput) (done bool) {
	frame = input.WithoutTyping(frame)
	for _, r := range frame.Text {
		if len(p.name) < records.MaxName && r >= ' ' {
			p.name = append(p.name, r)
		}
	}
	for _, a := range frame.Actions {
		switch a {
		case input.ActionBackspace:
			if len(p.name) > 0 {
				p.name = p.name[:len(p.name)-1]
			}
		case input.ActionSelect, input.ActionPause:
			return true
		}
	}
	return false
}

// Entry is the result with the name typed so far.
func (p *Prompt) Entry() records.Entry {
	e := p.entry
	e.Name = string(p.name)
	return e
}

// Overlay shows the result and the name being typed.
func (p *Prompt) Overlay() overlay.Model {
	return overlay.NewInfo(Title, []overlay.Section{
		overlay.Text(fmt.Sprintf("Score: %d", p.entry.Score)),
		overlay.Text("Name: "+string(p.name)+"_", "[enter] save"),
	})
}

// Lines are a table, best first, with the result at rank (from 1; 0 for
// none) marked, and a note when it could not be saved.
func Lines(top []records.Entry, rank int, saveErr error) []string {
	lines := []string{"HIGH SCORES"}
	for i, e := range top {
		line := fmt.Sprintf("%d. %-*s %8d", i+1, records.MaxName, e.Name, e.Score)
		if e.Seconds > 0 {
			line += fmt.Sprintf(" %02d:%02d", e.Seconds/60, e.Seconds%60)
		}
		if i+1 == rank {
			line += " ◂"
		}
		lines = append(lines, line)
	}
	if len(top) == 0 {
		lines = append(lines, "none yet")
	}
	if saveErr != nil {
		lines = append(lines, "(could not save the table)")
	}
	return lines
}
