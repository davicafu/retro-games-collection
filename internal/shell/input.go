package shell

import "github.com/davidbarreiro/cli-games-collection/internal/engine/input"

// ChannelInput implements input.Provider backed by a buffered channel.
// Send is non-blocking — fast key presses are dropped if the channel is full.
type ChannelInput struct {
	ch chan input.Action
}

// NewChannelInput creates a ChannelInput with the given buffer size.
func NewChannelInput(buffer int) *ChannelInput {
	return &ChannelInput{ch: make(chan input.Action, buffer)}
}

// Send queues an action. Drops if the channel is full (non-blocking).
func (c *ChannelInput) Send(action input.Action) {
	select {
	case c.ch <- action:
	default:
	}
}

// Poll drains all queued actions into a single FrameInput.
func (c *ChannelInput) Poll() input.FrameInput {
	var actions []input.Action
	for {
		select {
		case action := <-c.ch:
			actions = append(actions, action)
		default:
			return input.FrameInput{Actions: actions}
		}
	}
}
