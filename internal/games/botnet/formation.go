package botnet

// Kind is a scanner's rank: higher rows are worth more.
type Kind int

const (
	KindLow Kind = iota
	KindMid
	KindHigh
)

// Points is what destroying a scanner of this kind scores.
func (k Kind) Points() int { return [...]int{10, 20, 30}[k] }

const (
	ScannerWidth   = 3
	scannerSpacing = 1

	FormationRows = 4
	FormationCols = 8
	// formationRowGap is the number of rows from one formation row to the next.
	formationRowGap = 2
)

// Wave difficulty: each wave starts lower, steps faster and fires more,
// up to these limits.
const (
	MaxFormationTop  = 4
	MinWaveStepTicks = 12
	MinFireOdds      = 15
	MaxShotCap       = 5
	// MinStepTicks is the fastest the formation steps, with one scanner left.
	MinStepTicks = 2
)

// Scanner is one antivirus scanner covering columns X..X+ScannerWidth-1 of its row.
type Scanner struct {
	X, Row int
	Kind   Kind
}

func (s Scanner) covers(col, row int) bool {
	return row == s.Row && col >= s.X && col < s.X+ScannerWidth
}

func formationTop(wave int) int { return min(wave, MaxFormationTop) }

// waveStepTicks is how often a full formation steps in this wave.
func waveStepTicks(wave int) int { return max(MinWaveStepTicks, 30-3*(wave-1)) }

// fireOdds is the 1-in-n chance per tick that a scanner fires.
func fireOdds(wave int) int { return max(MinFireOdds, 60-8*(wave-1)) }

// shotCap is the most quarantine shots on screen at once.
func shotCap(wave int) int { return min(MaxShotCap, 1+wave) }

func (g *Game) fireOdds() int { return fireOdds(g.wave) }
func (g *Game) shotCap() int  { return shotCap(g.wave) }

// stepInterval shrinks with the scanners left, so the formation speeds up.
func (g *Game) stepInterval() int {
	return max(MinStepTicks, waveStepTicks(g.wave)*len(g.scanners)/g.total)
}

// startWave builds a fresh formation; score, lives and cover carry over.
func (g *Game) startWave(wave int) {
	g.wave = wave
	g.waveTicks = 0
	g.scanners = nil
	stride := ScannerWidth + scannerSpacing
	left := (g.width - (FormationCols*stride - scannerSpacing)) / 2
	for r := 0; r < FormationRows; r++ {
		kind := KindMid
		switch r {
		case 0:
			kind = KindHigh
		case FormationRows - 1:
			kind = KindLow
		}
		for c := 0; c < FormationCols; c++ {
			g.scanners = append(g.scanners, Scanner{X: left + c*stride, Row: formationTop(wave) + r*formationRowGap, Kind: kind})
		}
	}
	g.total = len(g.scanners)
	g.dir = 1
	g.stepTimer = 0
	g.packet = false
	g.shots = nil
}

// formationStep moves every scanner one cell sideways; when one would leave
// the playfield, the formation drops one row instead and reverses.
func (g *Game) formationStep() {
	g.steps++
	drop := false
	for _, s := range g.scanners {
		if x := s.X + g.dir; x < 0 || x+ScannerWidth > g.width {
			drop = true
			break
		}
	}
	for i := range g.scanners {
		if drop {
			g.scanners[i].Row++
		} else {
			g.scanners[i].X += g.dir
		}
	}
	if drop {
		g.dir = -g.dir
	}
	for _, s := range g.scanners {
		for col := s.X; col < s.X+ScannerWidth; col++ {
			g.removeProxy(col, s.Row)
		}
		if s.Row >= g.serverRow() {
			g.lost = true
		}
	}
}
