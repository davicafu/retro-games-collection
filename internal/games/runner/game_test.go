package runner

import (
	"reflect"
	"testing"
)

const testPlayerX = 4

func newTestGame(obstacles ...Obstacle) *Game {
	return NewGame(Config{Width: 40, PlayerX: testPlayerX, Obstacles: obstacles})
}

// stepHeights advances the game once per action and records the player height after each tick.
func stepHeights(g *Game, actions ...Action) []int {
	heights := make([]int, 0, len(actions))
	for _, a := range actions {
		g.Step(a)
		heights = append(heights, g.PlayerHeight())
	}
	return heights
}

func repeat(a Action, n int) []Action {
	out := make([]Action, n)
	for i := range out {
		out[i] = a
	}
	return out
}

func TestJumpFollowsHeightTableThenLands(t *testing.T) {
	g := newTestGame()

	got := stepHeights(g, append([]Action{ActionJump}, repeat(ActionNone, 9)...)...)

	want := []int{1, 2, 3, 4, 4, 4, 3, 2, 1, 0}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("jump heights = %v, want %v", got, want)
	}
}

func TestJumpWhileAirborneIsIgnored(t *testing.T) {
	g := newTestGame()

	got := stepHeights(g, ActionJump, ActionNone, ActionJump, ActionNone)

	want := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("heights = %v, want %v (mid-air jump must not restart the jump)", got, want)
	}
}

func TestDuckLastsFixedTicksAndRepressRestartsIt(t *testing.T) {
	g := newTestGame()

	g.Step(ActionDuck)
	for i := 1; i < DuckTicks; i++ {
		if !g.Ducking() {
			t.Fatalf("expected ducking on tick %d of %d", i+1, DuckTicks)
		}
		g.Step(ActionNone)
	}
	if !g.Ducking() {
		t.Fatalf("expected ducking on last duck tick")
	}
	g.Step(ActionNone)
	if g.Ducking() {
		t.Fatalf("expected duck to end after %d ticks", DuckTicks)
	}

	g.Step(ActionDuck)
	for i := 0; i < DuckTicks-2; i++ {
		g.Step(ActionNone)
	}
	g.Step(ActionDuck) // restart the timer before it runs out
	for i := 1; i < DuckTicks; i++ {
		g.Step(ActionNone)
	}
	if !g.Ducking() {
		t.Fatalf("expected re-pressing duck to restart the timer")
	}
}

func TestDuckIsIgnoredWhileAirborne(t *testing.T) {
	g := newTestGame()

	g.Step(ActionJump)
	g.Step(ActionDuck)

	if g.Ducking() {
		t.Fatalf("duck must be ignored while airborne")
	}
	if g.PlayerHeight() != 2 {
		t.Fatalf("jump must continue, height = %d, want 2", g.PlayerHeight())
	}
}

func TestJumpCancelsDuck(t *testing.T) {
	g := newTestGame()

	g.Step(ActionDuck)
	g.Step(ActionJump)

	if g.Ducking() {
		t.Fatalf("jump must cancel duck")
	}
	if g.PlayerHeight() != 1 {
		t.Fatalf("height = %d, want 1", g.PlayerHeight())
	}
}

func TestWorldScrollsByAccumulatedSpeed(t *testing.T) {
	g := newTestGame(Obstacle{X: 30, Width: 1, Tall: 1, Kind: KindGround})

	for i := 0; i < 2; i++ {
		g.Step(ActionNone)
	}
	if x := g.Obstacles()[0].X; x != 30 {
		t.Fatalf("after 2 ticks at speed 35, x = %d, want 30 (70 < 100 accumulated)", x)
	}

	g.Step(ActionNone)
	if x := g.Obstacles()[0].X; x != 29 {
		t.Fatalf("after 3 ticks, x = %d, want 29", x)
	}

	for i := 0; i < 17; i++ {
		g.Step(ActionNone)
	}
	if x := g.Obstacles()[0].X; x != 23 {
		t.Fatalf("after 20 ticks, x = %d, want 23 (700 accumulated = 7 cells)", x)
	}
	if g.Score() != 7 {
		t.Fatalf("score = %d, want 7 (distance in cells)", g.Score())
	}
}

func TestObstaclesLeavingTheScreenAreDropped(t *testing.T) {
	g := newTestGame(Obstacle{X: 0, Width: 1, Tall: 1, Kind: KindGround})

	for i := 0; i < 3; i++ {
		g.Step(ActionNone)
	}

	if n := len(g.Obstacles()); n != 0 {
		t.Fatalf("obstacles = %d, want 0 once fully off-screen", n)
	}
}

func TestSpeedScalesWithDistanceAndCaps(t *testing.T) {
	g := newTestGame()
	if g.Speed() != BaseSpeed {
		t.Fatalf("initial speed = %d, want %d", g.Speed(), BaseSpeed)
	}

	for g.Score() < 39 {
		g.Step(ActionNone)
	}
	if g.Speed() != BaseSpeed {
		t.Fatalf("speed at distance 39 = %d, want %d", g.Speed(), BaseSpeed)
	}
	for g.Score() < 40 {
		g.Step(ActionNone)
	}
	if g.Speed() != BaseSpeed+SpeedStep {
		t.Fatalf("speed at distance 40 = %d, want %d (speed rises every 40 cells)", g.Speed(), BaseSpeed+SpeedStep)
	}

	for g.Score() < 2000 {
		g.Step(ActionNone)
		if g.Speed() > MaxSpeed {
			t.Fatalf("speed %d exceeded cap %d", g.Speed(), MaxSpeed)
		}
	}
	if g.Speed() != MaxSpeed {
		t.Fatalf("speed = %d, want cap %d", g.Speed(), MaxSpeed)
	}
}

// An obstacle placed one cell ahead of the player reaches the player column on
// tick 3 (accumulator 105) and stays there through tick 5 (next scroll on tick 6).

func TestStandingPlayerHitsGroundObstacle(t *testing.T) {
	g := newTestGame(Obstacle{X: testPlayerX + 1, Width: 1, Tall: 1, Kind: KindGround})

	stepHeights(g, repeat(ActionNone, 2)...)
	if g.Dead() {
		t.Fatalf("dead before the obstacle arrived")
	}
	g.Step(ActionNone)
	if !g.Dead() {
		t.Fatalf("expected collision when the ground obstacle reaches the player")
	}
}

func TestJumpClearsTallGroundObstacle(t *testing.T) {
	g := newTestGame(Obstacle{X: testPlayerX + 1, Width: 1, Tall: 2, Kind: KindGround})

	g.Step(ActionJump)
	for i := 0; i < 10; i++ {
		g.Step(ActionNone)
	}

	if g.Dead() {
		t.Fatalf("jump should clear a 2-tall obstacle (heights 3,4,4 while overlapping)")
	}
}

func TestStandingPlayerHitsFlyingObstacle(t *testing.T) {
	g := newTestGame(Obstacle{X: testPlayerX + 1, Width: 1, Kind: KindFlying})

	stepHeights(g, repeat(ActionNone, 3)...)

	if !g.Dead() {
		t.Fatalf("a standing player's head must hit a flying obstacle")
	}
}

func TestDuckPassesUnderFlyingObstacle(t *testing.T) {
	g := newTestGame(Obstacle{X: testPlayerX + 1, Width: 1, Kind: KindFlying})

	g.Step(ActionDuck)
	for i := 0; i < 6; i++ {
		g.Step(ActionNone)
	}

	if g.Dead() {
		t.Fatalf("ducking must pass under a flying obstacle")
	}
}

func TestDuckingPlayerStillHitsGroundObstacle(t *testing.T) {
	g := newTestGame(Obstacle{X: testPlayerX + 1, Width: 1, Tall: 1, Kind: KindGround})

	g.Step(ActionDuck)
	g.Step(ActionNone)
	g.Step(ActionNone)

	if !g.Dead() {
		t.Fatalf("ducking must not avoid a ground obstacle")
	}
}

func TestDeadGameDoesNotAdvance(t *testing.T) {
	g := newTestGame(Obstacle{X: testPlayerX + 1, Width: 1, Tall: 1, Kind: KindGround})
	stepHeights(g, repeat(ActionNone, 3)...)
	if !g.Dead() {
		t.Fatalf("setup: expected dead")
	}
	score := g.Score()
	x := g.Obstacles()[0].X

	stepHeights(g, ActionJump, ActionNone, ActionNone, ActionNone)

	if g.Score() != score || g.Obstacles()[0].X != x || g.PlayerHeight() != 0 {
		t.Fatalf("dead game advanced: score %d->%d, x %d->%d, height %d", score, g.Score(), x, g.Obstacles()[0].X, g.PlayerHeight())
	}
}

func TestMaxSpeedNeverScrollsMoreThanOneCellPerTick(t *testing.T) {
	if MaxSpeed >= scrollUnit {
		t.Fatalf("MaxSpeed %d must stay below %d so obstacles cannot tunnel through the player", MaxSpeed, scrollUnit)
	}
}
