package anchor

import "github.com/threeidiotsonegamejam/gmtk26/src/mathutil/vec"

type Anchor int

const (
	TopLeft Anchor = iota
	Top
	TopRight
	Left
	Center
	Right
	BottomLeft
	Bottom
	BottomRight
)

var AnchorCoords = map[Anchor]vec.Vec2{
	TopLeft:     {X: 0, Y: 0},
	Top:         {X: 0.5, Y: 0},
	TopRight:    {X: 1, Y: 0},
	Left:        {X: 0, Y: 0.5},
	Center:      {X: 0.5, Y: 0.5},
	Right:       {X: 1, Y: 0.5},
	BottomLeft:  {X: 0, Y: 1},
	Bottom:      {X: 0.5, Y: 1},
	BottomRight: {X: 1, Y: 1},
}
