package sokoban

import (
	"errors"
	"strings"

	"github.com/davidbarreiro/cli-games-collection/internal/engine/grid"
)

// Level definitions — progressively harder classic puzzles.
// Each level MUST have balanced boxes ($) and targets (.) counts.
var Levels = []string{
	// Level 1: Simple push (1 box, 1 target)
	`#####
#@  #
# $ #
#  .#
#####`,

	// Level 2: Two boxes, L-shaped (2 boxes, 2 targets)
	`######
#    #
# $. #
# @$ #
# .  #
######`,

	// Level 3: Push around corner (2 boxes, 2 targets)
	`  ####
###  #
#.$  #
### $.#
#  @ #
#######`,

	// Level 4: Two boxes, linear (2 boxes, 2 targets)
	`######
# @  #
# $  #
# .  #
#  $.#
######`,

	// Level 5: Three boxes, cross (3 boxes, 3 targets)
	`#######
#  @  #
# $.$. #
#  $  #
#  .  #
#######`,

	// Level 6: Compact, requires planning (2 boxes, 2 targets)
	` ####
##  #
# $.#
# $.#
# @ #
####`,

	// Level 7: Diagonal pattern (2 boxes, 2 targets)
	`######
# .  #
# $  #
# @$ #
#  . #
######`,

	// Level 8: Three boxes, staggered (3 boxes, 3 targets)
	`########
#  @   #
# $    #
#  $.  #
#   $  #
#   .  #
#    . #
########`,

	// Level 9: Three boxes, corridor (3 boxes, 3 targets)
	`########
#      #
# .$   #
#  .$  #
#   .$ #
#  @   #
#      #
########`,

	// Level 10: Four boxes, symmetric (4 boxes, 4 targets)
	`#########
#  @    #
#  $    #
# . $ . #
#  $ $  #
# .  .  #
#       #
#########`,
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
