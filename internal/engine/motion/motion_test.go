package motion

import "testing"

func TestHeadingsAreUnitVectorsTurningClockwiseFromUp(t *testing.T) {
	q := Headings / 4
	cases := map[int]Vec{0: {0, -Unit}, q: {Unit, 0}, 2 * q: {0, Unit}, 3 * q: {-Unit, 0}, 4 * q: {0, -Unit}, -q: {-Unit, 0}}
	for h, want := range cases {
		if got := Heading(h); got != want {
			t.Fatalf("Heading(%d) = %v, want %v", h, got, want)
		}
	}
	for h := 0; h < Headings; h++ {
		v := Heading(h)
		if l2 := v.X*v.X + v.Y*v.Y; l2 < (Unit-2)*(Unit-2) || l2 > (Unit+2)*(Unit+2) {
			t.Fatalf("Heading(%d) = %v has length %d², want about %d", h, v, l2, Unit)
		}
		if opp := Heading(h + Headings/2); opp != (Vec{-v.X, -v.Y}) {
			t.Fatalf("Heading(%d) = %v is not opposite to Heading(%d) = %v", h+Headings/2, opp, h, v)
		}
	}
}

func TestVecScaleAndDamp(t *testing.T) {
	if got := (Vec{1024, -512}).Scale(3, 1024); got != (Vec{3, -1}) {
		t.Fatalf("Scale = %v, want {3 -1}", got)
	}
	if got := Damp(Vec{640, -640}, 64); got != (Vec{630, -630}) {
		t.Fatalf("Damp = %v, want {630 -630}", got)
	}
	if got := Limit(Vec{300, 400}, 250); got != (Vec{150, 200}) {
		t.Fatalf("Limit = %v, want {150 200}", got)
	}
	if got := Limit(Vec{3, 4}, 250); got != (Vec{3, 4}) {
		t.Fatalf("Limit must keep a slower vector, got %v", got)
	}
}

func TestSpaceWrapsPositionsOntoTheTorus(t *testing.T) {
	s := Space{W: 100, H: 50}
	cases := map[Vec]Vec{{5, 5}: {5, 5}, {100, 50}: {0, 0}, {-1, -1}: {99, 49}, {250, -120}: {50, 30}}
	for in, want := range cases {
		if got := s.Wrap(in); got != want {
			t.Fatalf("Wrap(%v) = %v, want %v", in, got, want)
		}
	}
}

func TestSpaceMeasuresTheShortWayAcrossTheEdges(t *testing.T) {
	s := Space{W: 100, H: 50}
	if got := s.Delta(Vec{95, 2}, Vec{3, 48}); got != (Vec{8, -4}) {
		t.Fatalf("Delta = %v, want {8 -4} across both edges", got)
	}
	if got := s.Delta(Vec{10, 10}, Vec{20, 30}); got != (Vec{10, 20}) {
		t.Fatalf("Delta = %v, want {10 20}", got)
	}
	if !s.Near(Vec{98, 25}, Vec{2, 25}, 4) || s.Near(Vec{98, 25}, Vec{2, 25}, 3) {
		t.Fatalf("Near must measure 4 across the edge")
	}
}

func TestBodyMovesByItsVelocityAndWraps(t *testing.T) {
	s := Space{W: 100, H: 50}
	b := Body{Pos: Vec{98, 1}, Vel: Vec{5, -3}}
	b.Move(s)
	if b.Pos != (Vec{3, 48}) {
		t.Fatalf("Pos = %v, want {3 48}", b.Pos)
	}
}

// TestDampBringsAnyVelocityToRest: drag removes at least one unit a tick,
// so a velocity smaller than den does not drift forever.
func TestDampBringsAnyVelocityToRest(t *testing.T) {
	if got := Damp(Vec{100, -50}, 120); got != (Vec{99, -49}) {
		t.Fatalf("Damp = %v, want {99 -49}", got)
	}
	v := Vec{300, -170}
	for i := 0; i < 1000 && v != (Vec{}); i++ {
		v = Damp(v, 120)
	}
	if v != (Vec{}) {
		t.Fatalf("after 1000 ticks of drag v = %v, want at rest", v)
	}
}
