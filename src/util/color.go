package util

import (
	"image/color"
)

type ColorState int

const (
	DefaultState ColorState = iota
	HoverState
	ClickState
)

type ColorSet struct {
	Default *color.RGBA
	Hover   *color.RGBA
	Click   *color.RGBA
}

func NewColorSet(defaultColor *color.RGBA) ColorSet {
	if defaultColor == nil {
		panic("ColorSet.Default must not be nil")
	}

	return ColorSet{
		Default: defaultColor,
	}
}

func NewColorSetHover(defaultColor, hoverColor *color.RGBA) ColorSet {
	if defaultColor == nil {
		panic("ColorSet.Default must not be nil")
	}

	return ColorSet{
		Default: defaultColor,
		Hover:   hoverColor,
	}
}

func NewColorSetClick(defaultColor, clickColor *color.RGBA) ColorSet {
	if defaultColor == nil {
		panic("ColorSet.Default must not be nil")
	}

	return ColorSet{
		Default: defaultColor,
		Click:   clickColor,
	}
}

func NewColorSetHoverClick(defaultColor, hoverColor, clickColor *color.RGBA) ColorSet {
	if defaultColor == nil {
		panic("ColorSet.Default must not be nil")
	}

	return ColorSet{
		Default: defaultColor,
		Hover:   hoverColor,
		Click:   clickColor,
	}
}

func (cs ColorSet) Color(state ColorState) *color.RGBA {
	if cs.Default == nil {
		panic("ColorSet.Default must not be nil")
	}

	switch state {
	case ClickState:
		if cs.Click != nil {
			return cs.Click
		}
		fallthrough

	case HoverState:
		if cs.Hover != nil {
			return cs.Hover
		}
	default:
		return cs.Default
	}

	return cs.Default
}

func SimpleGrayscaleColor(c uint8) *color.RGBA {
	return &color.RGBA{
		R: c,
		G: c,
		B: c,
		A: 255,
	}
}

func ColorAdd(color color.RGBA, n uint8) *color.RGBA {
	color.R = ClampByte(int32(color.R) + int32(n))
	color.G = ClampByte(int32(color.G) + int32(n))
	color.B = ClampByte(int32(color.B) + int32(n))
	return &color
}

func ColorSub(color color.RGBA, n uint8) *color.RGBA {
	color.R = ClampByte(int32(color.R) - int32(n))
	color.G = ClampByte(int32(color.G) - int32(n))
	color.B = ClampByte(int32(color.B) - int32(n))
	return &color
}
