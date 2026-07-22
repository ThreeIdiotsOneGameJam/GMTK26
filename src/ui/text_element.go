package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/math"
)

type TextElement struct {
	Pos       func(renderWidth, renderHeight int32) math.Vec2i
	Text      string
	TextSize  int32
	TextColor color.RGBA

	x, y, w int32
}

func (e *TextElement) update(deltaNano int64) {
	e.w = rl.MeasureText(e.Text, e.TextSize)

	pos := e.Pos(int32(rl.GetRenderWidth()), int32(rl.GetRenderHeight()))
	e.x, e.y = pos.X-e.w/2, pos.Y-e.TextSize/2
}

func (e *TextElement) draw() {
	rl.DrawText(e.Text, e.x, e.y, e.TextSize, e.TextColor)
}
