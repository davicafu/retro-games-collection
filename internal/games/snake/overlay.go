package snake

type OverlayAction string

const (
	OverlayActionContinue   OverlayAction = "continue"
	OverlayActionRestart    OverlayAction = "restart"
	OverlayActionBackToMenu OverlayAction = "back_to_menu"
	OverlayActionQuit       OverlayAction = "quit"
)

type OverlayOption struct {
	Label  string
	Action OverlayAction
}

type OverlayModel struct {
	title   string
	options []OverlayOption
	index   int
}

func NewOverlayModel(title string, options []OverlayOption) OverlayModel {
	return OverlayModel{title: title, options: append([]OverlayOption(nil), options...)}
}

func (o *OverlayModel) MoveDown() {
	if o.index < len(o.options)-1 {
		o.index++
	}
}

func (o *OverlayModel) MoveUp() {
	if o.index > 0 {
		o.index--
	}
}

func (o OverlayModel) Resolve() OverlayAction {
	if len(o.options) == 0 {
		return ""
	}
	return o.options[o.index].Action
}

func (o OverlayModel) SelectedIndex() int { return o.index }
func (o OverlayModel) Title() string      { return o.title }
func (o OverlayModel) Options() []OverlayOption {
	return append([]OverlayOption(nil), o.options...)
}
