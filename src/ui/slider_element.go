package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

// Pos and Size do not account for the outline, shadow, or thumb overhang, which
// are rendered outside the element's layout box.

func Slider() *SliderElement {
	el := &SliderElement{
		Value:         0,
		Min:           0,
		Max:           1,
		TrackHeight:   12,
		ThumbWidth:    18,
		ThumbHeight:   28,
		OutlineWidth:  4,
		TrackColors:   DefaultTrackColors,
		FillColors:    DefaultActiveTrackColors,
		ThumbColors:   DefaultThumbColors,
		OutlineColors: DefaultTrackOutlineColors,
		Callback:      func(float32) {},
	}
	el.DropShadowElement = NewDropShadowElement(el)

	return el.WithSize(vec.Vec2i{X: 360, Y: 36})
}

func (el *SliderElement) WithValue(value float32) *SliderElement {
	el.Value = value
	el.ValueProvider = nil
	return el
}

func (el *SliderElement) WithValueDynamic(valueProvider func() float32) *SliderElement {
	el.ValueProvider = valueProvider
	return el
}

func (el *SliderElement) WithRange(minValue, maxValue float32) *SliderElement {
	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}
	el.Min = minValue
	el.Max = maxValue
	el.Value = el.clampValue(el.Value)
	return el
}

func (el *SliderElement) WithTrackHeight(trackHeight int32) *SliderElement {
	el.TrackHeight = trackHeight
	return el
}

func (el *SliderElement) WithThumbSize(width, height int32) *SliderElement {
	el.ThumbWidth = width
	el.ThumbHeight = height
	return el
}

func (el *SliderElement) WithOutlineWidth(outlineWidth int32) *SliderElement {
	el.OutlineWidth = outlineWidth
	return el
}

func (el *SliderElement) WithTrackColors(trackColors ColorSet) *SliderElement {
	el.TrackColors = trackColors
	return el
}

func (el *SliderElement) WithFillColors(fillColors ColorSet) *SliderElement {
	el.FillColors = fillColors
	return el
}

func (el *SliderElement) WithThumbColors(thumbColors ColorSet) *SliderElement {
	el.ThumbColors = thumbColors
	return el
}

func (el *SliderElement) WithOutlineColors(outlineColors ColorSet) *SliderElement {
	el.OutlineColors = outlineColors
	return el
}

func (el *SliderElement) WithCallback(callback func(value float32)) *SliderElement {
	if callback == nil {
		callback = func(float32) {}
	}
	el.Callback = callback
	return el
}

func (el *SliderElement) WithCommit(commit func(value float32)) *SliderElement {
	el.Commit = commit
	return el
}

func (el *SliderElement) WithDefaultValue(value float32) *SliderElement {
	el.defaultValue = value
	el.hasDefault = true
	return el
}

func (el *SliderElement) HasDefault() bool {
	return el.hasDefault
}

func (el *SliderElement) ResetToDefault() {
	if !el.hasDefault {
		return
	}

	el.dragging = false
	el.setValue(el.defaultValue)
	if el.Commit != nil {
		el.Commit(el.Value)
	}
}

type SliderElement struct {
	DropShadowElement[*SliderElement]
	Value         float32
	ValueProvider func() float32
	Min, Max      float32
	TrackHeight   int32
	ThumbWidth    int32
	ThumbHeight   int32
	OutlineWidth  int32
	TrackColors   ColorSet
	FillColors    ColorSet
	ThumbColors   ColorSet
	OutlineColors ColorSet
	Callback      func(value float32)
	Commit        func(value float32)

	defaultValue float32
	hasDefault   bool

	x, y, w, h int32
	trackY     int32
	thumbX     int32
	thumbY     int32

	hovered, dragging bool
}

func (el *SliderElement) clampValue(value float32) float32 {
	return min(max(value, el.Min), el.Max)
}

func (el *SliderElement) normalized() float32 {
	if el.Max <= el.Min {
		return 0
	}
	return (el.clampValue(el.Value) - el.Min) / (el.Max - el.Min)
}

func (el *SliderElement) valueFromMouseX(mouseX int32) float32 {
	if el.w <= 0 {
		return el.Min
	}

	t := float32(mouseX-el.x) / float32(el.w)
	t = min(max(t, 0), 1)
	return el.Min + t*(el.Max-el.Min)
}

func (el *SliderElement) setValue(value float32) {
	value = el.clampValue(value)
	if value == el.Value {
		return
	}

	el.Value = value
	if el.Callback != nil {
		el.Callback(el.Value)
	}
}

func (el *SliderElement) syncValueFromProvider() {
	if el.ValueProvider != nil && !el.dragging {
		el.Value = el.clampValue(el.ValueProvider())
	}
}

func (el *SliderElement) prepare() {
	el.syncValueFromProvider()

	size := el.Size()
	el.w = max(size.X, 1)
	el.h = max(size.Y, el.ThumbHeight)

	pos := el.AbsolutePos()
	el.x, el.y = pos.X, pos.Y

	el.trackY = el.y + (el.h-el.TrackHeight)/2
	el.thumbX = el.x + int32(el.normalized()*float32(el.w)) - el.ThumbWidth/2
	el.thumbY = el.y + (el.h-el.ThumbHeight)/2
}

func (el *SliderElement) update(deltaNano int64) {
	if !el.Enabled() || global.UIModalBlocksInput {
		if el.dragging {
			el.dragging = false
			if el.Commit != nil {
				el.Commit(el.Value)
			}
		}
		el.hovered = false
		return
	}

	el.syncValueFromProvider()

	mouse := mousePosition()
	mouseX := mouse.X
	inBounds := (elementRect{X: el.x, Y: el.y, Width: el.w, Height: el.h}).contains(mouse)
	inThumb := (elementRect{
		X: el.thumbX, Y: el.thumbY,
		Width: el.ThumbWidth, Height: el.ThumbHeight,
	}).contains(mouse)

	el.hovered = inBounds || inThumb || el.dragging

	if el.hovered {
		claimPointer(rl.MouseCursorPointingHand)
	}

	if rl.IsMouseButtonPressed(rl.MouseButtonRight) && (inBounds || inThumb) && el.hasDefault {
		el.ResetToDefault()
		return
	}

	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && (inBounds || inThumb) {
		el.dragging = true
		el.setValue(el.valueFromMouseX(mouseX))
	}

	if el.dragging {
		global.UIBlocksWorldInput = true
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			el.setValue(el.valueFromMouseX(mouseX))
		} else {
			el.dragging = false
			if el.Commit != nil {
				el.Commit(el.Value)
			}
		}
	}
}

func (el *SliderElement) draw() {
	state := controlState(el.Enabled(), el.hovered, el.dragging)

	opacity := el.Opacity()
	outlineColor := util.ColorOpacity(*el.OutlineColors.Color(state), opacity)
	trackColor := util.ColorOpacity(*el.TrackColors.Color(state), opacity)
	fillColor := util.ColorOpacity(*el.FillColors.Color(state), opacity)
	thumbColor := util.ColorOpacity(*el.ThumbColors.Color(state), opacity)

	el.drawShadowedOutlinedRectangle(
		elementRect{X: el.x, Y: el.trackY, Width: el.w, Height: el.TrackHeight},
		el.OutlineWidth,
		outlineColor,
		trackColor,
		opacity,
	)

	fillW := int32(el.normalized() * float32(el.w))
	if fillW > 0 {
		rl.DrawRectangle(el.x, el.trackY, fillW, el.TrackHeight, fillColor)
	}

	el.drawShadowedOutlinedRectangle(
		elementRect{
			X: el.thumbX, Y: el.thumbY,
			Width: el.ThumbWidth, Height: el.ThumbHeight,
		},
		el.OutlineWidth,
		outlineColor,
		thumbColor,
		opacity,
	)
}
