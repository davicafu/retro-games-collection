package snake

import (
	"fmt"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

type Position = grid.Point

type Direction struct {
	DX int
	DY int
}

var (
	DirectionUp    = Direction{DX: 0, DY: -1}
	DirectionDown  = Direction{DX: 0, DY: 1}
	DirectionLeft  = Direction{DX: -1, DY: 0}
	DirectionRight = Direction{DX: 1, DY: 0}
)

type Config struct {
	Width    int
	Height   int
	Start    Position
	Food     []Position
	Selector FoodSelector
}

type Game struct {
	width     int
	height    int
	body      []Position
	direction Direction
	food      []Position
	selector  FoodSelector
	score     int
	dead      bool
}

func NewGame(cfg Config) *Game {
	return &Game{
		width:     cfg.Width,
		height:    cfg.Height,
		body:      []Position{cfg.Start},
		direction: DirectionRight,
		food:      append([]Position(nil), cfg.Food...),
		selector:  cfg.Selector,
	}
}

func (g *Game) Update(direction Direction) error {
	if g.dead {
		return fmt.Errorf("game over")
	}
	g.direction = direction
	head := g.Head()
	next := Position{X: head.X + direction.DX, Y: head.Y + direction.DY}
	if next.X < 0 || next.Y < 0 || next.X >= g.width || next.Y >= g.height {
		g.dead = true
		return nil
	}
	for _, segment := range g.body {
		if next == segment {
			g.dead = true
			return nil
		}
	}
	g.body = append([]Position{next}, g.body...)
	if len(g.food) > 0 && next == g.food[0] {
		g.food = g.food[1:]
		g.score++
		if len(g.food) == 0 {
			g.respawnFood()
		}
		return nil
	}
	g.body = g.body[:len(g.body)-1]
	return nil
}

func (g *Game) respawnFood() {
	candidates := g.availableFoodCells()
	if len(candidates) == 0 {
		g.dead = true
		return
	}

	if g.selector == nil {
		g.food = []Position{candidates[0]}
		return
	}

	next, ok := g.selector.Select(candidates)
	if !ok {
		g.dead = true
		return
	}

	g.food = []Position{next}
}

func (g *Game) availableFoodCells() []Position {
	occupied := make(map[Position]struct{}, len(g.body)+len(g.food))
	for _, segment := range g.body {
		occupied[segment] = struct{}{}
	}
	for _, food := range g.food {
		occupied[food] = struct{}{}
	}

	var cells []Position
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			candidate := Position{X: x, Y: y}
			if _, blocked := occupied[candidate]; blocked {
				continue
			}
			cells = append(cells, candidate)
		}
	}

	return cells
}

func (g *Game) Head() Position { return g.body[0] }
func (g *Game) Length() int    { return len(g.body) }
func (g *Game) Score() int     { return g.score }
func (g *Game) Dead() bool     { return g.dead }
func (g *Game) Body() []Position {
	return append([]Position(nil), g.body...)
}
func (g *Game) Food() []Position {
	return append([]Position(nil), g.food...)
}
func (g *Game) Direction() Direction { return g.direction }
