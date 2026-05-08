package sokoban

import (
	"errors"
	"strings"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// Level definitions — progressively harder puzzles.
// Each level MUST have balanced boxes ($) and targets (.) counts.
var Levels = []string{
	// Level 1: Tight space, 1 box (intro)
	`#####
#@  #
# $ #
# . #
#####`,

	// Level 2: Two boxes, corner trick
	`######
#    #
# $@ #
#  . #
# $ .#
######`,

	// Level 3: Three boxes, need planning
	`#######
#  @  #
# $.$.#
#  $  #
#  .  #
#######`,

	// Level 4: Three boxes, narrow corridor
	`########
#      #
# .$@$ #
#      #
# . $ .#
########`,

	// Level 5: Four boxes, strategy required
	`##########
#   @    #
#  $ $ $ #
# . $ .  #
# .   .  #
##########`,

	// Level 6: Three boxes, dead-end traps
	`########
#   @  #
# $    #
# . $ .#
#    $ #
#   .  #
########`,

	// Level 7: Five boxes, multiple paths to deadlock
	`##########
# @      #
# $ $ $  #
# . $ .  #
#  $     #
# . . .  #
##########`,

	// Level 8: Five boxes, tight squeeze
	`##########
#   @    #
#  $ $   #
# . $ .  #
#  $ $   #
# . . .  #
##########`,

	// Level 9: Five boxes, the gauntlet
	`###########
#  @      #
#  $      #
#  . $ .  #
#   $$    #
# . $.    #
#    .    #
###########`,

	// Level 10: Five boxes, final challenge
	`############
#   @      #
#   $      #
#  . $ .   #
#   $ $    #
#  . $ .   #
#   .      #
############`,
}

func ParseLevel(lines []string) (*Game, error) {
	if len(lines) < 3 {
		return nil, errors.New("level must have at least 3 rows")
	}

	height := len(lines)
	width := 0
	for _, line := range lines {
		if len(line) > width {
			width = len(line)
		}
	}

	var player grid.Point
	var playerFound bool
	var boxes []grid.Point
	var targets []grid.Point
	var walls []grid.Point

	for y, line := range lines {
		for x := 0; x < len(line); x++ {
			ch := line[x]
			switch ch {
			case '#':
				walls = append(walls, grid.Point{X: x, Y: y})
			case '@':
				player = grid.Point{X: x, Y: y}
				playerFound = true
			case '$':
				boxes = append(boxes, grid.Point{X: x, Y: y})
			case '.':
				targets = append(targets, grid.Point{X: x, Y: y})
			case '*': // box on target
				boxes = append(boxes, grid.Point{X: x, Y: y})
				targets = append(targets, grid.Point{X: x, Y: y})
			case '+': // player on target
				player = grid.Point{X: x, Y: y}
				playerFound = true
				targets = append(targets, grid.Point{X: x, Y: y})
			}
		}
	}

	if !playerFound {
		return nil, errors.New("no player found")
	}

	return NewGame(width, height, player, walls, boxes, targets), nil
}

func ParseLevelString(level string) (*Game, error) {
	lines := strings.Split(strings.TrimSpace(level), "\n")
	// Remove trailing empty line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return ParseLevel(lines)
}
