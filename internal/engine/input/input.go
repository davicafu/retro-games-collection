package input

type Action string

const (
	ActionUp     Action = "up"
	ActionDown   Action = "down"
	ActionLeft   Action = "left"
	ActionRight  Action = "right"
	ActionQuit   Action = "quit"
	ActionSelect Action = "select"
	ActionPause  Action = "pause"
	ActionUndo   Action = "undo"
	ActionHelp   Action = "help"
	ActionFire   Action = "fire"
	ActionGuides Action = "guides"
	// ActionBackspace deletes the last typed character.
	ActionBackspace Action = "backspace"
)

// FrameInput is the input of one frame: the actions pressed and the
// characters typed, in order. A typed character that stands for an action
// (see RuneAction) arrives both ways.
type FrameInput struct {
	Actions []Action
	Text    []rune
}

// RuneAction is the action a typed character stands for, or "" when it
// stands for none: WASD move, space selects, Q quits, P pauses, U undoes,
// ? and H ask for help, F fires and G toggles guides, in either case.
func RuneAction(r rune) Action {
	switch r {
	case 'w', 'W':
		return ActionUp
	case 'a', 'A':
		return ActionLeft
	case 's', 'S':
		return ActionDown
	case 'd', 'D':
		return ActionRight
	case ' ':
		return ActionSelect
	case 'q', 'Q':
		return ActionQuit
	case 'p', 'P':
		return ActionPause
	case 'u', 'U':
		return ActionUndo
	case '?', 'h', 'H':
		return ActionHelp
	case 'f', 'F':
		return ActionFire
	case 'g', 'G':
		return ActionGuides
	}
	return ""
}

// WithoutTyping returns the frame without the actions its typed characters
// stand for, one per character, for a game that wants the text alone.
func WithoutTyping(f FrameInput) FrameInput {
	actions := append([]Action(nil), f.Actions...)
	for _, r := range f.Text {
		a := RuneAction(r)
		if a == "" {
			continue
		}
		for i, b := range actions {
			if b == a {
				actions = append(actions[:i], actions[i+1:]...)
				break
			}
		}
	}
	return FrameInput{Actions: actions, Text: append([]rune(nil), f.Text...)}
}

type Provider interface {
	Poll() FrameInput
}

type staticProvider struct{ frame FrameInput }

func Static(actions []Action) Provider {
	return staticProvider{frame: FrameInput{Actions: append([]Action(nil), actions...)}}
}

func (s staticProvider) Poll() FrameInput {
	return FrameInput{Actions: append([]Action(nil), s.frame.Actions...)}
}
