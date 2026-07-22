package ui

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/math"
)

type Element interface {
	GetPos() math.Vec2i
	GetSize() math.Vec2i

	Update(delta float32)
	Draw()
}
