package runner

// Spawner adds new obstacles at the right edge as the world scrolls.
// It receives the current obstacles, the playfield width and the current speed,
// and returns the updated obstacle list.
type Spawner interface {
	Spawn(obstacles []Obstacle, width, speed int) []Obstacle
}

// Rand is the randomness the spawner needs; *math/rand.Rand satisfies it.
type Rand interface {
	Intn(n int) int
}

// obstacleShapes lists every obstacle the spawner can create. All of them are
// clearable: ground ones by jumping, flying ones by ducking.
var obstacleShapes = []Obstacle{
	{Width: 1, Tall: 1, Kind: KindGround},
	{Width: 2, Tall: 1, Kind: KindGround},
	{Width: 1, Tall: 2, Kind: KindGround},
	{Width: 2, Kind: KindFlying},
	{Width: 3, Kind: KindFlying},
}

const (
	gapMargin   = 2  // cells to land and react after a full jump
	maxExtraGap = 12 // random extra spacing, exclusive upper bound
)

// MinGap is the smallest free distance between two obstacles at the given speed:
// the cells a full jump covers plus a reaction margin, so the player can always
// land before the next obstacle needs an action.
func MinGap(speed int) int {
	jumpCells := (len(jumpHeights)*speed + scrollUnit - 1) / scrollUnit
	return jumpCells + gapMargin
}

// RandomSpawner picks obstacle shapes and extra spacing from an injected Rand.
type RandomSpawner struct {
	rng      Rand
	extraGap int
}

func NewRandomSpawner(rng Rand) *RandomSpawner {
	return &RandomSpawner{rng: rng}
}

func (s *RandomSpawner) Spawn(obstacles []Obstacle, width, speed int) []Obstacle {
	if len(obstacles) > 0 {
		last := obstacles[len(obstacles)-1]
		if width-(last.X+last.Width) < MinGap(speed)+s.extraGap {
			return obstacles
		}
	}

	shape := obstacleShapes[s.rng.Intn(len(obstacleShapes))]
	shape.X = width
	s.extraGap = s.rng.Intn(maxExtraGap)
	return append(obstacles, shape)
}
