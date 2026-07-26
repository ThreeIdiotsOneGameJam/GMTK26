package game

import (
	"math"

	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type Hex struct {
	vec.Vec2i
}

func NewHex(x, y int32) Hex {
	return Hex{Vec2i: vec.Vec2i{X: x, Y: y}}
}

func (h Hex) Add(other Hex) Hex {
	return Hex{Vec2i: h.Vec2i.Add(other.Vec2i)}
}

// Distance returns the number of steps in the shortest unobstructed path
// between two hexes. Hex uses odd-column offset coordinates, so convert them
// to cube coordinates before measuring the distance.
func (h Hex) Distance(other Hex) int32 {
	hx, hy, hz := h.cubeCoordinates()
	ox, oy, oz := other.cubeCoordinates()

	return (abs32(hx-ox) + abs32(hy-oy) + abs32(hz-oz)) / 2
}

func (h Hex) cubeCoordinates() (x, y, z int32) {
	x = h.X
	z = h.Y - (h.X-(h.X&1))/2
	y = -x - z
	return
}

func abs32(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}

type Cube struct {
	vec.Vec3
}

func NewCube(x, y, z float32) Cube {
	return Cube{Vec3: vec.Vec3{X: x, Y: y, Z: z}}
}

func (c Cube) ToAxial() Axial {
	return NewAxial(c.X, c.Y)
}

func (c Cube) Round() Cube {
	q := float32(math.Round(float64(c.X)))
	r := float32(math.Round(float64(c.Y)))
	s := float32(math.Round(float64(c.Z)))

	qDiff := math.Abs(float64(q - c.X))
	rDiff := math.Abs(float64(r - c.Y))
	sDiff := math.Abs(float64(s - c.Z))

	if qDiff > rDiff && qDiff > sDiff {
		q = -r - s
	} else if rDiff > sDiff {
		r = -q - s
	} else {
		s = -q - r
	}

	return NewCube(q, r, s)
}

type Axial struct {
	vec.Vec2
}

func NewAxial(x, y float32) Axial {
	return Axial{Vec2: vec.Vec2{X: x, Y: y}}
}

func (a Axial) ToCube() Cube {
	return NewCube(a.X, a.Y, -a.X-a.Y)
}

func (a Axial) Round() Axial {
	return a.ToCube().Round().ToAxial()
}

func (a Axial) ToHex() Hex {
	axial := a.Round()
	parity := float32(int32(axial.X) & 1)
	col := axial.X
	row := axial.Y + (axial.X-parity)/2
	return NewHex(int32(col), int32(row))
}
