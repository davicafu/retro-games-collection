package overlay

type Kind string

const (
	KindAction Kind = "action"
	KindInfo   Kind = "info"
)

type Intent string

const (
	IntentMoveUp   Intent = "move_up"
	IntentMoveDown Intent = "move_down"
	IntentConfirm  Intent = "confirm"
	IntentClose    Intent = "close"
)

type DismissPolicy string

const (
	DismissPolicyNone  DismissPolicy = "none"
	DismissPolicyClose DismissPolicy = "close"
)

type ResultKind string

const (
	ResultKindNone      ResultKind = "none"
	ResultKindConfirmed ResultKind = "confirmed"
	ResultKindClosed    ResultKind = "closed"
)

type Result struct {
	Kind     ResultKind
	OptionID string
}

type Option struct {
	ID      string
	Label   string
	Enabled bool
}

type Section struct {
	Kind  string
	Lines []string
}

func Text(lines ...string) Section {
	return Section{Kind: "text", Lines: append([]string(nil), lines...)}
}

func cloneSectionsDeep(sections []Section) []Section {
	cloned := make([]Section, len(sections))
	for i := range sections {
		cloned[i] = Section{
			Kind:  sections[i].Kind,
			Lines: append([]string(nil), sections[i].Lines...),
		}
	}
	return cloned
}

func firstEnabledIndex(options []Option) int {
	for i := range options {
		if options[i].Enabled {
			return i
		}
	}
	return -1
}

func nextEnabledIndex(options []Option, start int, step int) (int, bool) {
	for i := start + step; i >= 0 && i < len(options); i += step {
		if options[i].Enabled {
			return i, true
		}
	}
	return 0, false
}

type Model struct {
	kind          Kind
	title         string
	sections      []Section
	options       []Option
	selectedIndex int
	dismissPolicy DismissPolicy
}

func NewAction(title string, sections []Section, options []Option) Model {
	clonedSections := cloneSectionsDeep(sections)
	clonedOptions := make([]Option, len(options))
	copy(clonedOptions, options)

	return Model{
		kind:          KindAction,
		title:         title,
		sections:      clonedSections,
		options:       clonedOptions,
		selectedIndex: firstEnabledIndex(clonedOptions),
		dismissPolicy: DismissPolicyNone,
	}
}

func NewInfo(title string, sections []Section) Model {
	clonedSections := cloneSectionsDeep(sections)

	return Model{
		kind:          KindInfo,
		title:         title,
		sections:      clonedSections,
		options:       nil,
		selectedIndex: 0,
		dismissPolicy: DismissPolicyClose,
	}
}

func (m *Model) Apply(intent Intent) Result {
	switch intent {
	case IntentMoveUp:
		if m.kind == KindAction && len(m.options) > 0 {
			start := m.selectedIndex
			if start < 0 {
				start = len(m.options)
			}
			if idx, ok := nextEnabledIndex(m.options, start, -1); ok {
				m.selectedIndex = idx
			}
		}
	case IntentMoveDown:
		if m.kind == KindAction && len(m.options) > 0 {
			if idx, ok := nextEnabledIndex(m.options, m.selectedIndex, 1); ok {
				m.selectedIndex = idx
			}
		}
	case IntentConfirm:
		if m.kind == KindAction && m.selectedIndex >= 0 && m.selectedIndex < len(m.options) && m.options[m.selectedIndex].Enabled {
			return Result{Kind: ResultKindConfirmed, OptionID: m.options[m.selectedIndex].ID}
		}
	case IntentClose:
		if m.dismissPolicy == DismissPolicyClose {
			return Result{Kind: ResultKindClosed}
		}
	}

	return Result{Kind: ResultKindNone}
}

func (m Model) Title() string { return m.title }

func (m Model) Kind() Kind { return m.kind }

func (m Model) Sections() []Section {
	return cloneSectionsDeep(m.sections)
}

func (m Model) Options() []Option {
	return append([]Option(nil), m.options...)
}

func (m Model) SelectedIndex() int { return m.selectedIndex }

func (m Model) DismissPolicy() DismissPolicy { return m.dismissPolicy }

func (m Model) WithDismissPolicy(policy DismissPolicy) Model {
	m.dismissPolicy = policy
	return m
}
