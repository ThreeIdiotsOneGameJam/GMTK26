//go:build !web

package render

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type hexBatch struct{}

func newHexBatch() *hexBatch {
	rl.Begin(rl.Triangles)
	return &hexBatch{}
}

func (b *hexBatch) Add(x, y float32, size vec.Vec2, tint color.RGBA) {
	drawHexagonBuffered(x, y, size, tint)
}

func (b *hexBatch) Draw() {
	rl.End()
}
