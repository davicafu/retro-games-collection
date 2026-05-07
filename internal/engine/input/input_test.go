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
