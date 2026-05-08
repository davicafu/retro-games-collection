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
