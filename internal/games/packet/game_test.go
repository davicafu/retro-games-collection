package packet

import (
	"testing"
	"time"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/scroll"
)

func newTestGame(link int, walls []Firewall, tokens []Token) *Game {
	return NewGame(Config{Width: 40, PlayerX: 5, Link: link, Firewalls: walls, Tokens: tokens})
}

// stepUntil steps with no input until done or n ticks pass.
func stepUntil(g *Game, n int, done func() bool) {
	for i := 0; i < n && !done(); i++ {
		g.Step(Input{})
	}
}

func TestSwitchingMovesOneLinkPerPressWithinTheLinks(t *testing.T) {
	g := newTestGame(1, nil, nil)
	g.Step(Input{Switch: 1})
	if g.Link() != 2 {
		t.Fatalf("down from link 1 = link %d, want 2", g.Link())
	}
	g.Step(Input{Switch: 1})
	g.Step(Input{Switch: 1})
	if g.Link() != Links-1 {
		t.Fatalf("the bottom link must stop the packet, at %d", g.Link())
	}
	for i := 0; i < Links+1; i++ {
		g.Step(Input{Switch: -1})
	}
	if g.Link() != 0 {
		t.Fatalf("the top link must stop the packet, at %d", g.Link())
	}
}

func TestScrollingIntoAFirewallDropsThePacket(t *testing.T) {
	g := newTestGame(1, []Firewall{{Link: 1, X: 10, Length: 3}}, nil)
	stepUntil(g, 1000, func() bool { return g.Dropped() || g.Distance() > 10 })
	if !g.Dropped() || g.Distance() != 5 {
		t.Fatalf("dropped %v at distance %d, want dropped when the firewall reaches column 5", g.Dropped(), g.Distance())
	}
	d := g.Distance()
	steps(g, 100)
	if g.Distance() != d {
		t.Fatalf("a dropped packet must not advance")
	}
}

func TestSwitchingIntoAFirewallDropsThePacket(t *testing.T) {
	g := newTestGame(1, []Firewall{{Link: 2, X: 4, Length: 3}}, nil)
	g.Step(Input{Switch: 1})
	if !g.Dropped() {
		t.Fatalf("switching into a firewall must drop the packet")
	}
}

func TestSwitchingOntoTheLastCellOfAFirewallDropsThePacket(t *testing.T) {
	// The world scrolls one cell every 4th tick at BaseSpeed: the switch on
	// tick 4 meets the firewall's last cell before it scrolls past.
	g := newTestGame(1, []Firewall{{Link: 2, X: 5, Length: 1}}, nil)
	steps(g, 3)
	g.Step(Input{Switch: 1})
	if !g.Dropped() || g.Distance() != 0 {
		t.Fatalf("dropped %v at distance %d, want dropped before the scroll", g.Dropped(), g.Distance())
	}
}

func TestFirewallsOnOtherLinksPassByAndLeave(t *testing.T) {
	g := newTestGame(0, []Firewall{{Link: 1, X: 6, Length: 2}, {Link: 3, X: 20, Length: 4}}, nil)
	stepUntil(g, 10000, func() bool { return len(g.Firewalls()) == 0 })
	if g.Dropped() || len(g.Firewalls()) != 0 {
		t.Fatalf("dropped %v, firewalls left %v", g.Dropped(), g.Firewalls())
	}
	if g.Distance() != 24 {
		t.Fatalf("the last firewall left at distance %d, want 24 (once fully past column 0)", g.Distance())
	}
}

func TestTakingATokenScores(t *testing.T) {
	g := newTestGame(2, nil, []Token{{Link: 2, X: 8}, {Link: 1, X: 8}})
	stepUntil(g, 1000, func() bool { return g.Distance() == 3 })
	if g.Taken() != 1 || g.Score() != 3+PointsToken {
		t.Fatalf("taken %d, score %d, want 1 and %d", g.Taken(), g.Score(), 3+PointsToken)
	}
	if len(g.Tokens()) != 1 || g.Tokens()[0].Link != 1 {
		t.Fatalf("tokens = %v, want only the one on link 1", g.Tokens())
	}
	g.Step(Input{Switch: -1}) // onto the other token, also at column 5
	if g.Taken() != 2 {
		t.Fatalf("switching onto a token must take it, taken %d", g.Taken())
	}
}

func TestSpeedRisesWithDistanceUpToTheMax(t *testing.T) {
	g := newTestGame(0, nil, nil)
	if g.Speed() != BaseSpeed {
		t.Fatalf("start speed %d, want %d", g.Speed(), BaseSpeed)
	}
	stepUntil(g, 100000, func() bool { return g.Distance() >= SpeedDistance })
	if g.Speed() != BaseSpeed+SpeedStep {
		t.Fatalf("speed after %d cells = %d, want %d", SpeedDistance, g.Speed(), BaseSpeed+SpeedStep)
	}
	stepUntil(g, 1000000, func() bool { return g.Distance() >= 100*SpeedDistance })
	if g.Speed() != MaxSpeed || MaxSpeed >= scroll.Unit {
		t.Fatalf("top speed %d, want %d below one cell per tick", g.Speed(), MaxSpeed)
	}
}

// TestRealTimeSpeedsDoNotDependOnTheTickRate pins speeds per second, so a
// change of TickPeriod must rescale the per-tick constants with it.
func TestRealTimeSpeedsDoNotDependOnTheTickRate(t *testing.T) {
	perSecond := int(time.Second / TickPeriod)
	cellsPerSecond := func(speed int) float64 { return float64(speed*perSecond) / scroll.Unit }
	if got := cellsPerSecond(BaseSpeed); got != 15 {
		t.Fatalf("start speed %.1f cells/s, want 15", got)
	}
	if got := cellsPerSecond(MaxSpeed); got < 39 || got > 40 {
		t.Fatalf("top speed %.1f cells/s, want about 40", got)
	}
	if got := (time.Duration(ReactTicks) * TickPeriod).Round(time.Millisecond); got != 300*time.Millisecond {
		t.Fatalf("reaction window %v, want 300ms", got)
	}
}

func steps(g *Game, n int) {
	for i := 0; i < n; i++ {
		g.Step(Input{})
	}
}
