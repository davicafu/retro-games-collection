package tcellscreen

import (
	"github.com/davidbarreiro/cli-games-collection/internal/engine/input"
	"github.com/gdamore/tcell/v2"
)

type InputProvider struct {
	screen pendingEventScreen
}

type pendingEventScreen interface {
	HasPendingEvent() bool
	PollEvent() tcell.Event
}

func NewInputProvider(screen tcell.Screen) *InputProvider {
	return &InputProvider{screen: screen}
}

func (p *InputProvider) Poll() input.FrameInput {
	if !p.screen.HasPendingEvent() {
		return input.FrameInput{}
	}

	evt := p.screen.PollEvent()
	keyEvent, ok := evt.(*tcell.EventKey)
	if !ok {
		return input.FrameInput{}
	}

	action, ok := mapKeyEvent(keyEvent)
	if !ok {
		return input.FrameInput{}
	}

	return input.FrameInput{Actions: []input.Action{action}}
}

func mapKeyEvent(evt *tcell.EventKey) (input.Action, bool) {
	switch evt.Key() {
	case tcell.KeyUp:
		return input.ActionUp, true
	case tcell.KeyDown:
		return input.ActionDown, true
	case tcell.KeyLeft:
		return input.ActionLeft, true
	case tcell.KeyRight:
		return input.ActionRight, true
	case tcell.KeyEscape:
		return input.ActionPause, true
	case tcell.KeyEnter:
		return input.ActionSelect, true
	}

	switch evt.Rune() {
	case 'w', 'W':
		return input.ActionUp, true
	case 's', 'S':
		return input.ActionDown, true
	case 'a', 'A':
		return input.ActionLeft, true
	case 'd', 'D':
		return input.ActionRight, true
	case 'q', 'Q':
		return input.ActionQuit, true
	case ' ':
		return input.ActionSelect, true
	case 'p', 'P':
		return input.ActionPause, true
	default:
		return "", false
	}
}
