// Package motion is integer inertial motion: vectors, 16 headings and a
// space whose edges wrap around (a torus). There is no floating point, so
// simulations stay deterministic on every platform.
package motion

// Vec is a position, velocity or direction in a game's integer units.
type Vec struct{ X, Y int }

func (v Vec) Add(o Vec) Vec { return Vec{v.X + o.X, v.Y + o.Y} }

// Scale returns v*num/den, each component rounded towards zero.
func (v Vec) Scale(num, den int) Vec { return Vec{v.X * num / den, v.Y * num / den} }

// Len is the length of v, rounded down.
func (v Vec) Len() int { return isqrt(v.X*v.X + v.Y*v.Y) }

// Damp is drag: it removes 1/den of v.
func Damp(v Vec, den int) Vec { return Vec{v.X - v.X/den, v.Y - v.Y/den} }

// Limit shortens v to at most max, keeping its direction.
func Limit(v Vec, max int) Vec {
	if l := v.Len(); l > max {
		return v.Scale(max, l)
	}
	return v
}

func isqrt(n int) int {
	if n < 2 {
		return n
	}
	x := n
	for y := (x + 1) / 2; y < x; y = (x + n/x) / 2 {
		x = y
	}
	return x
}

// Headings is the number of directions a heading can take; Unit is the
// length of a heading's vector.
const (
	Headings = 32
	Unit     = 1024
)

// headings are the unit vectors, scaled by Unit, from up (0) turning
// clockwise in steps of 11.25 degrees; y grows downwards.
var headings = [Headings]Vec{
	{0, -1024}, {200, -1004}, {392, -946}, {569, -851},
	{724, -724}, {851, -569}, {946, -392}, {1004, -200},
	{1024, 0}, {1004, 200}, {946, 392}, {851, 569},
	{724, 724}, {569, 851}, {392, 946}, {200, 1004},
	{0, 1024}, {-200, 1004}, {-392, 946}, {-569, 851},
	{-724, 724}, {-851, 569}, {-946, 392}, {-1004, 200},
	{-1024, 0}, {-1004, -200}, {-946, -392}, {-851, -569},
	{-724, -724}, {-569, -851}, {-392, -946}, {-200, -1004},
}

// Heading returns the unit vector of heading h, taken modulo Headings.
func Heading(h int) Vec { return headings[((h%Headings)+Headings)%Headings] }

// Space is a W by H torus: what leaves one edge comes back at the opposite one.
type Space struct{ W, H int }

// Wrap brings p into the space.
func (s Space) Wrap(p Vec) Vec { return Vec{mod(p.X, s.W), mod(p.Y, s.H)} }

// Delta is the shortest displacement from a to b, across the edges if
// that is shorter.
func (s Space) Delta(a, b Vec) Vec { return Vec{short(b.X-a.X, s.W), short(b.Y-a.Y, s.H)} }

// Near reports whether a and b are at most r apart.
func (s Space) Near(a, b Vec, r int) bool {
	d := s.Delta(a, b)
	return d.X*d.X+d.Y*d.Y <= r*r
}

func mod(a, n int) int { return ((a % n) + n) % n }

// short maps a difference to the shortest one modulo n, in (-n/2, n/2].
func short(d, n int) int {
	d = mod(d, n)
	if d > n/2 {
		d -= n
	}
	return d
}

// Body is something that drifts: each Move adds its velocity to its position.
type Body struct {
	Pos, Vel Vec
}

// Move advances b one tick and wraps it into s.
func (b *Body) Move(s Space) { b.Pos = s.Wrap(b.Pos.Add(b.Vel)) }
