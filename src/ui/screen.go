package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
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

func (el *ScreenElement) WithEnter(enter func()) *ScreenElement {
	el.OnEnter = enter
	return el
}

func (el *ScreenElement) WithExit(exit func()) *ScreenElement {
	el.OnExit = exit
	return el
}

type ScreenElement struct {
	BaseElement[*ScreenElement]
	BackgroundColor color.RGBA
	OnEnter         func()
	OnExit          func()
}

func (el *ScreenElement) Enter() {
	if el.OnEnter != nil {
		el.OnEnter()
	}
}

func (el *ScreenElement) Exit() {
	if el.OnExit != nil {
		el.OnExit()
	}
}

func (el *ScreenElement) Update(deltaNano int64) {
	for _, child := range el.Children {
		child.updateTree(deltaNano)
	}
}

// Prepare recalculates render data without processing input or advancing state.
// Draw calls this automatically so a newly activated screen is laid out before
// its first frame is rendered.
func (el *ScreenElement) Prepare() {
	for _, child := range el.Children {
		child.prepareTree()
	}
}

func (el *ScreenElement) Draw() {
	el.Prepare()
	el.Clear()

	for _, child := range el.Children {
		child.drawTree()
	}
}

func (el *ScreenElement) Clear() {
	pos, size := el.AbsolutePos(), el.Size()
	rl.DrawRectangle(pos.X, pos.Y, size.X, size.Y, el.BackgroundColor)
}
