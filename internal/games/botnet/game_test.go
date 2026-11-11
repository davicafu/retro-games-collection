package botnet

import (
	"math/rand"
	"testing"
	"time"
)

const (
	testWidth  = 40
	testHeight = 18
	serverRow  = testHeight - 1
)

// scriptedRand returns values in order, then always n-1 (never 0, so it
// never fires unless scripted to).
type scriptedRand struct {
	values []int
	idx    int
}

func (r *scriptedRand) Intn(n int) int {
	if r.idx >= len(r.values) {
		return n - 1
	}
	v := r.values[r.idx] % n
	r.idx++
	return v
}

// newTestGame builds a game whose formation is only the given scanners and
// whose cover is only the given proxies; the RNG never fires.
func newTestGame(scanners []Scanner, cover []Proxy) *Game {
	g := NewGame(Config{Width: testWidth, Height: testHeight, Rand: &scriptedRand{}})
	g.scanners = scanners
	g.total = max(1, len(scanners))
	g.cover = cover
	return g
}

// lone is a scanner far from the server's line of fire.
var lone = Scanner{X: 0, Row: 1, Kind: KindLow}

func TestServerStartsCentredOnTheBottomRow(t *testing.T) {
	g := newTestGame([]Scanner{lone}, nil)
	x, w := g.Server()
	if w != ServerWidth || x != (testWidth-ServerWidth)/2 {
		t.Fatalf("server at %d width %d, want centred width %d", x, w, ServerWidth)
	}
}

func TestServerGlidesAndStopsAtTheWalls(t *testing.T) {
	g := newTestGame([]Scanner{lone}, nil)
	start, _ := g.Server()
	g.Step(Input{Move: 1})
	if x, _ := g.Server(); x != start+1 {
		t.Fatalf("server x = %d, want %d on the press tick", x, start+1)
	}
	for i := 0; i < 200; i++ {
		g.Step(Input{Move: 1})
	}
	if x, w := g.Server(); x != testWidth-w {
		t.Fatalf("server x = %d, want stopped at %d", x, testWidth-w)
	}
	for i := 0; i < 200; i++ {
		g.Step(Input{Move: -1})
	}
	if x, _ := g.Server(); x != 0 {
		t.Fatalf("server x = %d, want stopped at 0", x)
	}
}

func TestFireLaunchesOnePacketFromTheServerCentre(t *testing.T) {
	g := newTestGame([]Scanner{lone}, nil)
	x, w := g.Server()
	g.Step(Input{Fire: true})
	col, row, ok := g.Packet()
	if !ok || col != x+w/2 || row != serverRow-1 {
		t.Fatalf("packet = (%d,%d,%v), want (%d,%d) in flight", col, row, ok, x+w/2, serverRow-1)
	}
	_, y0 := g.packetCol, g.packetY
	g.Step(Input{Fire: true})
	if g.packetY != y0-PacketSpeed {
		t.Fatalf("second fire must not replace the packet in flight: y = %d, want %d", g.packetY, y0-PacketSpeed)
	}
}

func TestPacketLeavesAtTheTopAndCanBeFiredAgain(t *testing.T) {
	g := newTestGame([]Scanner{lone}, nil)
	g.Step(Input{Fire: true})
	for i := 0; i < testHeight*SubCells/PacketSpeed+2; i++ {
		g.Step(Input{})
	}
	if _, _, ok := g.Packet(); ok {
		t.Fatalf("packet must leave the playfield at the top")
	}
	g.Step(Input{Fire: true})
	if _, _, ok := g.Packet(); !ok {
		t.Fatalf("a new packet must fire once the last one is gone")
	}
}

func TestPacketDestroysTheScannerItReachesAndScoresByKind(t *testing.T) {
	for _, kind := range []Kind{KindHigh, KindMid, KindLow} {
		x := (testWidth - ServerWidth) / 2
		target := Scanner{X: x, Row: 10, Kind: kind}
		g := newTestGame([]Scanner{target, lone}, nil)
		g.Step(Input{Fire: true})
		for i := 0; i < 100 && len(g.Scanners()) == 2; i++ {
			g.Step(Input{})
		}
		if got := g.Scanners(); len(got) != 1 || got[0] != lone {
			t.Fatalf("kind %d: scanners = %+v, want only the lone one left", kind, got)
		}
		if _, _, ok := g.Packet(); ok {
			t.Fatalf("kind %d: the packet must be spent on the hit", kind)
		}
		if g.Score() != kind.Points() {
			t.Fatalf("kind %d: score = %d, want %d", kind, g.Score(), kind.Points())
		}
	}
}

func TestPointsByKind(t *testing.T) {
	if KindHigh.Points() != 30 || KindMid.Points() != 20 || KindLow.Points() != 10 {
		t.Fatalf("points = %d/%d/%d, want 30/20/10", KindHigh.Points(), KindMid.Points(), KindLow.Points())
	}
}

func TestScannerEdgesAreExact(t *testing.T) {
	s := Scanner{X: 10, Row: 4}
	cases := map[[2]int]bool{{9, 4}: false, {10, 4}: true, {10 + ScannerWidth - 1, 4}: true, {10 + ScannerWidth, 4}: false, {10, 3}: false, {10, 5}: false}
	for cell, want := range cases {
		if got := s.covers(cell[0], cell[1]); got != want {
			t.Fatalf("covers(%d,%d) = %v, want %v", cell[0], cell[1], got, want)
		}
	}
}

func TestFormationStepsSidewaysThenDropsAndReversesAtTheWall(t *testing.T) {
	s := Scanner{X: testWidth - ScannerWidth - 1, Row: 3, Kind: KindLow}
	g := newTestGame([]Scanner{s, lone}, nil)
	g.formationStep()
	if got := g.Scanners()[0]; got.X != s.X+1 || got.Row != 3 {
		t.Fatalf("after one step = %+v, want one cell right", got)
	}
	g.formationStep()
	if got := g.Scanners()[0]; got.X != s.X+1 || got.Row != 4 {
		t.Fatalf("at the wall = %+v, want dropped one row in place", got)
	}
	g.formationStep()
	if got := g.Scanners()[0]; got.X != s.X || got.Row != 4 {
		t.Fatalf("after reversing = %+v, want one cell left", got)
	}
}

func TestFormationStepsFasterAsScannersFall(t *testing.T) {
	g := newTestGame(make([]Scanner, 32), nil)
	full := g.stepInterval()
	g.scanners = g.scanners[:8]
	quarter := g.stepInterval()
	g.scanners = g.scanners[:1]
	last := g.stepInterval()
	if !(full > quarter && quarter > last && last >= MinStepTicks) {
		t.Fatalf("step intervals = %d, %d, %d, want strictly shorter and at least %d", full, quarter, last, MinStepTicks)
	}
}

func TestFormationReachingTheServerRowEndsTheRound(t *testing.T) {
	g := newTestGame([]Scanner{{X: testWidth - ScannerWidth, Row: serverRow - 1, Kind: KindLow}}, nil)
	g.formationStep()
	if !g.Over() {
		t.Fatalf("a scanner on the server row must end the round")
	}
}

func TestScannersEraseTheCoverTheyMoveInto(t *testing.T) {
	g := newTestGame([]Scanner{{X: 10, Row: 13, Kind: KindLow}, lone}, []Proxy{{X: 12, Row: 13, HP: 2}, {X: 20, Row: 13, HP: 2}})
	g.formationStep()
	if got := g.Cover(); len(got) != 1 || got[0].X != 20 {
		t.Fatalf("cover = %+v, want the proxy under the scanner erased", got)
	}
}

func TestPacketErodesCover(t *testing.T) {
	x := (testWidth - ServerWidth) / 2
	g := newTestGame([]Scanner{lone}, []Proxy{{X: x + 1, Row: 14, HP: 2}})
	for shot := 1; shot <= 2; shot++ {
		g.Step(Input{Fire: true})
		for i := 0; i < 100; i++ {
			if _, _, ok := g.Packet(); !ok {
				break
			}
			g.Step(Input{})
		}
		if want := 2 - shot; want > 0 && (len(g.Cover()) != 1 || g.Cover()[0].HP != want) {
			t.Fatalf("after shot %d: cover = %+v, want HP %d", shot, g.Cover(), want)
		}
	}
	if len(g.Cover()) != 0 {
		t.Fatalf("cover = %+v, want eroded away", g.Cover())
	}
}

func TestQuarantineShotHittingTheServerCostsALifeAndClearsShots(t *testing.T) {
	x, _ := newTestGame(nil, nil).Server()
	g := newTestGame([]Scanner{lone}, nil)
	g.shots = []Shot{{Col: x, Y: (serverRow-1)*SubCells + SubCells - 1}, {Col: 0, Y: 5 * SubCells}}
	g.Step(Input{})
	if g.Lives() != StartLives-1 || len(g.Shots()) != 0 {
		t.Fatalf("lives = %d shots = %+v, want one life lost and every shot cleared", g.Lives(), g.Shots())
	}
}

func TestQuarantineShotMissingTheServerLeavesThePlayfield(t *testing.T) {
	g := newTestGame([]Scanner{lone}, nil)
	g.shots = []Shot{{Col: 0, Y: serverRow*SubCells + SubCells - 1}}
	g.Step(Input{})
	if len(g.Shots()) != 0 || g.Lives() != StartLives {
		t.Fatalf("shots = %+v lives = %d, want the shot gone and no life lost", g.Shots(), g.Lives())
	}
}

func TestQuarantineShotErodesCover(t *testing.T) {
	g := newTestGame([]Scanner{lone}, []Proxy{{X: 5, Row: 14, HP: 2}})
	g.shots = []Shot{{Col: 5, Y: 13*SubCells + SubCells - 1}}
	g.Step(Input{})
	if len(g.Shots()) != 0 || len(g.Cover()) != 1 || g.Cover()[0].HP != 1 {
		t.Fatalf("shots = %+v cover = %+v, want the shot spent on the proxy", g.Shots(), g.Cover())
	}
}

func TestLosingTheLastLifeEndsTheRound(t *testing.T) {
	x, _ := newTestGame(nil, nil).Server()
	g := newTestGame([]Scanner{lone}, nil)
	for i := 0; i < StartLives; i++ {
		g.shots = []Shot{{Col: x, Y: (serverRow-1)*SubCells + SubCells - 1}}
		g.Step(Input{})
	}
	if !g.Over() || g.Lives() != 0 {
		t.Fatalf("over = %v lives = %d, want the round over", g.Over(), g.Lives())
	}
	before := g.Score()
	g.Step(Input{Fire: true})
	if _, _, ok := g.Packet(); ok || g.Score() != before {
		t.Fatalf("an over round must not advance")
	}
}

func TestLowestScannerOfARandomColumnFires(t *testing.T) {
	top := Scanner{X: 8, Row: 2, Kind: KindHigh}
	bottom := Scanner{X: 8, Row: 6, Kind: KindLow}
	other := Scanner{X: 20, Row: 2, Kind: KindHigh}
	// Intn(odds) = 0 fires; Intn(len(scanners)) = 0 picks the top one of column 8.
	g := NewGame(Config{Width: testWidth, Height: testHeight, Rand: &scriptedRand{values: []int{0, 0}}})
	g.scanners, g.total = []Scanner{top, other, bottom}, 3
	g.Step(Input{})
	shots := g.Shots()
	if len(shots) != 1 || shots[0].Col != 8+ScannerWidth/2 || shots[0].Y/SubCells != 7 {
		t.Fatalf("shots = %+v, want one shot below the lowest scanner of column 8", shots)
	}
}

func TestEnemyFireRespectsTheWaveCap(t *testing.T) {
	g := NewGame(Config{Width: testWidth, Height: testHeight, Rand: &scriptedRand{values: make([]int, 100)}})
	for i := 0; i < 20; i++ {
		g.Step(Input{})
	}
	if got := len(g.Shots()); got != g.shotCap() {
		t.Fatalf("shots = %d, want capped at %d", got, g.shotCap())
	}
}

func TestStartingFormationIsFourRowsOfEight(t *testing.T) {
	g := NewGame(Config{Width: testWidth, Height: testHeight})
	scanners := g.Scanners()
	if len(scanners) != FormationRows*FormationCols || FormationRows != 4 || FormationCols != 8 {
		t.Fatalf("formation = %d scanners, want 4 rows of 8", len(scanners))
	}
	kinds := map[Kind]int{}
	for _, s := range scanners {
		kinds[s.Kind]++
		if s.X < 0 || s.X+ScannerWidth > testWidth {
			t.Fatalf("scanner %+v outside the playfield", s)
		}
	}
	if kinds[KindHigh] != 8 || kinds[KindMid] != 16 || kinds[KindLow] != 8 {
		t.Fatalf("kinds = %v, want 8 high, 16 mid, 8 low", kinds)
	}
	if len(g.Cover()) == 0 {
		t.Fatalf("the round must start with cover")
	}
}

func TestClearingAWaveStartsAHarderOneKeepingScoreLivesAndCover(t *testing.T) {
	x := (testWidth - ServerWidth) / 2
	g := newTestGame([]Scanner{{X: x, Row: 10, Kind: KindHigh}}, []Proxy{{X: 0, Row: 14, HP: 1}})
	g.lives = 1
	interval1, odds1, cap1, top1 := g.stepInterval(), g.fireOdds(), g.shotCap(), formationTop(1)
	g.Step(Input{Fire: true})
	for i := 0; i < 100 && g.Wave() == 1; i++ {
		g.Step(Input{})
	}
	if g.Wave() != 2 || len(g.Scanners()) != FormationRows*FormationCols {
		t.Fatalf("wave = %d scanners = %d, want wave 2 with a full formation", g.Wave(), len(g.Scanners()))
	}
	if g.Score() != KindHigh.Points() || g.Lives() != 1 || len(g.Cover()) != 1 {
		t.Fatalf("score = %d lives = %d cover = %+v, want them carried over", g.Score(), g.Lives(), g.Cover())
	}
	if !(g.stepInterval() < interval1 && g.fireOdds() < odds1 && g.shotCap() > cap1 && formationTop(2) > top1) {
		t.Fatalf("wave 2 must step faster, fire more often, allow more shots and start lower")
	}
	if g.Scanners()[0].Row != formationTop(2) {
		t.Fatalf("wave 2 top row = %d, want %d", g.Scanners()[0].Row, formationTop(2))
	}
}

func TestWaveDifficultyLevelsOff(t *testing.T) {
	for wave := 1; wave < 30; wave++ {
		if formationTop(wave) > MaxFormationTop || waveStepTicks(wave) < MinWaveStepTicks || fireOdds(wave) < MinFireOdds || shotCap(wave) > MaxShotCap {
			t.Fatalf("wave %d beyond the difficulty limits", wave)
		}
	}
}

// TestRealTimeSpeedsDoNotDependOnTheTickRate pins speeds per second, so a
// change of TickPeriod must rescale the per-tick constants with it.
func TestRealTimeSpeedsDoNotDependOnTheTickRate(t *testing.T) {
	if TickPeriod != time.Second/60 {
		t.Fatalf("tick period = %v, want 60 ticks per second", TickPeriod)
	}
	perSecond := int(time.Second / TickPeriod)
	if got := PacketSpeed * perSecond / SubCells; got != 30 {
		t.Fatalf("packets fly %d cells/s, want 30", got)
	}
	if got := ShotSpeed * perSecond / SubCells; got != 12 {
		t.Fatalf("quarantine shots fall %d cells/s, want 12", got)
	}
	if PacketSpeed >= SubCells || ShotSpeed >= SubCells {
		t.Fatalf("shots must move less than a cell per tick")
	}
	if got := (time.Duration(waveStepTicks(1)) * TickPeriod).Round(time.Millisecond); got != 600*time.Millisecond {
		t.Fatalf("wave 1 formation steps every %v, want 600ms", got)
	}
}

// ViableWaves is how many waves the autopilot must clear, and SafetyRows how
// far above the server row the formation must stay while it does: every
// wave must be clearable with room to spare.
const (
	ViableWaves = 10
	SafetyRows  = 3
)

// TestAutopilotClearsWavesBeforeTheyLand plays the real formation: a server
// that fires whenever a packet fired now would hit a scanner, and otherwise
// heads for the nearest column (drilling through cover in its way), must
// clear the first ViableWaves waves while the formation stays SafetyRows
// above it. Lives are topped up: this checks that the formation's pace
// leaves time to clear it, not the autopilot's dodging.
func TestAutopilotClearsWavesBeforeTheyLand(t *testing.T) {
	for seed := int64(1); seed <= 3; seed++ {
		g := NewGame(Config{Width: testWidth, Height: testHeight, Rand: rand.New(rand.NewSource(seed))})
		for tick := 0; tick < 400000 && g.Wave() <= ViableWaves && !g.Over(); tick++ {
			g.lives = StartLives
			x, w := g.Server()
			target := nearestColumn(g, x+w/2)
			move := 0
			switch {
			case target < x+w/2:
				move = -1
			case target > x+w/2:
				move = 1
			}
			outcome := fireOutcome(g, move)
			fire := outcome == hitScanner || (outcome == hitCover && move == 0)
			g.Step(Input{Move: move, Fire: fire})
			for _, s := range g.Scanners() {
				if s.Row > serverRow-SafetyRows {
					t.Fatalf("seed %d wave %d: scanner reached row %d, want at most %d", seed, g.Wave(), s.Row, serverRow-SafetyRows)
				}
			}
		}
		if g.Over() || g.Wave() <= ViableWaves {
			t.Fatalf("seed %d: over = %v at wave %d with %d scanners left", seed, g.Over(), g.Wave(), len(g.Scanners()))
		}
	}
}

type outcome int

const (
	miss outcome = iota
	hitScanner
	hitCover
)

// fireOutcome plays a copy of g, without enemy fire, to see what a packet
// fired now, while making this move, would hit.
func fireOutcome(g *Game, move int) outcome {
	if g.packet {
		return miss
	}
	c := *g
	c.rng = nil
	c.scanners = g.Scanners()
	c.cover = g.Cover()
	c.shots = nil
	scanners, cover := len(c.scanners), coverHP(c.cover)
	c.Step(Input{Move: move, Fire: true})
	for c.packet && !c.lost {
		c.Step(Input{})
	}
	switch {
	case c.score > g.score:
		return hitScanner
	case coverHP(c.cover) < cover && len(c.scanners) == scanners:
		return hitCover
	}
	return miss
}

func coverHP(cover []Proxy) int {
	hp := 0
	for _, p := range cover {
		hp += p.HP
	}
	return hp
}

// nearestColumn returns the centre column of the scanner closest to col.
func nearestColumn(g *Game, col int) int {
	best := -1
	for _, s := range g.Scanners() {
		c := s.X + ScannerWidth/2
		if best < 0 || abs(c-col) < abs(best-col) {
			best = c
		}
	}
	return best
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
