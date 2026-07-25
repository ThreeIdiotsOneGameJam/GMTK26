package game

import "math"

type Hex struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

func (h Hex) Add(other Hex) Hex {
	return Hex{X: h.X + other.X, Y: h.Y + other.Y}
}

type Cube struct {
	X float32
	Y float32
	Z float32
}

func (c Cube) ToAxial() Axial {
	return Axial{X: c.X, Y: c.Y}
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

	return Cube{X: q, Y: r, Z: s}
}

type Axial struct {
	X float32
	Y float32
}

func (a Axial) ToCube() Cube {
	return Cube{X: a.X, Y: a.Y, Z: -a.X - a.Y}
}

func (a Axial) Round() Axial {
	return a.ToCube().Round().ToAxial()
}

func (a Axial) ToHex() Hex {
	axial := a.Round()
	parity := float32(int32(axial.X) & 1)
	col := axial.X
	row := axial.Y + (axial.X-parity)/2
	return Hex{X: int32(col), Y: int32(row)}
}
