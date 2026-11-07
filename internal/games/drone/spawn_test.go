package drone

import (
	"math/rand"
	"testing"
)

// scriptedRand returns values in order, then zeros.
type scriptedRand struct {
	values []int
	idx    int
}

func (r *scriptedRand) Intn(n int) int {
	if r.idx >= len(r.values) {
		return 0
	}
	v := r.values[r.idx] % n
	r.idx++
	return v
}

func TestFirstAntennaSpawnsAtTheRightEdgeWithinBounds(t *testing.T) {
	s := NewRandomSpawner(&scriptedRand{values: []int{0}})
	got := s.Spawn(nil, 40, 12)
	if len(got) != 1 || got[0].X != 40 {
		t.Fatalf("spawned %+v, want one antenna at X=40", got)
	}
	if got[0].GapTop != minGapTop {
		t.Fatalf("gap top = %d, want lowest allowed %d for Intn=0", got[0].GapTop, minGapTop)
	}
}

func TestNextAntennaWaitsForTheSpacing(t *testing.T) {
	s := NewRandomSpawner(&scriptedRand{})
	last := Antenna{X: 40 - AntennaSpacing + 1, GapTop: 4}
	if got := s.Spawn([]Antenna{last}, 40, 12); len(got) != 1 {
		t.Fatalf("spawned before the spacing was reached: %+v", got)
	}
	last.X--
	if got := s.Spawn([]Antenna{last}, 40, 12); len(got) != 2 || got[1].X != 40 {
		t.Fatalf("expected a new antenna at X=40 once the spacing is reached, got %+v", got)
	}
}

// TestGapsStayReachableOverALongRun scrolls 5000 cells with a seeded RNG and
// checks every antenna: gap inside the sky with a visible mast above and
// below, at most MaxGapShift rows from the previous gap, fixed spacing.
func TestGapsStayReachableOverALongRun(t *testing.T) {
	const width, height = 48, 14
	s := NewRandomSpawner(rand.New(rand.NewSource(7)))
	var world, all []Antenna
	for cell := 0; cell < 5000; cell++ {
		kept := world[:0]
		for _, a := range world {
			a.X--
			if a.X+AntennaWidth > 0 {
				kept = append(kept, a)
			}
		}
		before := len(kept)
		world = s.Spawn(kept, width, height)
		all = append(all, world[before:]...)
	}

	if len(all) < 100 {
		t.Fatalf("only %d antennas spawned", len(all))
	}
	shifts := map[int]bool{}
	for i, a := range all {
		if a.GapTop < minGapTop || a.GapTop+GapSize > height-1 {
			t.Fatalf("antenna %d gap %d..%d leaves no mast inside sky of %d rows", i, a.GapTop, a.GapTop+GapSize-1, height)
		}
		if i == 0 {
			continue
		}
		shift := a.GapTop - all[i-1].GapTop
		if shift < -MaxGapShift || shift > MaxGapShift {
			t.Fatalf("antenna %d gap moved %d rows, max %d", i, shift, MaxGapShift)
		}
		shifts[shift] = true
	}
	if len(shifts) < 3 {
		t.Fatalf("gaps barely vary: shifts seen %v", shifts)
	}
}

// autopilot thrusts when the drone sinks below the lower part of the next gap.
func autopilot(g *Game) bool {
	for _, a := range g.Antennas() {
		if a.X+AntennaWidth > g.DroneX() {
			floor := (a.GapTop + GapSize - 2) * SubCells
			return g.Position()+g.Velocity() > floor
		}
	}
	return g.Position() > g.Height()/2*SubCells
}

// TestEveryGapIsReachableAtMaxSpeed plays a long seeded run with the speed
// already at its cap: a simple autopilot must never crash, so the spawner
// never produces an impossible gap.
func TestEveryGapIsReachableAtMaxSpeed(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		g := NewGame(Config{Width: 48, Height: 14, DroneX: 8, Spawner: NewRandomSpawner(rand.New(rand.NewSource(seed)))})
		g.distance = 100000 // speed at MaxSpeed from the start
		g.Step(true)
		for tick := 0; tick < 8000; tick++ {
			g.Step(autopilot(g))
			if g.Dead() {
				t.Fatalf("seed %d: crashed at tick %d, row %d, score %d, antennas %+v", seed, tick, g.Row(), g.Score(), g.Antennas())
			}
		}
		if g.Score() < 300 {
			t.Fatalf("seed %d: score %d, autopilot barely passed antennas", seed, g.Score())
		}
	}
}
