package shell

import "github.com/davidbarreiro/cli-games-collection/internal/engine/input"

// ChannelInput implements input.Provider backed by a buffered channel.
// Send is non-blocking — fast key presses are dropped if the channel is full.
type ChannelInput struct {
	ch chan key
}

// key is one key press: the action it stands for, the character it types,
// or both.
type key struct {
	action input.Action
	r      rune
}

// NewChannelInput creates a ChannelInput with the given buffer size.
func NewChannelInput(buffer int) *ChannelInput {
	return &ChannelInput{ch: make(chan key, buffer)}
}

// Send queues an action. Drops if the channel is full (non-blocking).
func (c *ChannelInput) Send(action input.Action) { c.send(key{action: action}) }

// Type queues a typed character with the action it stands for, if any.
func (c *ChannelInput) Type(r rune) { c.send(key{action: input.RuneAction(r), r: r}) }

func (c *ChannelInput) send(k key) {
	select {
	case c.ch <- k:
	default:
	}
}

// Poll drains all queued keys into a single FrameInput.
func (c *ChannelInput) Poll() input.FrameInput {
	var frame input.FrameInput
	for {
		select {
		case k := <-c.ch:
			if k.action != "" {
				frame.Actions = append(frame.Actions, k.action)
			}
			if k.r != 0 {
				frame.Text = append(frame.Text, k.r)
			}
		default:
			return frame
		}
	}
}
