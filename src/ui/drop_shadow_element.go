package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var (
	defaultDropShadowColor  = color.RGBA{R: 0, G: 0, B: 0, A: 160}
	defaultDropShadowOffset = vec.Vec2i{X: 4, Y: 4}
)

// DropShadowElement adds configurable drop-shadow styling to an element.
// Shadows are decorative and do not affect layout or input bounds.
type DropShadowElement[T Element] struct {
	BaseElement[T]

	ShadowColor  *color.RGBA
	ShadowOffset vec.Vec2i

	shadowOffsetProvider func(T) vec.Vec2i
}

func NewDropShadowElement[T Element](self T) DropShadowElement[T] {
	shadowColor := defaultDropShadowColor

	return DropShadowElement[T]{
		BaseElement:  NewBaseElement(self),
		ShadowColor:  &shadowColor,
		ShadowOffset: defaultDropShadowOffset,
	}
}

func (el *DropShadowElement[T]) WithShadow(shadowColor color.RGBA, offset vec.Vec2i) T {
	el.ShadowColor = &shadowColor
	el.ShadowOffset = offset
	el.shadowOffsetProvider = nil
	return el.self
}

func (el *DropShadowElement[T]) WithoutShadow() T {
	el.ShadowColor = nil
	return el.self
}

func (el *DropShadowElement[T]) resolvedShadowOffset() vec.Vec2i {
	if el.shadowOffsetProvider != nil {
		return el.shadowOffsetProvider(el.self)
	}

	return el.ShadowOffset
}

func (el *DropShadowElement[T]) drawRectangleShadow(
	x, y, width, height int32,
	opacity float32,
) {
	if el.ShadowColor == nil {
		return
	}

	offset := el.resolvedShadowOffset()
	rl.DrawRectangle(
		x+offset.X,
		y+offset.Y,
		width,
		height,
		util.ColorOpacity(*el.ShadowColor, opacity),
	)
}

func (el *DropShadowElement[T]) drawShadowedOutlinedRectangle(
	rect elementRect,
	outlineWidth int32,
	outline, fill rl.Color,
	opacity float32,
) {
	el.drawRectangleShadow(
		rect.X-outlineWidth,
		rect.Y-outlineWidth,
		rect.Width+outlineWidth*2,
		rect.Height+outlineWidth*2,
		opacity,
	)
	drawOutlinedRectangle(rect, outlineWidth, outline, fill)
}

func fontScaledShadowOffset(textSize int32) vec.Vec2i {
	// Approximately 1/16 em, rounded to the nearest whole pixel, with enough
	// separation to remain visible at small font sizes.
	offset := max((textSize+8)/16, int32(2))
	return vec.Vec2i{X: offset, Y: offset}
}

func drawTextWithShadow(
	text string,
	x, y, textSize int32,
	textColor color.RGBA,
	shadowColor *color.RGBA,
	shadowOffset vec.Vec2i,
	opacity float32,
) {
	if shadowColor != nil {
		rl.DrawText(
			text,
			x+shadowOffset.X,
			y+shadowOffset.Y,
			textSize,
			util.ColorOpacity(*shadowColor, opacity),
		)
	}

	rl.DrawText(text, x, y, textSize, util.ColorOpacity(textColor, opacity))
}
