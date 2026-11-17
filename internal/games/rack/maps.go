package rack

// Width and Height are every level's size in cells.
const (
	Width  = 36
	Height = 18
)

// Levels are the data center's floors, played in turn. Legend: '=' a rack
// floor, 'H' a ladder, '$' a disk, 'P' the technician's start, 'E' the server
// room door, 'C' the cable spool, 'B' a bot's start and 'T' a toolbox. Each
// floor has a gap at one end, alternating sides, so cables zigzag down to
// the bottom.
var Levels = [][]string{
	{
		"                                    ",
		" E        $      C                  ",
		"======H=============H=============  ",
		"      H             H               ",
		"      H             H         $     ",
		"  ==========H===============H=======",
		"            H               H       ",
		"   $        H     B         H       ",
		"========H===============H=========  ",
		"        H               H           ",
		"        H               H        $  ",
		"  ==============H=============H=====",
		"                H             H     ",
		"         $      H        T    H     ",
		"====H=================H===========  ",
		"    H                 H             ",
		"  P H                 H    $        ",
		"====================================",
	},
	{
		"                                    ",
		" E             C              $     ",
		"==========H===============H=======  ",
		"          H               H         ",
		"  $       H         B     H      $  ",
		"  ===H============H=================",
		"     H            H                 ",
		"     H            H   $             ",
		"==============H================H==  ",
		"              H                H    ",
		"     $        H         T      H    ",
		"  =======H============H=============",
		"         H            H             ",
		"         H     B      H       $     ",
		"=================H==========H=====  ",
		"                 H          H       ",
		"  P         $    H          H       ",
		"====================================",
	},
	{
		"                                    ",
		" E            C         $           ",
		"====H============H===========H====  ",
		"    H            H           H      ",
		"    H         $  H           H      ",
		"  =======H=============H============",
		"         H             H            ",
		"         H$            H B        $ ",
		"======H============H==============  ",
		"      H            H                ",
		"  $   H B    T     H        $       ",
		"  ===========H=============H========",
		"             H             H        ",
		"             H    B $      H        ",
		"========H===============H=========  ",
		"        H               H           ",
		"  P     H               H      $    ",
		"====================================",
	},
}
