package overlay

import "testing"

func TestActionOverlayApplyMovesSelectionAndConfirmsOptionID(t *testing.T) {
	model := NewAction(
		"Paused",
		[]Section{Text("Choose an action")},
		[]Option{
			{ID: "resume", Label: "Resume", Enabled: true},
			{ID: "restart", Label: "Restart", Enabled: true},
			{ID: "quit", Label: "Quit", Enabled: true},
		},
	)

	if got := model.Kind(); got != KindAction {
		t.Fatalf("expected kind %q, got %q", KindAction, got)
	}

	if got := model.Apply(IntentMoveDown); got.Kind != ResultKindNone {
		t.Fatalf("expected no result on move down, got %+v", got)
	}
	if got := model.SelectedIndex(); got != 1 {
		t.Fatalf("expected selected index 1, got %d", got)
	}

	model.Apply(IntentMoveDown)
	model.Apply(IntentMoveDown)
	if got := model.SelectedIndex(); got != 2 {
		t.Fatalf("expected selected index to stay at last option (2), got %d", got)
	}

	model.Apply(IntentMoveUp)
	if got := model.SelectedIndex(); got != 1 {
		t.Fatalf("expected selected index 1 after move up, got %d", got)
	}

	if got := model.Apply(IntentConfirm); got != (Result{Kind: ResultKindConfirmed, OptionID: "restart"}) {
		t.Fatalf("expected confirm result with option id restart, got %+v", got)
	}
}

func TestInfoOverlayApplyClosesWithoutSelection(t *testing.T) {
	model := NewInfo(
		"Help",
		[]Section{Text("Press arrows to move.")},
	)

	if got := model.Kind(); got != KindInfo {
		t.Fatalf("expected kind %q, got %q", KindInfo, got)
	}

	if got := len(model.Options()); got != 0 {
		t.Fatalf("expected no options in info overlay, got %d", got)
	}

	if got := model.Apply(IntentConfirm); got.Kind != ResultKindNone {
		t.Fatalf("expected no result on confirm for info overlay, got %+v", got)
	}

	if got := model.Apply(IntentClose); got != (Result{Kind: ResultKindClosed}) {
		t.Fatalf("expected close result without selection, got %+v", got)
	}
}

func TestNewActionPreservesDisabledOptionsAndSkipsThemInApply(t *testing.T) {
	model := NewAction(
		"Paused",
		nil,
		[]Option{
			{ID: "resume", Label: "Resume", Enabled: true},
			{ID: "restart", Label: "Restart", Enabled: false},
			{ID: "quit", Label: "Quit", Enabled: true},
		},
	)

	options := model.Options()
	if options[1].Enabled {
		t.Fatalf("expected disabled option to stay disabled")
	}

	if got := model.Apply(IntentMoveDown); got.Kind != ResultKindNone {
		t.Fatalf("expected move down to return none, got %+v", got)
	}
	if got := model.SelectedIndex(); got != 2 {
		t.Fatalf("expected selection to skip disabled option and land at index 2, got %d", got)
	}

	if got := model.Apply(IntentConfirm); got != (Result{Kind: ResultKindConfirmed, OptionID: "quit"}) {
		t.Fatalf("expected confirm on enabled option, got %+v", got)
	}
}

func TestActionOverlayConfirmDoesNotConfirmDisabledOrMissingSelection(t *testing.T) {
	model := NewAction(
		"Paused",
		nil,
		[]Option{
			{ID: "resume", Label: "Resume", Enabled: false},
			{ID: "restart", Label: "Restart", Enabled: true},
		},
	)

	if got := model.Apply(IntentConfirm); got != (Result{Kind: ResultKindConfirmed, OptionID: "restart"}) {
		t.Fatalf("expected first confirm to target first enabled option, got %+v", got)
	}

	empty := NewAction("Paused", nil, nil)
	if got := empty.Apply(IntentConfirm); got.Kind != ResultKindNone {
		t.Fatalf("expected empty overlay confirm to be safe no-op, got %+v", got)
	}
}

func TestActionOverlayCloseBehaviorCanBeConfigured(t *testing.T) {
	base := NewAction(
		"Paused",
		nil,
		[]Option{{ID: "resume", Label: "Resume", Enabled: true}},
	)

	if got := base.Apply(IntentClose); got.Kind != ResultKindNone {
		t.Fatalf("expected default action overlay close to be ignored, got %+v", got)
	}

	closable := base.WithDismissPolicy(DismissPolicyClose)
	if got := closable.Apply(IntentClose); got != (Result{Kind: ResultKindClosed}) {
		t.Fatalf("expected configured action overlay close to return closed, got %+v", got)
	}

	if got := base.Apply(IntentClose); got.Kind != ResultKindNone {
		t.Fatalf("expected builder-style API to preserve original model, got %+v", got)
	}
}

func TestSectionsAreDeepCopiedOnConstructionAndAccess(t *testing.T) {
	sections := []Section{{Kind: "text", Lines: []string{"line 1", "line 2"}}}
	model := NewInfo("Help", sections)

	sections[0].Lines[0] = "mutated input"

	got := model.Sections()
	if got[0].Lines[0] != "line 1" {
		t.Fatalf("expected model to keep original section lines, got %q", got[0].Lines[0])
	}

	got[0].Lines[1] = "mutated output"
	nextRead := model.Sections()
	if nextRead[0].Lines[1] != "line 2" {
		t.Fatalf("expected accessor deep-copy isolation, got %q", nextRead[0].Lines[1])
	}
}
