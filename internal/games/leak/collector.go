package leak

import "github.com/davidbarreiro/cli-games-collection/internal/engine/grid"

// Kind is a collector's chasing strategy.
type Kind int

const (
	// KindTracer heads for the leak's cell.
	KindTracer Kind = iota
	// KindInterceptor heads for the cell up to LeadCells ahead of the leak.
	KindInterceptor
	// KindWanderer turns at random until the leak is within SightCells,
	// then heads for it.
	KindWanderer
	kindCount
)

type collectorState int

const (
	stateWaiting collectorState = iota // in the pool until releaseAt
	stateLeaving                       // walking out through the gate
	stateActive                        // in the corridors
)

// Collector is a garbage collector chasing the leak.
type Collector struct {
	Kind Kind
	Actor
	home      grid.Point
	state     collectorState
	releaseAt int // lifeTicks at which a waiting collector leaves
	fleeing   bool
}

// Fleeing reports that a malloc made this collector prey.
func (c Collector) Fleeing() bool { return c.fleeing }

// InPool reports that the collector has not reached the corridors yet.
func (c Collector) InPool() bool { return c.state != stateActive }

// sendHome puts an eaten collector back in the pool until the given tick.
func (c *Collector) sendHome(releaseAt int) {
	c.Actor = Actor{Cell: c.home}
	c.state = stateWaiting
	c.releaseAt = releaseAt
	c.fleeing = false
}

func (g *Game) moveCollector(c *Collector) {
	if c.state == stateWaiting {
		if g.lifeTicks < c.releaseAt {
			return
		}
		c.state = stateLeaving
	}
	speed := g.collectorSpeed()
	if c.fleeing {
		speed = FleeSpeed
	}
	c.advance(speed, func() grid.Point { return g.collectorChoice(c) })
}

// collectorChoice picks the next direction at a cell centre. A collector
// never turns back unless it has no other way; one leaving the pool becomes
// active on the cell outside the gate.
func (g *Game) collectorChoice(c *Collector) grid.Point {
	if c.state == stateLeaving && c.Cell == g.exit {
		c.state = stateActive
	}
	open := g.collectorCanEnter(c.state)
	var options []grid.Point
	for _, d := range grid.Neighbours {
		if d != neg(c.Dir) && open(c.Cell.Add(d)) {
			options = append(options, d)
		}
	}
	if len(options) == 0 {
		if back := neg(c.Dir); back != (grid.Point{}) && open(c.Cell.Add(back)) {
			return back
		}
		return grid.Point{}
	}

	var field grid.DistanceField
	switch {
	case c.state == stateLeaving:
		field = g.toExit
	case c.fleeing:
		return options[g.rng.Intn(len(options))]
	case g.paused():
		field = g.corners[c.Kind]
	default:
		toLeak := grid.Distances(g.width, g.height, g.leak.Occupied(), open)
		target := g.leak.Occupied()
		switch c.Kind {
		case KindInterceptor:
			target = g.ahead(LeadCells)
		case KindWanderer:
			if d := toLeak.At(c.Cell); d == grid.Unreachable || d > SightCells {
				return options[g.rng.Intn(len(options))]
			}
		}
		field = toLeak
		if target != g.leak.Occupied() {
			field = grid.Distances(g.width, g.height, target, open)
		}
	}

	best, bestDist := options[0], -1
	for _, d := range options {
		dist := field.At(c.Cell.Add(d))
		if dist == grid.Unreachable {
			continue
		}
		if bestDist == -1 || dist < bestDist {
			best, bestDist = d, dist
		}
	}
	return best
}

// cornerOf returns the open cell nearest to the map corner where a kind
// pauses: Tracer top right, Interceptor top left, Wanderer bottom right.
func (g *Game) cornerOf(k Kind) grid.Point {
	corner := [...]grid.Point{
		KindTracer:      {X: g.width - 1, Y: 0},
		KindInterceptor: {X: 0, Y: 0},
		KindWanderer:    {X: g.width - 1, Y: g.height - 1},
	}[k]
	best, bestDist := grid.Point{}, -1
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			p := grid.Point{X: x, Y: y}
			if !g.leakCanEnter(p) {
				continue
			}
			if d := abs(p.X-corner.X) + abs(p.Y-corner.Y); bestDist < 0 || d < bestDist {
				best, bestDist = p, d
			}
		}
	}
	return best
}

// ahead walks from the leak along its direction for up to n open cells.
func (g *Game) ahead(n int) grid.Point {
	p := g.leak.Occupied()
	if g.leak.Dir == (grid.Point{}) {
		return p
	}
	for i := 0; i < n && g.leakCanEnter(p.Add(g.leak.Dir)); i++ {
		p = p.Add(g.leak.Dir)
	}
	return p
}

// collectorCanEnter returns where a collector in the given state may go:
// the gate and the pool only while it is still leaving.
func (g *Game) collectorCanEnter(state collectorState) func(grid.Point) bool {
	return func(p grid.Point) bool {
		t, ok := g.tile(p)
		if !ok || t == TileWall {
			return false
		}
		if t == TileGate || t == TilePool {
			return state != stateActive
		}
		return true
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
