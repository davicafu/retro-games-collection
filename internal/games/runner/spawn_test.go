package runner

import (
	"math/rand"
	"testing"
)

// scriptedRand returns the given values in order, wrapping each into [0, n).
type scriptedRand struct {
	values []int
	next   int
}

func (r *scriptedRand) Intn(n int) int {
	v := r.values[r.next%len(r.values)]
	r.next++
	return v % n
}

func TestSpawnerAddsObstacleAtRightEdgeWhenWorldIsEmpty(t *testing.T) {
	s := NewRandomSpawner(&scriptedRand{values: []int{0}})

	got := s.Spawn(nil, 40, BaseSpeed)

	if len(got) != 1 {
		t.Fatalf("obstacles = %d, want 1", len(got))
	}
	if got[0].X != 40 {
		t.Fatalf("spawn x = %d, want right edge 40", got[0].X)
	}
	if got[0] != (Obstacle{X: 40, Width: obstacleShapes[0].Width, Tall: obstacleShapes[0].Tall, Kind: obstacleShapes[0].Kind}) {
		t.Fatalf("spawned %+v, want shape 0 %+v", got[0], obstacleShapes[0])
	}
}

func TestSpawnerWaitsForMinimumGap(t *testing.T) {
	// Shape 0, then extra gap 0 for the next obstacle.
	s := NewRandomSpawner(&scriptedRand{values: []int{0, 0}})
	world := s.Spawn(nil, 40, BaseSpeed)
	need := MinGap(BaseSpeed)

	// Scroll until the gap behind the last obstacle is one cell short.
	for 40-(world[0].X+world[0].Width) < need-1 {
		world[0].X--
	}
	if got := s.Spawn(world, 40, BaseSpeed); len(got) != 1 {
		t.Fatalf("spawned with gap %d < %d", need-1, need)
	}

	world[0].X--
	if got := s.Spawn(world, 40, BaseSpeed); len(got) != 2 {
		t.Fatalf("expected a spawn once the gap reached %d", need)
	}
}

func TestMinGapCoversAFullJumpAtThatSpeed(t *testing.T) {
	for _, speed := range []int{BaseSpeed, 60, MaxSpeed} {
		jumpCells := (len(jumpHeights)*speed + scrollUnit - 1) / scrollUnit
		if MinGap(speed) <= jumpCells {
			t.Fatalf("MinGap(%d) = %d, must exceed jump distance %d cells", speed, MinGap(speed), jumpCells)
		}
	}
	if MinGap(MaxSpeed) <= MinGap(BaseSpeed) {
		t.Fatalf("gap must grow with speed: %d at max vs %d at base", MinGap(MaxSpeed), MinGap(BaseSpeed))
	}
}

// Drive the spawner the way Game.scroll does, across speeds, and check every
// consecutive pair of obstacles keeps the minimum gap for the speed it spawned at.
func TestSpawnedObstaclesAlwaysKeepMinimumGap(t *testing.T) {
	const width = 48
	s := NewRandomSpawner(rand.New(rand.NewSource(7)))
	var world []Obstacle
	spawnSpeed := map[int]int{} // index in spawn order -> speed at spawn
	var all []Obstacle

	for step := 0; step < 5000; step++ {
		speed := BaseSpeed + (MaxSpeed-BaseSpeed)*step/5000
		for i := range world {
			world[i].X--
		}
		before := len(world)
		world = s.Spawn(world, width, speed)
		if len(world) > before {
			spawnSpeed[len(all)] = speed
			o := world[len(world)-1]
			o.X += step // world coordinates, independent of scrolling
			all = append(all, o)
		}
	}

	if len(all) < 50 {
		t.Fatalf("only %d obstacles spawned in 5000 cells", len(all))
	}
	for i := 1; i < len(all); i++ {
		gap := all[i].X - (all[i-1].X + all[i-1].Width)
		if gap < MinGap(spawnSpeed[i]) {
			t.Fatalf("obstacle %d gap = %d, want >= %d", i, gap, MinGap(spawnSpeed[i]))
		}
	}
	for i, o := range all {
		if !validShape(o) {
			t.Fatalf("obstacle %d has invalid shape %+v", i, o)
		}
	}
}

func validShape(o Obstacle) bool {
	for _, s := range obstacleShapes {
		if s.Width == o.Width && s.Tall == o.Tall && s.Kind == o.Kind {
			return true
		}
	}
	return false
}

func TestGameUsesSpawnerWhileScrolling(t *testing.T) {
	g := NewGame(Config{Width: 40, PlayerX: testPlayerX, Spawner: NewRandomSpawner(&scriptedRand{values: []int{0}})})

	for i := 0; i < 3; i++ { // first scroll happens on tick 3
		g.Step(ActionNone)
	}

	obstacles := g.Obstacles()
	if len(obstacles) != 1 || obstacles[0].X != 40 {
		t.Fatalf("obstacles = %+v, want one spawned at x=40", obstacles)
	}
}
