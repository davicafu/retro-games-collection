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
)

type FrameInput struct {
	Actions []Action
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
