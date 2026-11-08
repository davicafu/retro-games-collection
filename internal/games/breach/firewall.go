package breach

const (
	BrickWidth = 4

	// FirewallTop is the first block row; the rows above it stay open so the
	// virus can travel behind the wall once it breaks through.
	FirewallTop  = 1
	firewallRows = 5
)

// Brick is a firewall block covering columns X..X+BrickWidth-1 on row Y.
// HP is the number of hits left.
type Brick struct {
	X, Y, HP int
}

func (b Brick) covers(col, row int) bool {
	return row == b.Y && col >= b.X && col < b.X+BrickWidth
}

// Firewall returns the default wall for a playfield of the given width:
// five full rows of blocks with a hardened (two-hit) top row.
func Firewall(width int) []Brick {
	var bricks []Brick
	for row := FirewallTop; row < FirewallTop+firewallRows; row++ {
		hp := 1
		if row == FirewallTop {
			hp = 2
		}
		for x := 0; x+BrickWidth <= width; x += BrickWidth {
			bricks = append(bricks, Brick{X: x, Y: row, HP: hp})
		}
	}
	return bricks
}
