package breach

import "testing"

// strength is the number of hits needed to clear a level.
func strength(l Level) int {
	total := 0
	for _, b := range l.Bricks() {
		if !b.Solid {
			total += b.HP
		}
	}
	return total
}

func TestCampaignHasAtLeastTenLevelsThatFitThePlayfield(t *testing.T) {
	if len(Campaign) < 10 {
		t.Fatalf("campaign has %d levels, want at least 10", len(Campaign))
	}
	for i, l := range Campaign {
		if len(l.Wall) == 0 || FirewallTop+len(l.Wall) > playfieldHeight-6 {
			t.Fatalf("level %d: %d wall rows leave too little room above the injector", i+1, len(l.Wall))
		}
		for _, row := range l.Wall {
			if len(row)*BrickWidth != playfieldWidth {
				t.Fatalf("level %d row %q does not span the %d-cell playfield", i+1, row, playfieldWidth)
			}
		}
	}
}

func TestEveryLevelIsHarderThanTheOneBefore(t *testing.T) {
	for i := 1; i < len(Campaign); i++ {
		prev, cur := Campaign[i-1], Campaign[i]
		if strength(cur) <= strength(prev) {
			t.Fatalf("level %d strength %d, not above level %d's %d", i+1, strength(cur), i, strength(prev))
		}
		if cur.Speed <= prev.Speed {
			t.Fatalf("level %d speed %d%%, not above level %d's %d%%", i+1, cur.Speed, i, prev.Speed)
		}
		if cur.DropOdds < prev.DropOdds {
			t.Fatalf("level %d drops more often than level %d", i+1, i)
		}
	}
	if Campaign[0].Speed != 100 {
		t.Fatalf("level 1 speed = %d%%, want the base 100%%", Campaign[0].Speed)
	}
}

func TestLaterLevelsIntroduceTougherBlocks(t *testing.T) {
	has := func(l Level, match func(Brick) bool) bool {
		for _, b := range l.Bricks() {
			if match(b) {
				return true
			}
		}
		return false
	}
	threeHit := func(b Brick) bool { return b.HP == 3 }
	solid := func(b Brick) bool { return b.Solid }
	if has(Campaign[0], threeHit) || has(Campaign[0], solid) {
		t.Fatalf("level 1 must only have basic blocks")
	}
	last := Campaign[len(Campaign)-1]
	if !has(last, threeHit) || !has(last, solid) {
		t.Fatalf("the last level must have 3-hit and indestructible blocks")
	}
}

func TestEveryLevelSpeedKeepsTheVirusBelowOneCellPerTick(t *testing.T) {
	for i, l := range Campaign {
		for _, v := range []Velocity{Steep, Diagonal, Shallow} {
			s := v.scaled(l.Speed)
			if s.X >= SubCells || s.Y >= SubCells {
				t.Fatalf("level %d: angle %+v at %d%% = %+v tunnels", i+1, v, l.Speed, s)
			}
		}
	}
}

func TestLevelWallParsing(t *testing.T) {
	l := Level{Wall: []string{"1.#", "23."}}
	want := []Brick{
		{X: 0, Y: FirewallTop, HP: 1},
		{X: 2 * BrickWidth, Y: FirewallTop, Solid: true},
		{X: 0, Y: FirewallTop + 1, HP: 2},
		{X: BrickWidth, Y: FirewallTop + 1, HP: 3},
	}
	got := l.Bricks()
	if len(got) != len(want) {
		t.Fatalf("bricks = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("brick %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestSolidBlockBouncesWithoutDamageOrScore(t *testing.T) {
	g := newTestGame(Brick{X: 8, Y: 3, Solid: true}, farBrick)
	place(g, 9, 4, 0, -60)
	g.Step(Input{})
	if _, vy := g.Velocity(); vy != 60 {
		t.Fatalf("vy = %d, want reflected", vy)
	}
	if len(g.Bricks()) != 2 || g.Score() != 0 {
		t.Fatalf("solid block damaged or scored: bricks=%+v score=%d", g.Bricks(), g.Score())
	}
}

func TestOnlySolidBlocksLeftClearsTheLevel(t *testing.T) {
	g := newTestGame(Brick{X: 8, Y: 3, HP: 1}, Brick{X: 0, Y: 1, Solid: true})
	place(g, 9, 4, 0, -60)
	g.Step(Input{})
	if !g.Won() {
		t.Fatalf("a level with only indestructible blocks left must be cleared")
	}
}

func twoLevels() *Game {
	first := Level{Wall: []string{"1"}, Speed: 100, DropOdds: 5}
	second := Level{Wall: []string{"22", "11"}, Speed: 150, DropOdds: 8}
	return NewGame(Config{Width: testWidth, Height: testHeight, Levels: []Level{first, second}})
}

func TestClearingALevelLoadsTheNextOneKeepingScoreAndLives(t *testing.T) {
	g := twoLevels()
	g.lives = 2
	g.drops = []Drop{{Col: 15, Y: 100, Kind: PowerUpLife}}
	g.wideTicks = 50
	place(g, 1, FirewallTop+1, 0, -60)
	g.Step(Input{})

	if g.Won() || g.Level() != 2 || g.LevelCount() != 2 {
		t.Fatalf("won=%v level=%d/%d, want level 2 of 2 in play", g.Won(), g.Level(), g.LevelCount())
	}
	if len(g.Bricks()) != 4 || g.Launched() {
		t.Fatalf("level 2 must load its wall with the virus back on the injector")
	}
	if g.Score() != PointsPerHit || g.Lives() != 2 {
		t.Fatalf("score=%d lives=%d, want them carried over", g.Score(), g.Lives())
	}
	if len(g.Drops()) != 0 || g.WideLeft() != 0 {
		t.Fatalf("drops and Wide must reset between levels")
	}
}

func TestLevelSpeedScalesTheLaunch(t *testing.T) {
	g := twoLevels()
	place(g, 1, FirewallTop+1, 0, -60)
	g.Step(Input{})
	g.Step(Input{Launch: true})
	want := Diagonal.scaled(150)
	if vx, vy := g.Velocity(); vx != want.X || vy != -want.Y {
		t.Fatalf("launch velocity = (%d,%d), want (%d,%d) at 150%%", vx, vy, want.X, -want.Y)
	}
}

func TestClearingTheLastLevelBreachesTheFirewall(t *testing.T) {
	g := twoLevels()
	g.level = 1
	g.bricks = []Brick{{X: 8, Y: 3, HP: 1}}
	place(g, 9, 4, 0, -60)
	g.Step(Input{})
	if !g.Won() {
		t.Fatalf("clearing the last level must win")
	}
}

func TestLevelDropOddsAreUsed(t *testing.T) {
	g := twoLevels()
	g.level = 1
	rng := &recordingRand{}
	g.rng = rng
	g.bricks = []Brick{{X: 8, Y: 3, HP: 1}, farBrick}
	place(g, 9, 4, 0, -60)
	g.Step(Input{})
	if len(rng.asked) == 0 || rng.asked[0] != 8 {
		t.Fatalf("drop roll asked Intn(%v), want the level's odds 8", rng.asked)
	}
}

type recordingRand struct{ asked []int }

func (r *recordingRand) Intn(n int) int {
	r.asked = append(r.asked, n)
	return 1
}

// TestNoBlockIsSealedBetweenIndestructibleOnes keeps every breakable block
// reachable from below or above: a column never holds two indestructible
// blocks, which would seal the blocks between them in a pocket only a
// sideways shot through a narrow corridor can reach.
func TestNoBlockIsSealedBetweenIndestructibleOnes(t *testing.T) {
	for i, l := range Campaign {
		solidRows := map[int]int{} // column -> first row with a solid block
		for _, b := range l.Bricks() {
			if !b.Solid {
				continue
			}
			if row, seen := solidRows[b.X]; seen {
				t.Fatalf("level %d: column at x=%d has indestructible blocks on rows %d and %d", i+1, b.X, row, b.Y)
			}
			solidRows[b.X] = b.Y
		}
	}
}
