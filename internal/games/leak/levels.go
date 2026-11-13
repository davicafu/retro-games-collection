package leak

// Difficulty per level: collectors speed up, mallocs get shorter and the
// pool releases sooner, each levelling off at a playable limit.
const (
	BaseCollectorSpeed = 13 // 6.5 cells/s at level 1
	MaxCollectorSpeed  = 15 // 7.5 cells/s, still below LeakSpeed

	BaseFrightTicks = 420 // 7 s
	FrightStep      = 60
	MinFrightTicks  = 120 // 2 s

	// GC cycles: collectors sweep (chase) and pause (head for their corner)
	// in turn; every life starts with a pause.
	PauseTicks = 360 // 6 s
	SweepTicks = 900 // 15 s

	BaseReleaseTicks = 240 // 4 s between collectors leaving the pool
	ReleaseStep      = 30
	MinReleaseTicks  = 60 // 1 s
)

func (g *Game) collectorSpeed() int {
	return min(MaxCollectorSpeed, BaseCollectorSpeed+(g.level-1)/2)
}

func (g *Game) frightTicks() int {
	return max(MinFrightTicks, BaseFrightTicks-FrightStep*(g.level-1))
}

func (g *Game) releaseInterval() int {
	return max(MinReleaseTicks, BaseReleaseTicks-ReleaseStep*(g.level-1))
}

// paused reports that the collectors are in the pause of their GC cycle.
func (g *Game) paused() bool {
	return g.lifeTicks%(PauseTicks+SweepTicks) < PauseTicks
}
