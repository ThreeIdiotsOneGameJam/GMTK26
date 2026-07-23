package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/mathutil/vec"
)

func renderSize(_ *ScreenElement) vec.Vec2i {
	return vec.Vec2i{X: int32(rl.GetRenderWidth()), Y: int32(rl.GetRenderHeight())}
}

func Screen() *ScreenElement {
	el := &ScreenElement{
		BackgroundColor: rl.RayWhite,
	}
	el.BaseElement = NewBaseElement(el)

	// default screen size is full screen
	return el.WithSizeDynamic(renderSize)
}

func (el *ScreenElement) WithBackgroundColor(backgroundColor color.RGBA) *ScreenElement {
	el.BackgroundColor = backgroundColor
	return el
}

type ScreenElement struct {
	BaseElement[*ScreenElement]
	BackgroundColor color.RGBA
}

func (el *ScreenElement) Update(deltaNano int64) {
	for _, child := range el.Children {
		child.updateTree(deltaNano)
	}
}

func (el *ScreenElement) Draw() {
	el.Clear()

	for _, child := range el.Children {
		child.drawTree()
	}
}

func (el *ScreenElement) Clear() {
	pos, size := el.AbsolutePos(), el.Size()
	rl.DrawRectangle(pos.X, pos.Y, size.X, size.Y, el.BackgroundColor)
}
