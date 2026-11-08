package botnet

// Proxy is one cell of cover; any shot that reaches it erodes one HP.
type Proxy struct {
	X, Row int
	HP     int
}

const (
	ProxyHP     = 2
	proxyBlocks = 3
	proxyWidth  = 4
	proxyHeight = 2
)

// startingCover places proxyBlocks blocks of proxies, evenly spread, a
// couple of rows above the server.
func startingCover(width, height int) []Proxy {
	var cover []Proxy
	gap := (width - proxyBlocks*proxyWidth) / (proxyBlocks + 1)
	top := height - 2 - proxyHeight
	for b := 0; b < proxyBlocks; b++ {
		left := gap + b*(proxyWidth+gap)
		for row := top; row < top+proxyHeight; row++ {
			for x := left; x < left+proxyWidth; x++ {
				cover = append(cover, Proxy{X: x, Row: row, HP: ProxyHP})
			}
		}
	}
	return cover
}

// erode takes one HP off the proxy at the cell, removing it at zero, and
// reports whether there was one.
func (g *Game) erode(col, row int) bool {
	for i, p := range g.cover {
		if p.X == col && p.Row == row {
			if p.HP <= 1 {
				g.cover = append(g.cover[:i], g.cover[i+1:]...)
			} else {
				g.cover[i].HP--
			}
			return true
		}
	}
	return false
}

func (g *Game) removeProxy(col, row int) {
	for i, p := range g.cover {
		if p.X == col && p.Row == row {
			g.cover = append(g.cover[:i], g.cover[i+1:]...)
			return
		}
	}
}
