package packet

import "github.com/davidbarreiro/cli-games-collection/internal/engine/scroll"

// Rand is the randomness the spawner needs; *math/rand.Rand satisfies it.
type Rand interface {
	Intn(n int) int
}

const (
	// ReactTicks is the reaction window the gates leave for each switch:
	// 0.3 s.
	ReactTicks = 18

	// MaxGateLength is the longest a gate's firewalls run, in cells.
	MaxGateLength = 10
	// MaxExtraGap is the random spacing added between gates, exclusive.
	MaxExtraGap = 16
	// TokenChance is one in how many gates carry a token.
	TokenChance = 3
)

// MinGap is the fewest free cells between two gates at a speed: the cells
// the world scrolls during one reaction window, plus one.
func MinGap(speed int) int {
	return (speed*ReactTicks+scroll.Unit-1)/scroll.Unit + 1
}

// Reach is how many links the packet can be asked to switch across a gap
// of free cells: one per reaction window, and at least one.
func Reach(gap, speed int) int {
	return max(1, gap/MinGap(speed))
}

// links is a set of links, bit i for link i.
type links uint8

const allLinks links = 1<<Links - 1

func (s links) has(link int) bool { return s&(1<<link) != 0 }

// fair reports whether every link of from has a link of to within reach.
func fair(from, to links, reach int) bool {
	for a := 0; a < Links; a++ {
		if !from.has(a) {
			continue
		}
		ok := false
		for b := max(a-reach, 0); b <= min(a+reach, Links-1); b++ {
			ok = ok || to.has(b)
		}
		if !ok {
			return false
		}
	}
	return true
}

// RandomSpawner places gates: each blocks 1 to Links-1 links over one
// length, and leaves free links the packet can reach from every free link
// of the gate before it (see fair and Reach).
type RandomSpawner struct {
	rng      Rand
	free     links // the last gate's free links
	extraGap int
}

func NewRandomSpawner(rng Rand) *RandomSpawner {
	return &RandomSpawner{rng: rng, free: allLinks}
}

func (s *RandomSpawner) Spawn(firewalls []Firewall, tokens []Token, width, speed int) ([]Firewall, []Token) {
	gap := width
	if len(firewalls) > 0 {
		end := 0
		for _, f := range firewalls {
			end = max(end, f.X+f.Length)
		}
		if gap = width - end; gap < MinGap(speed)+s.extraGap {
			return firewalls, tokens
		}
	}
	// Choose among every fair set of free links. There is always one: the
	// last gate's own free links, or the two middle links after a start
	// with every link free.
	var choices []links
	for free := links(1); free < allLinks; free++ {
		if fair(s.free, free, Reach(gap, speed)) {
			choices = append(choices, free)
		}
	}
	free := choices[s.rng.Intn(len(choices))]
	length := 1 + s.rng.Intn(MaxGateLength)
	for link := 0; link < Links; link++ {
		if !free.has(link) {
			firewalls = append(firewalls, Firewall{Link: link, X: width, Length: length})
		}
	}
	if s.rng.Intn(TokenChance) == 0 {
		var open []int
		for link := 0; link < Links; link++ {
			if free.has(link) {
				open = append(open, link)
			}
		}
		tokens = append(tokens, Token{Link: open[s.rng.Intn(len(open))], X: width + length/2})
	}
	s.free = free
	s.extraGap = s.rng.Intn(MaxExtraGap)
	return firewalls, tokens
}
