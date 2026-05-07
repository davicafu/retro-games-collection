package render

type Cell struct{ Rune rune }

type Screen interface {
	Size() (int, int)
	Clear()
	SetCell(x, y int, cell Cell)
	Present() error
}
