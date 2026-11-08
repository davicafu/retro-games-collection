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

// Gap sizes vary inside a range that starts at StartGap..StartGap+GapVariety-1
// rows and drops by one row every NarrowEvery antennas, until gaps are only
// MinGap or MinGap+1 rows. MinGap is the tightest gap a thrust can cross:
// one thrust climbs about 2.25 rows.
const (
	StartGap    = 4
	MinGap      = 3
	GapVariety  = 3
	NarrowEvery = 6
)

// RandomSpawner places antennas at a fixed spacing with gaps chosen by an injected Rand.
type RandomSpawner struct {
	rng     Rand
	spawned int
}

func NewRandomSpawner(rng Rand) *RandomSpawner {
	return &RandomSpawner{rng: rng}
}

func (s *RandomSpawner) Spawn(antennas []Antenna, width, height int) []Antenna {
	if len(antennas) > 0 && antennas[len(antennas)-1].X > width-AntennaSpacing {
		return antennas
	}

	stage := s.spawned / NarrowEvery
	narrowest := max(MinGap, StartGap-stage)
	widest := max(MinGap+1, StartGap+GapVariety-1-stage)
	gap := narrowest + s.rng.Intn(widest-narrowest+1)
	s.spawned++
	// maxGapTop keeps at least one mast cell below every gap.
	maxGapTop := height - gap - 1
	if len(antennas) == 0 {
		gapTop := minGapTop + s.rng.Intn(maxGapTop-minGapTop+1)
		return append(antennas, Antenna{X: width, GapTop: gapTop, Gap: gap})
	}

	last := antennas[len(antennas)-1]
	shift := s.rng.Intn(2*MaxGapShift+1) - MaxGapShift
	gapTop := min(max(last.GapTop+shift, minGapTop), maxGapTop)
	return append(antennas, Antenna{X: width, GapTop: gapTop, Gap: gap})
}
