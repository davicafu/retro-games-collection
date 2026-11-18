package records

import (
	"os"
	"path/filepath"
	"testing"
)

type run struct {
	Level int `json:"level"`
	Score int `json:"score"`
}

func openSaves(t *testing.T) (*Saves, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "saves.json")
	s, err := OpenSaves(path)
	if err != nil {
		t.Fatal(err)
	}
	return s, path
}

func TestASlotStoresLoadsAndClearsARun(t *testing.T) {
	s, path := openSaves(t)
	slot := s.Slot("hack")
	var got run
	if ok, err := slot.Load(&got); ok || err != nil || slot.Exists() {
		t.Fatalf("an empty slot loads nothing")
	}
	if err := slot.Store(run{Level: 3, Score: 1800}); err != nil {
		t.Fatal(err)
	}
	again, _ := OpenSaves(path)
	if ok, err := again.Slot("hack").Load(&got); !ok || err != nil || got != (run{3, 1800}) || !again.Slot("hack").Exists() {
		t.Fatalf("a stored run survives reopening, got %+v ok %v err %v", got, ok, err)
	}
	if again.Slot("snake").Exists() {
		t.Fatalf("each game has its own slot")
	}
	if err := again.Slot("hack").Clear(); err != nil {
		t.Fatal(err)
	}
	last, _ := OpenSaves(path)
	if last.Slot("hack").Exists() {
		t.Fatalf("a cleared slot stays empty after reopening")
	}
}

func TestARunThatDoesNotFitIsNoRun(t *testing.T) {
	s, _ := openSaves(t)
	s.Slot("hack").Store(map[string]string{"level": "three"})
	var got run
	if ok, err := s.Slot("hack").Load(&got); ok || err != nil {
		t.Fatalf("a run that does not decode must load as none, ok %v err %v", ok, err)
	}
}

func TestAnUnreadableSavesFileIsKeptAside(t *testing.T) {
	path := filepath.Join(t.TempDir(), "saves.json")
	os.WriteFile(path, []byte("garbage"), 0o600)
	s, err := OpenSaves(path)
	if err != nil || s.Slot("hack").Exists() {
		t.Fatalf("err %v", err)
	}
	if bad, _ := os.ReadFile(path + ".bad"); string(bad) != "garbage" {
		t.Fatalf("the unreadable file must be kept as .bad")
	}
}

func TestInMemorySavesNeverWrite(t *testing.T) {
	s, err := OpenSaves("")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Slot("hack").Store(run{Level: 1}); err != nil || !s.Slot("hack").Exists() {
		t.Fatalf("err %v", err)
	}
}
