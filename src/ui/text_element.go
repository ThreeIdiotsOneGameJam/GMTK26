package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func Text() *TextElement {
	el := &TextElement{}
	el.BaseElement = NewBaseElement(el)

	return el.WithSizeDynamic(func(el *TextElement) vec.Vec2i {
		return vec.Vec2i{
			X: rl.MeasureText(el.Text(), el.TextSize),
			Y: el.TextSize,
		}
	})
}

func (el *TextElement) WithText(text string) *TextElement {
	el.Text = func() string {
		return text
	}
	return el
}

func (el *TextElement) WithTextDynamic(textProvider func() string) *TextElement {
	el.Text = textProvider
	return el
}

func (el *TextElement) WithTextSize(textSize int32) *TextElement {
	el.TextSize = textSize
	return el
}

func (el *TextElement) WithTextColor(textColor color.RGBA) *TextElement {
	el.TextColor = textColor
	return el
}

type TextElement struct {
	BaseElement[*TextElement]
	Text      func() string
	TextSize  int32
	TextColor color.RGBA
}

func (el *TextElement) update(deltaNano int64) {
}

func (el *TextElement) draw() {
	pos := el.AbsolutePos()
	rl.DrawText(el.Text(), pos.X, pos.Y, el.TextSize, el.TextColor)
}
