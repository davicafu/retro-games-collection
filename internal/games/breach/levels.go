package breach

const (
	BrickWidth = 4

	// FirewallTop is the first block row; the row above it stays open so the
	// virus can travel behind the wall once it breaks through.
	FirewallTop = 1
)

// Brick is a firewall block covering columns X..X+BrickWidth-1 on row Y.
// HP is the number of hits left; a Solid block never breaks and does not
// count for clearing the level.
type Brick struct {
	X, Y, HP int
	Solid    bool
}

func (b Brick) covers(col, row int) bool {
	return row == b.Y && col >= b.X && col < b.X+BrickWidth
}

// Level is one firewall to break through.
type Level struct {
	// Wall has one string per block row, from FirewallTop down, and one
	// character per block: '.' empty, '1'-'3' hits to break, '#' indestructible.
	Wall []string
	// Speed scales the exit angles, in percent.
	Speed int
	// DropOdds: one destroyed block in DropOdds drops a power-up.
	DropOdds int
}

// Bricks builds the level's blocks from its wall.
func (l Level) Bricks() []Brick {
	var bricks []Brick
	for r, row := range l.Wall {
		for i, c := range row {
			b := Brick{X: i * BrickWidth, Y: FirewallTop + r}
			switch {
			case c == '#':
				b.Solid = true
			case c >= '1' && c <= '3':
				b.HP = int(c - '0')
			default:
				continue
			}
			bricks = append(bricks, b)
		}
	}
	return bricks
}

// Campaign is the default run of levels for the 40-cell playfield. Every
// level is harder than the one before: more total hits to clear, a faster
// virus and rarer power-ups; 3-hit blocks arrive at level 5 and
// indestructible ones at level 6.
var Campaign = []Level{
	{Speed: 100, DropOdds: 4, Wall: []string{
		"1111111111",
		"1111111111",
		"1111111111",
	}},
	{Speed: 106, DropOdds: 4, Wall: []string{
		"2222222222",
		"1111111111",
		"1111111111",
		"1111111111",
	}},
	{Speed: 112, DropOdds: 5, Wall: []string{
		"2222222222",
		"1111111111",
		"1111111111",
		"1111111111",
		"1111111111",
	}},
	{Speed: 118, DropOdds: 5, Wall: []string{
		"2222222222",
		"2221111222",
		"2211111122",
		"2221111222",
		"2222222222",
	}},
	{Speed: 124, DropOdds: 6, Wall: []string{
		"3333333333",
		"2222222222",
		"2111111112",
		"1111111111",
		"1111111111",
		"1111111111",
	}},
	{Speed: 130, DropOdds: 6, Wall: []string{
		"3333333333",
		"2#222222#2",
		"2222222222",
		"2211111122",
		"1111111111",
		"1111111111",
	}},
	{Speed: 136, DropOdds: 7, Wall: []string{
		"3333333333",
		"3#3#33#3#3",
		"2222222222",
		"2222222222",
		"1111111111",
		"1111111111",
		"1111111111",
	}},
	{Speed: 142, DropOdds: 8, Wall: []string{
		"3333333333",
		"3333333333",
		"#2#2##2#2#",
		"2222222222",
		"2222222222",
		"2222222222",
		"1111111111",
	}},
	{Speed: 148, DropOdds: 9, Wall: []string{
		"3333333333",
		"3333333333",
		"3##3333##3",
		"2222222222",
		"2222222222",
		"#222##222#",
		"2222222222",
		"1111111111",
	}},
	{Speed: 154, DropOdds: 10, Wall: []string{
		"3333333333",
		"3333333333",
		"3333333333",
		"#3#3##3#3#",
		"3333333333",
		"2222222222",
		"2#2#22#2#2",
		"2222222222",
	}},
}
