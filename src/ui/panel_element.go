package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
)

// Panel is a simple rounded surface for grouping related UI.
func Panel() *PanelElement {
	el := &PanelElement{
		BackgroundColor: PaletteSurfaceUp,
		OutlineColor:    PaletteBorder,
		OutlineWidth:    2,
		Roundness:       0.08,
	}
	el.BaseElement = NewBaseElement(el)
	return el
}

func (el *PanelElement) WithBackgroundColor(backgroundColor color.RGBA) *PanelElement {
	el.BackgroundColor = backgroundColor
	return el
}

func (el *PanelElement) WithOutlineColor(outlineColor color.RGBA) *PanelElement {
	el.OutlineColor = outlineColor
	return el
}

func (el *PanelElement) WithOutlineWidth(outlineWidth float32) *PanelElement {
	el.OutlineWidth = outlineWidth
	return el
}

func (el *PanelElement) WithRoundness(roundness float32) *PanelElement {
	el.Roundness = max(float32(0), min(roundness, float32(1)))
	return el
}

type PanelElement struct {
	BaseElement[*PanelElement]
	BackgroundColor color.RGBA
	OutlineColor    color.RGBA
	OutlineWidth    float32
	Roundness       float32
}

func (el *PanelElement) draw() {
	pos := el.AbsolutePos()
	size := el.Size()
	if size.X <= 0 || size.Y <= 0 {
		return
	}

	rect := rl.Rectangle{
		X:      float32(pos.X),
		Y:      float32(pos.Y),
		Width:  float32(size.X),
		Height: float32(size.Y),
	}
	opacity := el.Opacity()
	rl.DrawRectangleRounded(
		rect,
		el.Roundness,
		8,
		util.ColorOpacity(el.BackgroundColor, opacity),
	)
	if el.OutlineWidth > 0 && el.OutlineColor.A > 0 {
		rl.DrawRectangleRoundedLinesEx(
			rect,
			el.Roundness,
			8,
			el.OutlineWidth,
			util.ColorOpacity(el.OutlineColor, opacity),
		)
	}
}
