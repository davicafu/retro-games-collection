package packet

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/scroll"
)

// Viability (S8): an autopilot that acts only once per reaction window, one
// switch at a time, survives SurviveTicks in every one of AutopilotSeeds
// games.
const (
	AutopilotSeeds = 30
	SurviveTicks   = 3 * 60 * 60 // 3 minutes, past top speed
)

func TestAutopilotSurvivesPastTopSpeed(t *testing.T) {
	for seed := 1; seed <= AutopilotSeeds; seed++ {
		t.Run(fmt.Sprintf("seed %d", seed), func(t *testing.T) {
			t.Parallel()
			g := NewGame(Config{Width: playfieldWidth, PlayerX: playerColumn, Link: startLink, Spawner: NewRandomSpawner(rand.New(rand.NewSource(int64(seed))))})
			for tick := 0; tick < SurviveTicks && !g.Dropped(); tick++ {
				var in Input
				if tick%ReactTicks == 0 {
					in.Switch = autopilot(g)
				}
				g.Step(in)
			}
			if g.Dropped() {
				t.Fatalf("dropped at distance %d, speed %d", g.Distance(), g.Speed())
			}
			if g.Speed() != MaxSpeed {
				t.Fatalf("speed %d after %d ticks, want the top speed %d", g.Speed(), SurviveTicks, MaxSpeed)
			}
		})
	}
}

// autopilot returns the switch to make now. It plans one decision per
// reaction window, a window being the cells the world will scroll in
// ReactTicks at the speed it will have by the window's end (rounded up, so
// the scroller's carry fits too): at each decision the packet keeps its link
// or moves one, and must then find that link free over the whole window. It
// looks as far as the right edge, where nothing is known yet.
func autopilot(g *Game) int {
	longest := MinGap(MaxSpeed) // a window at top speed, plus one
	window := func(x int) int {
		speed := speedRamp.At(g.Distance() + x - g.PlayerX() + longest)
		return (speed*ReactTicks + scroll.Unit - 1) / scroll.Unit
	}
	walls := g.Firewalls()
	free := func(link, from, to int) bool {
		for _, w := range walls {
			if w.Link == link && w.X <= to && w.X+w.Length > from {
				return false
			}
		}
		return true
	}
	type state struct{ x, link int }
	memo := map[state]bool{}
	// safe reports whether a packet deciding at column x on link can keep
	// going to the right edge.
	var safe func(x, link int) bool
	safe = func(x, link int) bool {
		if x >= g.Width() {
			return true
		}
		s := state{x, link}
		if v, ok := memo[s]; ok {
			return v
		}
		ok := false
		next := x + window(x)
		for _, d := range []int{0, -1, 1} {
			l := link + d
			if l >= 0 && l < Links && free(l, x, next) && safe(next, l) {
				ok = true
				break
			}
		}
		memo[s] = ok
		return ok
	}
	x := g.PlayerX()
	next := x + window(x)
	for _, d := range []int{0, -1, 1} {
		l := g.Link() + d
		if l >= 0 && l < Links && free(l, x, next) && safe(next, l) {
			return d
		}
	}
	return 0
}
