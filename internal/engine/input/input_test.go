package input

import "testing"

func TestStaticMakesDefensiveCopies(t *testing.T) {
	original := []Action{ActionUp, ActionLeft}
	p := Static(original)

	original[0] = ActionQuit

	first := p.Poll()
	if first.Actions[0] != ActionUp {
		t.Fatalf("expected provider to keep original value, got %q", first.Actions[0])
	}

	first.Actions[0] = ActionDown
	second := p.Poll()
	if second.Actions[0] != ActionUp {
		t.Fatalf("expected poll result to be isolated copy, got %q", second.Actions[0])
	}
}

func TestStaticProviderSupportsGenericGameplayActions(t *testing.T) {
	p := Static([]Action{ActionPause, ActionHelp, ActionUndo})

	frame := p.Poll()
	if len(frame.Actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(frame.Actions))
	}
	if frame.Actions[0] != ActionPause || frame.Actions[1] != ActionHelp || frame.Actions[2] != ActionUndo {
		t.Fatalf("expected generic gameplay actions to remain intact, got %v", frame.Actions)
	}
}

func TestRuneActionIsWhatATypedCharacterStandsFor(t *testing.T) {
	cases := map[rune]Action{
		'w': ActionUp, 'W': ActionUp, 'a': ActionLeft, 'A': ActionLeft,
		's': ActionDown, 'S': ActionDown, 'd': ActionRight, 'D': ActionRight,
		'q': ActionQuit, 'Q': ActionQuit, ' ': ActionSelect,
		'p': ActionPause, 'P': ActionPause, 'u': ActionUndo, 'U': ActionUndo,
		'?': ActionHelp, 'h': ActionHelp, 'H': ActionHelp,
		'f': ActionFire, 'F': ActionFire, 'g': ActionGuides, 'G': ActionGuides,
		'x': "", '1': "", '/': "",
	}
	for r, want := range cases {
		if got := RuneAction(r); got != want {
			t.Fatalf("RuneAction(%q) = %q, want %q", r, got, want)
		}
	}
}

func TestWithoutTypingDropsOneActionPerTypedCharacter(t *testing.T) {
	frame := FrameInput{Actions: []Action{ActionQuit, ActionUp, ActionSelect, ActionQuit}, Text: []rune{'q', 'x', ' '}}
	got := WithoutTyping(frame)
	want := []Action{ActionUp, ActionQuit}
	if len(got.Actions) != len(want) || got.Actions[0] != want[0] || got.Actions[1] != want[1] {
		t.Fatalf("actions = %v, want %v: an arrow and a quit not typed stay", got.Actions, want)
	}
	if string(got.Text) != "qx " {
		t.Fatalf("text = %q, want it kept", string(got.Text))
	}
	if len(frame.Actions) != 4 {
		t.Fatalf("the original frame must not change: %v", frame.Actions)
	}
}
