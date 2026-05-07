package snake

import "math/rand"

type FoodSelector interface {
	Select(candidates []Position) (Position, bool)
}

type SequenceSelector struct {
	indices []int
	idx     int
}

func NewSequenceSelector(indices []int) *SequenceSelector {
	return &SequenceSelector{indices: append([]int(nil), indices...)}
}

func (s *SequenceSelector) Select(candidates []Position) (Position, bool) {
	if len(candidates) == 0 {
		return Position{}, false
	}
	if s == nil {
		return candidates[0], true
	}
	if len(s.indices) == 0 {
		return candidates[0], true
	}
	choice := s.indices[s.idx%len(s.indices)]
	s.idx++
	candidateIdx := choice % len(candidates)
	if candidateIdx < 0 {
		candidateIdx += len(candidates)
	}
	return candidates[candidateIdx], true
}

type RandomSelector struct{ rng *rand.Rand }

func NewRandomSelector(rng *rand.Rand) *RandomSelector { return &RandomSelector{rng: rng} }

func (s *RandomSelector) Select(candidates []Position) (Position, bool) {
	if len(candidates) == 0 {
		return Position{}, false
	}
	if s == nil || s.rng == nil {
		return candidates[0], true
	}
	return candidates[s.rng.Intn(len(candidates))], true
}
