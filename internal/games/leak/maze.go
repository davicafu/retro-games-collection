package leak

import "github.com/davidbarreiro/cli-games-collection/internal/engine/grid"

// Tile is what a map cell holds.
type Tile byte

const (
	TileFree   Tile = iota // corridor with nothing left to eat
	TileWall               // allocated block
	TileByte               // free byte to eat
	TileMalloc             // power-up that turns the collectors into prey
	TileGate               // pool gate, crossed only by collectors leaving the pool
	TilePool               // inside the collector pool
)

// RAM is the original map: '#' allocated block, '.' free byte, 'o' malloc,
// '-' pool gate, 'G' a collector's place in the pool, 'L' the leak's start
// (a corridor with no byte).
var RAM = []string{
	"###################",
	"#o.......#.......o#",
	"#.##.###.#.###.##.#",
	"#.................#",
	"#.##.#.#####.#.##.#",
	"#....#...#...#....#",
	"####.###.#.###.####",
	"####.#.......#.####",
	"####.#.##-##.#.####",
	"####.#.#GGG#.#.####",
	"####.#.#####.#.####",
	"####.#.......#.####",
	"####.###.#.###.####",
	"#....#...#...#....#",
	"#.##.#.#####.#.##.#",
	"#........L........#",
	"#.##.###.#.###.##.#",
	"#o.......#.......o#",
	"###################",
}

// Heap is the second map: long side corridors around rows of blocks.
var Heap = []string{
	"###################",
	"#o..#.........#..o#",
	"#.#.#.##.#.##.#.#.#",
	"#.................#",
	"###.#.###.###.#.###",
	"#...#.........#...#",
	"#.#.#.##.#.##.#.#.#",
	"#...#.........#...#",
	"#.#.##.##-##.##.#.#",
	"#.#.##.#GGG#.##.#.#",
	"#.#.##.#####.##.#.#",
	"#...#.........#...#",
	"#.#.#.##.#.##.#.#.#",
	"#...#.........#...#",
	"###.#.###.###.#.###",
	"#........L........#",
	"#.#.#.##.#.##.#.#.#",
	"#o..#.........#..o#",
	"###################",
}

// Stack is the third map: frames of short pillars between full-width rows.
var Stack = []string{
	"###################",
	"#o...............o#",
	"#.###.###.###.###.#",
	"#...#...#.#...#...#",
	"#.#.#.#.#.#.#.#.#.#",
	"#.#...#.....#...#.#",
	"#.#.###.#.#.###.#.#",
	"#.#.............#.#",
	"#.#.##.##-##.##.#.#",
	"#.#.##.#GGG#.##.#.#",
	"#.#.##.#####.##.#.#",
	"#.#.............#.#",
	"#.#.###.#.#.###.#.#",
	"#.#...#.....#...#.#",
	"#.#.#.#.#.#.#.#.#.#",
	"#...#...#L#...#...#",
	"#.###.###.###.###.#",
	"#o...............o#",
	"###################",
}

// Maps are played in turn, one per level.
var Maps = [][]string{RAM, Heap, Stack}

// layout is a parsed map.
type layout struct {
	tiles     [][]Tile
	leakStart grid.Point
	poolStart []grid.Point // in map order: left to right
	gate      grid.Point
}

func parseMap(rows []string) layout {
	l := layout{tiles: make([][]Tile, len(rows))}
	for y, row := range rows {
		l.tiles[y] = make([]Tile, len(row))
		for x, c := range row {
			p := grid.Point{X: x, Y: y}
			t := TileFree
			switch c {
			case '#':
				t = TileWall
			case '.':
				t = TileByte
			case 'o':
				t = TileMalloc
			case '-':
				t = TileGate
				l.gate = p
			case 'G':
				t = TilePool
				l.poolStart = append(l.poolStart, p)
			case 'L':
				l.leakStart = p
			}
			l.tiles[y][x] = t
		}
	}
	return l
}

// bannerRow is the map row for banners: the corridor just below the pool,
// or the middle row of a map without one.
func (g *Game) bannerRow() int {
	if len(g.layout.poolStart) == 0 {
		return g.height / 2
	}
	return g.layout.poolStart[0].Y + 2
}
