package debris

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/motion"
)

// Viability (S8): an autopilot that makes one move (a turn, a shot or a
// thrust) every ActTicks clears the first ViableWaves waves with its
// StartLives lives in at least ViableSeeds of AutopilotSeeds games.
const (
	ViableWaves    = 3
	AutopilotSeeds = 30
	ViableSeeds    = 27
	// ActTicks is the autopilot's pace: 15 moves a second, about a held
	// key's repeat rate.
	ActTicks = 4
)

func TestAutopilotClearsTheFirstWaves(t *testing.T) {
	won := make([]bool, AutopilotSeeds)
	t.Run("games", func(t *testing.T) {
		for i := range won {
			t.Run(fmt.Sprintf("seed %d", i+1), func(t *testing.T) {
				t.Parallel()
				g := NewGame(Config{Cols: playfieldCols, Rows: playfieldRows, Rand: rand.New(rand.NewSource(int64(i + 1)))})
				for tick := 0; tick < 60*60*10 && !g.Over() && g.Wave() <= ViableWaves; tick++ {
					var in Input
					if tick%ActTicks == 0 {
						in = autopilot(g)
					}
					g.Step(in)
				}
				won[i] = g.Wave() > ViableWaves
			})
		}
	})
	cleared := 0
	for _, w := range won {
		if w {
			cleared++
		}
	}
	if cleared < ViableSeeds {
		t.Fatalf("autopilot cleared %d waves in %d of %d games, want at least %d", ViableWaves, cleared, AutopilotSeeds, ViableSeeds)
	}
	t.Logf("autopilot cleared %d waves in %d of %d games", ViableWaves, cleared, AutopilotSeeds)
}

// Autopilot tuning: how far ahead it looks for a piece on a collision
// course, and the extra room it wants around the drone.
const (
	lookAhead = 45
	margin    = 1500
)

// autopilot makes one move. When a piece is on a collision course it turns
// to shoot it or to escape, whichever needs fewer turns, and shoots or
// thrusts away; otherwise it turns towards where the nearest piece will be,
// closes in when it is out of range, and fires.
func autopilot(g *Game) Input {
	d := g.Drone()
	space := g.Space()
	if threat, ok := closestThreat(g); ok && !g.Shielded() {
		toward := space.Delta(d.Pos, threat.Pos)
		aim := leadHeading(g, threat)
		if aim == d.Heading {
			return Input{Fire: true}
		}
		away := headingOf(motion.Vec{X: -toward.X, Y: -toward.Y})
		// Shooting needs the shot to fly; escaping needs a turn and a burst.
		if turnsBetween(d.Heading, aim) <= turnsBetween(d.Heading, away) {
			return Input{Turn: turnToward(d.Heading, aim)}
		}
		if turnsBetween(d.Heading, away) <= 1 {
			return Input{Thrust: true}
		}
		return Input{Turn: turnToward(d.Heading, away)}
	}
	var target Piece
	best := -1
	for _, p := range g.Pieces() {
		if l := space.Delta(d.Pos, p.Pos).Len(); best < 0 || l < best {
			target, best = p, l
		}
	}
	if best < 0 {
		return Input{}
	}
	aim := leadHeading(g, target)
	if aim != d.Heading {
		return Input{Turn: turnToward(d.Heading, aim)}
	}
	// Out of range, close in first, slowly.
	if best > ShotSpeed*ShotTicks && d.Vel.Len() < MaxDroneSpeed/3 {
		return Input{Thrust: true}
	}
	return Input{Fire: true}
}

// leadHeading aims at where p will be when a shot reaches it.
func leadHeading(g *Game, p Piece) int {
	d := g.Drone()
	space := g.Space()
	flight := space.Delta(d.Pos, p.Pos).Len() / ShotSpeed
	return headingOf(space.Delta(d.Pos, p.Pos.Add(p.Vel.Scale(flight, 1))))
}

// closestThreat returns the piece whose path comes nearest to the drone's
// within lookAhead ticks, if it comes within reach.
func closestThreat(g *Game) (Piece, bool) {
	d := g.Drone()
	space := g.Space()
	var threat Piece
	found, soonest := false, lookAhead+1
	for _, p := range g.Pieces() {
		rel := space.Delta(d.Pos, p.Pos)
		vel := motion.Vec{X: p.Vel.X - d.Vel.X, Y: p.Vel.Y - d.Vel.Y}
		reach := radius[p.Size] + droneRadius + margin
		for t := 0; t <= lookAhead && t < soonest; t++ {
			at := rel.Add(vel.Scale(t, 1))
			if at.X*at.X+at.Y*at.Y <= reach*reach {
				threat, found, soonest = p, true, t
				break
			}
		}
	}
	return threat, found
}

// headingOf is the heading closest to the direction of v.
func headingOf(v motion.Vec) int {
	best, bestDot := 0, 0
	for h := 0; h < motion.Headings; h++ {
		u := motion.Heading(h)
		if dot := u.X*v.X + u.Y*v.Y; h == 0 || dot > bestDot {
			best, bestDot = h, dot
		}
	}
	return best
}

// turnsBetween is how many single turns separate two headings.
func turnsBetween(a, b int) int {
	d := ((b-a)%motion.Headings + motion.Headings) % motion.Headings
	return min(d, motion.Headings-d)
}

// turnToward is the single turn that brings heading a closer to b.
func turnToward(a, b int) int {
	if ((b-a)%motion.Headings+motion.Headings)%motion.Headings <= motion.Headings/2 {
		return 1
	}
	return -1
}
