package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

// Pos and Size do not account for the outline, shadow, or thumb overhang, which
// are rendered outside the element's layout box.

func Toggle() *ToggleElement {
	el := &ToggleElement{
		Value:        false,
		TrackHeight:  24,
		ThumbWidth:   22,
		ThumbHeight:  28,
		OutlineWidth: 4,
		TrackColors: ColorSet{
			Default:  ptrColor(PaletteSurface),
			Hover:    ptrColor(PaletteSurfaceUp),
			Click:    ptrColor(PaletteSurfaceUp),
			Disabled: ptrColor(PaletteBase),
		},
		OnColors: ColorSet{
			Default:  ptrColor(PaletteIndigo),
			Hover:    ptrColor(PaletteIndigoHover),
			Click:    ptrColor(PaletteViolet),
			Disabled: ptrColor(PaletteIndigoDim),
		},
		ThumbColors: ColorSet{
			Default:  ptrColor(PaletteText),
			Hover:    ptrColor(PaletteText),
			Click:    ptrColor(PaletteText),
			Disabled: ptrColor(PaletteTextMuted),
		},
		OutlineColors: ColorSet{
			Default:  ptrColor(PaletteBorder),
			Disabled: ptrColor(PaletteBase),
		},
		Callback: func(bool) {},
	}
	el.DropShadowElement = NewDropShadowElement(el)

	return el.WithSize(vec.Vec2i{X: 64, Y: 36})
}

func (el *ToggleElement) WithValue(value bool) *ToggleElement {
	el.Value = value
	el.ValueProvider = nil
	return el
}

func (el *ToggleElement) WithValueDynamic(valueProvider func() bool) *ToggleElement {
	el.ValueProvider = valueProvider
	return el
}

func (el *ToggleElement) WithTrackHeight(trackHeight int32) *ToggleElement {
	el.TrackHeight = trackHeight
	return el
}

func (el *ToggleElement) WithThumbSize(width, height int32) *ToggleElement {
	el.ThumbWidth = width
	el.ThumbHeight = height
	return el
}

func (el *ToggleElement) WithOutlineWidth(outlineWidth int32) *ToggleElement {
	el.OutlineWidth = outlineWidth
	return el
}

func (el *ToggleElement) WithTrackColors(trackColors ColorSet) *ToggleElement {
	el.TrackColors = trackColors
	return el
}

func (el *ToggleElement) WithOnColors(onColors ColorSet) *ToggleElement {
	el.OnColors = onColors
	return el
}

func (el *ToggleElement) WithThumbColors(thumbColors ColorSet) *ToggleElement {
	el.ThumbColors = thumbColors
	return el
}

func (el *ToggleElement) WithOutlineColors(outlineColors ColorSet) *ToggleElement {
	el.OutlineColors = outlineColors
	return el
}

func (el *ToggleElement) WithCallback(callback func(value bool)) *ToggleElement {
	if callback == nil {
		callback = func(bool) {}
	}
	el.Callback = callback
	return el
}

func (el *ToggleElement) WithCommit(commit func(value bool)) *ToggleElement {
	el.Commit = commit
	return el
}

func (el *ToggleElement) WithDefaultValue(value bool) *ToggleElement {
	el.defaultValue = value
	el.hasDefault = true
	return el
}

func (el *ToggleElement) HasDefault() bool {
	return el.hasDefault
}

func (el *ToggleElement) ResetToDefault() {
	if !el.hasDefault {
		return
	}

	el.setValue(el.defaultValue)
	if el.Commit != nil {
		el.Commit(el.Value)
	}
}

type ToggleElement struct {
	DropShadowElement[*ToggleElement]
	Value         bool
	ValueProvider func() bool
	TrackHeight   int32
	ThumbWidth    int32
	ThumbHeight   int32
	OutlineWidth  int32
	TrackColors   ColorSet
	OnColors      ColorSet
	ThumbColors   ColorSet
	OutlineColors ColorSet
	Callback      func(value bool)
	Commit        func(value bool)

	defaultValue bool
	hasDefault   bool

	x, y, w, h int32
	trackY     int32
	thumbX     int32
	thumbY     int32

	hovered, hoveredPrevious bool
	clicked, clickedPrevious bool
}

func (el *ToggleElement) setValue(value bool) {
	if value == el.Value {
		return
	}

	el.Value = value
	if el.Callback != nil {
		el.Callback(el.Value)
	}
}

func (el *ToggleElement) syncValueFromProvider() {
	if el.ValueProvider != nil && !el.clicked {
		el.Value = el.ValueProvider()
	}
}

func (el *ToggleElement) prepare() {
	el.syncValueFromProvider()

	size := el.Size()
	el.w = max(size.X, el.ThumbWidth)
	el.h = max(size.Y, el.ThumbHeight)

	pos := el.AbsolutePos()
	el.x, el.y = pos.X, pos.Y

	el.trackY = el.y + (el.h-el.TrackHeight)/2
	el.thumbY = el.y + (el.h-el.ThumbHeight)/2
	if el.Value {
		el.thumbX = el.x + el.w - el.ThumbWidth
	} else {
		el.thumbX = el.x
	}
}

func (el *ToggleElement) update(deltaNano int64) {
	if !el.Enabled() || global.UIModalBlocksInput {
		el.hovered = false
		el.clicked = false
		el.hoveredPrevious = false
		el.clickedPrevious = false
		return
	}

	el.syncValueFromProvider()

	mouseX, mouseY := int32(global.MousePosition.X), int32(global.MousePosition.Y)

	inBounds := mouseX >= el.x &&
		mouseX <= el.x+el.w &&
		mouseY >= el.y &&
		mouseY <= el.y+el.h

	inThumb := mouseX >= el.thumbX &&
		mouseX <= el.thumbX+el.ThumbWidth &&
		mouseY >= el.thumbY &&
		mouseY <= el.thumbY+el.ThumbHeight

	el.hovered = inBounds || inThumb

	if el.hovered {
		global.MouseCursorState = rl.MouseCursorPointingHand
		global.UIBlocksWorldInput = true
	}

	if rl.IsMouseButtonPressed(rl.MouseButtonRight) && (inBounds || inThumb) && el.hasDefault {
		el.ResetToDefault()
		return
	}

	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		if el.clickedPrevious {
			el.clicked = true
		} else {
			el.clicked = el.hovered
		}
	} else {
		el.clicked = false

		if el.clickedPrevious && el.hovered {
			el.setValue(!el.Value)
			if el.Commit != nil {
				el.Commit(el.Value)
			}
		}
	}

	if el.clicked {
		global.UIBlocksWorldInput = true
	}

	el.hoveredPrevious, el.clickedPrevious = el.hovered, el.clicked
}

func (el *ToggleElement) draw() {
	state := StateDefault
	switch {
	case !el.Enabled():
		state = StateDisabled
	case el.clicked:
		state = StateClick
	case el.hovered:
		state = StateHover
	}

	opacity := el.Opacity()
	outlineColor := util.ColorOpacity(*el.OutlineColors.Color(state), opacity)
	trackColor := util.ColorOpacity(*el.TrackColors.Color(state), opacity)
	onColor := util.ColorOpacity(*el.OnColors.Color(state), opacity)
	thumbColor := util.ColorOpacity(*el.ThumbColors.Color(state), opacity)

	fillColor := trackColor
	if el.Value {
		fillColor = onColor
	}

	trackOuterW := el.w + el.OutlineWidth*2
	trackOuterH := el.TrackHeight + el.OutlineWidth*2
	trackOuterX := el.x - el.OutlineWidth
	trackOuterY := el.trackY - el.OutlineWidth

	el.drawRectangleShadow(trackOuterX, trackOuterY, trackOuterW, trackOuterH, opacity)
	rl.DrawRectangle(trackOuterX, trackOuterY, trackOuterW, trackOuterH, outlineColor)
	rl.DrawRectangle(el.x, el.trackY, el.w, el.TrackHeight, fillColor)

	thumbOuterW := el.ThumbWidth + el.OutlineWidth*2
	thumbOuterH := el.ThumbHeight + el.OutlineWidth*2
	el.drawRectangleShadow(
		el.thumbX-el.OutlineWidth,
		el.thumbY-el.OutlineWidth,
		thumbOuterW,
		thumbOuterH,
		opacity,
	)
	rl.DrawRectangle(
		el.thumbX-el.OutlineWidth,
		el.thumbY-el.OutlineWidth,
		thumbOuterW,
		thumbOuterH,
		outlineColor,
	)
	rl.DrawRectangle(el.thumbX, el.thumbY, el.ThumbWidth, el.ThumbHeight, thumbColor)
}
