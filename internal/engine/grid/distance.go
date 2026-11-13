package grid

// Unreachable is the distance of a cell no path reaches.
const Unreachable = -1

// Neighbours are the four orthogonal steps, in a fixed order: up, left,
// down, right.
var Neighbours = [4]Point{{X: 0, Y: -1}, {X: -1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 0}}

// Add returns p moved by d.
func (p Point) Add(d Point) Point { return Point{X: p.X + d.X, Y: p.Y + d.Y} }

// DistanceField holds, for every cell of a grid, the number of orthogonal
// steps to one target cell.
type DistanceField struct {
	width, height int
	cells         []int // row by row
}

// Distances runs a breadth-first search from target over the cells of a
// width×height grid where passable holds. Chasers move toward lower
// distances to reach the target along a shortest path.
func Distances(width, height int, target Point, passable func(Point) bool) DistanceField {
	f := DistanceField{width: width, height: height, cells: make([]int, width*height)}
	for i := range f.cells {
		f.cells[i] = Unreachable
	}
	if !f.inside(target) || !passable(target) {
		return f
	}
	f.cells[f.index(target)] = 0
	queue := make([]Point, 1, width*height)
	queue[0] = target
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, d := range Neighbours {
			n := p.Add(d)
			if f.inside(n) && f.cells[f.index(n)] == Unreachable && passable(n) {
				f.cells[f.index(n)] = f.cells[f.index(p)] + 1
				queue = append(queue, n)
			}
		}
	}
	return f
}

// At returns the distance from p to the target, or Unreachable.
func (f DistanceField) At(p Point) int {
	if !f.inside(p) {
		return Unreachable
	}
	return f.cells[f.index(p)]
}

func (f DistanceField) index(p Point) int { return p.Y*f.width + p.X }

func (f DistanceField) inside(p Point) bool {
	return p.Y >= 0 && p.Y < f.height && p.X >= 0 && p.X < f.width
}
