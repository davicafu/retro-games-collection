package snake

import (
	"math/rand"
	"testing"
)

func TestSequenceSelector_SelectsNextCandidateDeterministically(t *testing.T) {
	selector := NewSequenceSelector([]int{1})
	candidates := []Position{{X: 0, Y: 0}, {X: 2, Y: 2}, {X: 4, Y: 4}}

	got, ok := selector.Select(candidates)
	if !ok {
		t.Fatal("expected selector to choose a candidate")
	}
	if got != (Position{X: 2, Y: 2}) {
		t.Fatalf("expected second candidate, got %+v", got)
	}
}

func TestSequenceSelector_EmptyCandidatesReturnsFalse(t *testing.T) {
	selector := NewSequenceSelector([]int{0})

	_, ok := selector.Select(nil)
	if ok {
		t.Fatal("expected no selection for empty candidates")
	}
}

func TestSequenceSelector_EmptySequenceFallsBackToFirstCandidate(t *testing.T) {
	selector := NewSequenceSelector(nil)
	candidates := []Position{{X: 7, Y: 9}, {X: 2, Y: 3}}

	got, ok := selector.Select(candidates)
	if !ok {
		t.Fatal("expected fallback selection")
	}
	if got != candidates[0] {
		t.Fatalf("expected first candidate fallback, got %+v", got)
	}
}

func TestSequenceSelector_WrapsAcrossCalls(t *testing.T) {
	selector := NewSequenceSelector([]int{2, 0})
	candidates := []Position{{X: 1, Y: 1}, {X: 2, Y: 2}, {X: 3, Y: 3}}

	want := []Position{candidates[2], candidates[0], candidates[2], candidates[0]}
	for i := 0; i < len(want); i++ {
		got, ok := selector.Select(candidates)
		if !ok {
			t.Fatalf("call %d: expected selection", i)
		}
		if got != want[i] {
			t.Fatalf("call %d: expected %+v, got %+v", i, want[i], got)
		}
	}
}

func TestSequenceSelector_NegativeIndicesAreNormalized(t *testing.T) {
	selector := NewSequenceSelector([]int{-1, -4})
	candidates := []Position{{X: 10, Y: 10}, {X: 20, Y: 20}, {X: 30, Y: 30}}

	got1, ok1 := selector.Select(candidates)
	if !ok1 {
		t.Fatal("expected selection for first negative index")
	}
	if got1 != candidates[2] {
		t.Fatalf("expected normalized index to pick last candidate, got %+v", got1)
	}

	got2, ok2 := selector.Select(candidates)
	if !ok2 {
		t.Fatal("expected selection for second negative index")
	}
	if got2 != candidates[2] {
		t.Fatalf("expected normalized index to pick last candidate, got %+v", got2)
	}
}

func TestRandomSelector_EmptyCandidatesReturnsFalse(t *testing.T) {
	selector := NewRandomSelector(rand.New(rand.NewSource(42)))

	_, ok := selector.Select(nil)
	if ok {
		t.Fatal("expected no selection for empty candidates")
	}
}

func TestRandomSelector_DeterministicRNG(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	selector := NewRandomSelector(rng)
	candidates := []Position{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2}, {X: 3, Y: 3}}

	want := []Position{candidates[0], candidates[3], candidates[1]}
	for i := 0; i < len(want); i++ {
		got, ok := selector.Select(candidates)
		if !ok {
			t.Fatalf("call %d: expected selection", i)
		}
		if got != want[i] {
			t.Fatalf("call %d: expected %+v, got %+v", i, want[i], got)
		}
	}
}

func TestRandomSelector_NilRNGFallsBackToFirstCandidate(t *testing.T) {
	selector := NewRandomSelector(nil)
	candidates := []Position{{X: 5, Y: 5}, {X: 6, Y: 6}}

	got, ok := selector.Select(candidates)
	if !ok {
		t.Fatal("expected fallback selection")
	}
	if got != candidates[0] {
		t.Fatalf("expected first candidate fallback, got %+v", got)
	}
}
