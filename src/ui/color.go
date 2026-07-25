package ui

import "image/color"

//go:generate stringer -type=UIState -trimprefix=State

type UIState int

const (
	StateDefault UIState = iota
	StateHover
	StateClick
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

func (cs ColorSet) Color(state UIState) *color.RGBA {
	if cs.Default == nil {
		panic("ColorSet.Default must not be nil")
	}

	switch state {
	case StateClick:
		if cs.Click != nil {
			return cs.Click
		}
		fallthrough

	case StateHover:
		if cs.Hover != nil {
			return cs.Hover
		}
	default:
		return cs.Default
	}

	return cs.Default
}
