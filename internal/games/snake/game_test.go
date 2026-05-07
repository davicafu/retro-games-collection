package snake

import "testing"

func TestGame_UpdateMovesHeadAndGrowsOnFood(t *testing.T) {
	g := NewGame(Config{
		Width:  8,
		Height: 8,
		Start:  Position{X: 2, Y: 2},
		Food:   []Position{{X: 3, Y: 2}},
	})

	if err := g.Update(DirectionRight); err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if got := g.Head(); got != (Position{X: 3, Y: 2}) {
		t.Fatalf("expected head at food tile, got %+v", got)
	}

	if g.Score() != 1 {
		t.Fatalf("expected score 1, got %d", g.Score())
	}

	if g.Length() != 2 {
		t.Fatalf("expected snake to grow to length 2, got %d", g.Length())
	}
}

func TestGame_UpdateMovesWithoutFoodKeepsLength(t *testing.T) {
	g := NewGame(Config{
		Width:  8,
		Height: 8,
		Start:  Position{X: 2, Y: 2},
	})

	if err := g.Update(DirectionRight); err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if got := g.Head(); got != (Position{X: 3, Y: 2}) {
		t.Fatalf("expected moved head, got %+v", got)
	}

	if g.Length() != 1 {
		t.Fatalf("expected length to remain 1, got %d", g.Length())
	}

	if g.Score() != 0 {
		t.Fatalf("expected score 0, got %d", g.Score())
	}
}

func TestGame_UpdateWallCollisionMarksDeadAndNoError(t *testing.T) {
	g := NewGame(Config{
		Width:  3,
		Height: 3,
		Start:  Position{X: 2, Y: 1},
	})

	if err := g.Update(DirectionRight); err != nil {
		t.Fatalf("expected no error on fatal move, got %v", err)
	}

	if !g.Dead() {
		t.Fatalf("expected game to be dead after wall collision")
	}
}

func TestGame_UpdateAfterDeathReturnsGameOverErrorAndNoStateChange(t *testing.T) {
	g := NewGame(Config{
		Width:  3,
		Height: 3,
		Start:  Position{X: 2, Y: 1},
	})

	if err := g.Update(DirectionRight); err != nil {
		t.Fatalf("expected no error on fatal move, got %v", err)
	}

	headBefore := g.Head()
	lengthBefore := g.Length()
	scoreBefore := g.Score()

	err := g.Update(DirectionUp)
	if err == nil {
		t.Fatalf("expected game over error after death")
	}
	if err.Error() != "game over" {
		t.Fatalf("expected game over error, got %v", err)
	}

	if got := g.Head(); got != headBefore {
		t.Fatalf("expected head unchanged after death, got %+v", got)
	}
	if g.Length() != lengthBefore {
		t.Fatalf("expected length unchanged after death, got %d", g.Length())
	}
	if g.Score() != scoreBefore {
		t.Fatalf("expected score unchanged after death, got %d", g.Score())
	}
}

func TestGame_DeterministicMultiStepSimulation(t *testing.T) {
	g := NewGame(Config{
		Width:  8,
		Height: 8,
		Start:  Position{X: 1, Y: 1},
		Food: []Position{
			{X: 2, Y: 1},
			{X: 3, Y: 1},
		},
	})

	steps := []Direction{DirectionRight, DirectionRight, DirectionDown, DirectionLeft}
	for i, step := range steps {
		if err := g.Update(step); err != nil {
			t.Fatalf("step %d failed: %v", i, err)
		}
	}

	if got := g.Head(); got != (Position{X: 2, Y: 2}) {
		t.Fatalf("expected deterministic final head position, got %+v", got)
	}
	if g.Score() != 2 {
		t.Fatalf("expected score 2 after consuming queued food, got %d", g.Score())
	}
	if g.Length() != 3 {
		t.Fatalf("expected final length 3 after two pickups, got %d", g.Length())
	}

	food := g.Food()
	if len(food) != 1 {
		t.Fatalf("expected one respawned food after queue is exhausted, got %d", len(food))
	}
	if g.Dead() {
		t.Fatalf("expected snake alive")
	}
}

func TestGame_UpdateRespawnsFoodAfterPickup(t *testing.T) {
	g := NewGame(Config{
		Width:    6,
		Height:   6,
		Start:    Position{X: 2, Y: 2},
		Food:     []Position{{X: 3, Y: 2}},
		Selector: NewSequenceSelector([]int{1}),
	})

	if err := g.Update(DirectionRight); err != nil {
		t.Fatalf("expected pickup update to succeed, got %v", err)
	}

	if g.Score() != 1 {
		t.Fatalf("expected score 1, got %d", g.Score())
	}

	food := g.Food()
	if len(food) != 1 {
		t.Fatalf("expected one respawned food item, got %d", len(food))
	}
	if food[0] != (Position{X: 1, Y: 0}) {
		t.Fatalf("expected selector-driven respawn at (1,0), got %+v", food[0])
	}
	for _, segment := range g.Body() {
		if food[0] == segment {
			t.Fatalf("expected food not to overlap snake body, got %+v", food[0])
		}
	}
}

func TestGame_UpdateRespawnsFoodAfterPickupWithNilSelector(t *testing.T) {
	g := NewGame(Config{
		Width:  5,
		Height: 5,
		Start:  Position{X: 1, Y: 1},
		Food:   []Position{{X: 2, Y: 1}},
	})

	if err := g.Update(DirectionRight); err != nil {
		t.Fatalf("expected pickup update to succeed, got %v", err)
	}

	if g.Score() != 1 {
		t.Fatalf("expected score 1, got %d", g.Score())
	}

	food := g.Food()
	if len(food) != 1 {
		t.Fatalf("expected one respawned food item with nil selector, got %d", len(food))
	}

	for _, segment := range g.Body() {
		if food[0] == segment {
			t.Fatalf("expected food not to overlap snake body, got %+v", food[0])
		}
	}
}

func TestGame_SelfCollisionDetectsBodyHit(t *testing.T) {
	g := NewGame(Config{
		Width:  8,
		Height: 8,
		Start:  Position{X: 2, Y: 2},
		Food:   []Position{{X: 3, Y: 2}, {X: 4, Y: 2}, {X: 5, Y: 2}},
	})

	// Grow to length 4 heading right so the snake has a body to collide with.
	g.Update(DirectionRight) // eat (3,2), body = [(3,2), (2,2)], len 2
	g.Update(DirectionRight) // eat (4,2), body = [(4,2), (3,2), (2,2)], len 3
	g.Update(DirectionRight) // eat (5,2), body = [(5,2), (4,2), (3,2), (2,2)], len 4

	// Form a C-shape: right, down, left — then try to go up into the body.
	g.Update(DirectionDown) // body = [(5,3), (5,2), (4,2), (3,2)]
	g.Update(DirectionLeft) // body = [(4,3), (5,3), (5,2), (4,2)]

	// Up from (4,3) → next = (4,2) which IS the tail (body[3]).
	if err := g.Update(DirectionUp); err != nil {
		t.Fatalf("expected no error on fatal self-collision, got %v", err)
	}

	if !g.Dead() {
		t.Fatal("expected snake to die from self-collision")
	}
}

func TestGame_UpdateDoesNotFalsePositiveOnNonCollidingTurn(t *testing.T) {
	g := NewGame(Config{
		Width:  8,
		Height: 8,
		Start:  Position{X: 2, Y: 2},
		Food:   []Position{{X: 3, Y: 2}, {X: 4, Y: 2}},
	})

	// Grow to length 3 and spiral without hitting body.
	g.Update(DirectionRight) // eat (3,2), body = [(3,2), (2,2)], len 2
	g.Update(DirectionRight) // eat (4,2), body = [(4,2), (3,2), (2,2)], len 3
	g.Update(DirectionDown)  // body = [(4,3), (4,2), (3,2)], len 3
	g.Update(DirectionLeft)  // body = [(3,3), (4,3), (4,2)], len 3

	// Up from (3,3) → next = (3,2). (3,2) was the neck but tail moved away.
	if err := g.Update(DirectionUp); err != nil {
		t.Fatalf("expected non-colliding move to succeed, got %v", err)
	}

	if g.Dead() {
		t.Fatal("expected snake to be alive after non-colliding turn")
	}
}

func TestGame_UpdateEndsWhenNoFoodCellRemains(t *testing.T) {
	g := NewGame(Config{
		Width:    2,
		Height:   1,
		Start:    Position{X: 0, Y: 0},
		Food:     []Position{{X: 1, Y: 0}},
		Selector: NewSequenceSelector([]int{0}),
	})

	if err := g.Update(DirectionRight); err != nil {
		t.Fatalf("expected final pickup update to succeed, got %v", err)
	}

	if !g.Dead() {
		t.Fatal("expected game to end when no valid food cell remains")
	}
	if got := g.Food(); len(got) != 0 {
		t.Fatalf("expected no food remaining, got %+v", got)
	}
}
