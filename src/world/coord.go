package world

import (
	"math"

	v "github.com/threeidiotsonegamejam/gmtk26/src/math"
)

type Hex v.Vec2i

type Cube v.Vec3

func (c Cube) ToAxial() Axial {
	return Axial{X: c.X, Y: c.Y}
}
func (c Cube) Round() Cube {
	q := float32(math.Round(float64(c.X)))
	r := float32(math.Round(float64(c.Y)))
	s := float32(math.Round(float64(c.Z)))

	var q_diff = math.Abs(float64(q - c.X))
	var r_diff = math.Abs(float64(r - c.Y))
	var s_diff = math.Abs(float64(s - c.Z))

	if q_diff > r_diff && q_diff > s_diff {
		q = -r - s
	} else if r_diff > s_diff {
		r = -q - s
	} else {
		s = -q - r
	}

	return Cube{X: q, Y: r, Z: s}
}

type Axial v.Vec2

func (a Axial) ToCube() Cube {
	return Cube{X: a.X, Y: a.Y, Z: -a.X - a.Y}
}
func (a Axial) Round() Axial {
	return a.ToCube().Round().ToAxial()
}
func (a Axial) ToHex() Hex {
	axial := a.Round()
	var parity = float32(int32(axial.X) & 1)
	var col = axial.X
	var row = axial.Y + (axial.X-parity)/2
	return Hex{X: int32(col), Y: int32(row)}
}

func (w World) PixelToHex(p v.Vec2) Hex {
	q := (2.0 / 3.0 * p.X) / w.HexSize
	r := (-1.0/3.0*p.X + sqrt3/3.0*p.Y) / w.HexSize

	return Axial{X: q, Y: r}.ToHex()
}
