package ui

import "image/color"

//go:generate stringer -type=UIState -trimprefix=State

type UIState int

const (
	StateDefault UIState = iota
	StateHover
	StateClick
	StateDisabled
)

type ColorSet struct {
	Default  *color.RGBA
	Hover    *color.RGBA
	Click    *color.RGBA
	Disabled *color.RGBA
}

func (cs ColorSet) Color(state UIState) *color.RGBA {
	if cs.Default == nil {
		panic("ColorSet.Default must not be nil")
	}

	switch state {
	case StateDisabled:
		if cs.Disabled != nil {
			return cs.Disabled
		}
		return cs.Default

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
