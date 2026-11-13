package packet

import (
	"math/rand"
	"testing"
)

// gate is what one Spawn call placed.
type gate struct {
	x, length int
	free      links
	speed     int
	gap       int // free cells since the previous gate
	tokens    []Token
}

// spawnGates scrolls an empty world under a RandomSpawner for the given
// cells, at speeds rising like the game's, and returns the gates placed.
func spawnGates(seed int64, cells, width int) []gate {
	s := NewRandomSpawner(rand.New(rand.NewSource(seed)))
	var walls []Firewall
	var tokens []Token
	var gates []gate
	end := 0 // right end of the last gate, in world columns
	for d := 0; d < cells; d++ {
		speed := speedRamp.At(d + width)
		before, tokensBefore := len(walls), len(tokens)
		walls, tokens = s.Spawn(walls, tokens, width, speed)
		if len(walls) > before {
			g := gate{x: walls[before].X, length: walls[before].Length, free: allLinks, speed: speed, gap: width + d - end}
			for _, w := range walls[before:] {
				if w.X != g.x || w.Length != g.length {
					g.length = -1 // a gate's firewalls share column and length
				}
				g.free &^= 1 << w.Link
			}
			g.tokens = append(g.tokens, tokens[tokensBefore:]...)
			gates = append(gates, g)
			end = width + d + g.length
		}
		// Scroll one cell.
		kept := walls[:0]
		for _, w := range walls {
			if w.X--; w.X+w.Length > 0 {
				kept = append(kept, w)
			}
		}
		walls = kept
		for i := range tokens {
			tokens[i].X--
		}
	}
	return gates
}

func TestGatesBlockSomeButNeverAllLinks(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		for i, g := range spawnGates(seed, 5000, 60) {
			if g.free == 0 || g.free == allLinks {
				t.Fatalf("seed %d gate %d free links %04b, want 1 to %d", seed, i, g.free, Links-1)
			}
			if g.length < 1 || g.length > MaxGateLength {
				t.Fatalf("seed %d gate %d length %d, want firewalls of one length 1..%d", seed, i, g.length, MaxGateLength)
			}
		}
	}
}

func TestEveryGateIsReachableFromEveryFreeLinkOfThePrevious(t *testing.T) {
	seen := map[int]bool{} // reaches used, to show the rule is not trivial
	for seed := int64(1); seed <= 20; seed++ {
		gates := spawnGates(seed, 5000, 60)
		for i := 1; i < len(gates); i++ {
			prev, g := gates[i-1], gates[i]
			if g.gap < MinGap(g.speed) {
				t.Fatalf("seed %d gate %d: gap %d, want at least %d", seed, i, g.gap, MinGap(g.speed))
			}
			reach := Reach(g.gap, g.speed)
			if !fair(prev.free, g.free, reach) {
				t.Fatalf("seed %d gate %d: free %04b not within %d links of every free link of %04b", seed, i, g.free, reach, prev.free)
			}
			if !fair(prev.free, g.free, reach-1) || reach == 1 {
				seen[reach] = true
			}
		}
	}
	if !seen[1] || len(seen) < 2 {
		t.Fatalf("reaches needed %v, want switches of one link and of more", seen)
	}
}

func TestTokensSitOnAFreeLinkOfTheirGate(t *testing.T) {
	carried := 0
	for seed := int64(1); seed <= 20; seed++ {
		for i, g := range spawnGates(seed, 5000, 60) {
			for _, tk := range g.tokens {
				carried++
				if !g.free.has(tk.Link) || tk.X < g.x || tk.X >= g.x+g.length {
					t.Fatalf("seed %d gate %d: token %+v outside the free links %04b of the gate at %d+%d", seed, i, tk, g.free, g.x, g.length)
				}
			}
		}
	}
	if carried == 0 {
		t.Fatalf("no gate carried a token")
	}
}
