package snake

import "testing"

func TestOverlayModel_MoveSelectionAndResolveAction(t *testing.T) {
	overlay := NewOverlayModel("Game Over", []OverlayOption{
		{Label: "Play Again", Action: OverlayActionRestart},
		{Label: "Back to Menu", Action: OverlayActionBackToMenu},
		{Label: "Quit", Action: OverlayActionQuit},
	})

	overlay.MoveDown()
	if overlay.SelectedIndex() != 1 {
		t.Fatalf("expected selected index 1, got %d", overlay.SelectedIndex())
	}

	if got := overlay.Resolve(); got != OverlayActionBackToMenu {
		t.Fatalf("expected back-to-menu action, got %q", got)
	}
}
