package ui

import (
	"image/color"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func Text() *TextElement {
	el := &TextElement{}
	el.DropShadowElement = NewDropShadowElement(el)
	el.shadowOffsetProvider = func(el *TextElement) vec.Vec2i {
		return fontScaledShadowOffset(el.TextSize)
	}

	return el.WithSizeDynamic(func(el *TextElement) vec.Vec2i {
		return MeasureText(el.Text(), el.TextSize)
	})
}

func MeasureText(text string, textSize int32) vec.Vec2i {
	lines := strings.Split(text, "\n")
	size := vec.Vec2i{Y: int32(len(lines)) * textSize}
	for _, line := range lines {
		size.X = max(size.X, rl.MeasureText(line, textSize))
	}
	return size
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

func (el *TextElement) WithTextShadow(shadowColor color.RGBA, offset vec.Vec2i) *TextElement {
	return el.WithShadow(shadowColor, offset)
}

type TextElement struct {
	DropShadowElement[*TextElement]
	Text      func() string
	TextSize  int32
	TextColor color.RGBA
}

func (el *TextElement) update(deltaNano int64) {
}

func (el *TextElement) draw() {
	pos := el.AbsolutePos()
	text := el.Text()
	opacity := el.Opacity()

	drawTextWithShadow(
		text,
		pos.X,
		pos.Y,
		el.TextSize,
		el.TextColor,
		el.ShadowColor,
		el.resolvedShadowOffset(),
		opacity,
	)
}
