package drone

// Spawner adds new antennas at the right edge as the world scrolls. It
// receives the current antennas, the playfield width and the sky height.
type Spawner interface {
	Spawn(antennas []Antenna, width, height int) []Antenna
}

// Rand is the randomness the spawner needs; *math/rand.Rand satisfies it.
type Rand interface {
	Intn(n int) int
}

const (
	// AntennaSpacing is the distance between the left columns of two
	// consecutive antennas.
	AntennaSpacing = 16

	// MaxGapShift bounds how many rows a gap moves from the previous one. At
	// MaxSpeed the free columns between antennas still leave enough ticks to
	// climb or fall that far.
	MaxGapShift = 3

	// minGapTop keeps at least one mast cell above every gap.
	minGapTop = 1
)

// RandomSpawner places antennas at a fixed spacing with gaps chosen by an injected Rand.
type RandomSpawner struct {
	rng Rand
}

func NewRandomSpawner(rng Rand) *RandomSpawner {
	return &RandomSpawner{rng: rng}
}

func (s *RandomSpawner) Spawn(antennas []Antenna, width, height int) []Antenna {
	// maxGapTop keeps at least one mast cell below every gap.
	maxGapTop := height - GapSize - 1
	if len(antennas) == 0 {
		gapTop := minGapTop + s.rng.Intn(maxGapTop-minGapTop+1)
		return append(antennas, Antenna{X: width, GapTop: gapTop})
	}

	last := antennas[len(antennas)-1]
	if last.X > width-AntennaSpacing {
		return antennas
	}
	shift := s.rng.Intn(2*MaxGapShift+1) - MaxGapShift
	gapTop := min(max(last.GapTop+shift, minGapTop), maxGapTop)
	return append(antennas, Antenna{X: width, GapTop: gapTop})
}
