package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

// Pos and Size do not account for the outline, which is rendered outside this

func Button() *ButtonElement {
	el := &ButtonElement{
		Text:         "Button",
		TextSize:     48,
		Padding:      8,
		OutlineWidth: 4,
		ForegroundColors: ColorSet{
			Default:  ptrColor(PaletteText),
			Disabled: ptrColor(PaletteTextMuted),
		},
		BackgroundColors: ColorSet{
			Default:  ptrColor(PaletteIndigo),
			Hover:    ptrColor(PaletteIndigoHover),
			Click:    ptrColor(PaletteIndigoPress),
			Disabled: ptrColor(PaletteIndigoDim),
		},
		OutlineColors: ColorSet{
			Default:  ptrColor(PaletteBorder),
			Disabled: ptrColor(PaletteSurface),
		},
	}
	el.BaseElement = NewBaseElement(el)

	return el.WithSizeDynamic(func(el *ButtonElement) vec.Vec2i {
		return vec.Vec2i{
			X: rl.MeasureText(el.Text, el.TextSize) + el.Padding*2,
			Y: el.TextSize + el.Padding*2,
		}
	})
}

func (el *ButtonElement) WithText(text string) *ButtonElement {
	el.Text = text
	return el
}

func (el *ButtonElement) WithTextSize(textSize int32) *ButtonElement {
	el.TextSize = textSize
	return el
}

func (el *ButtonElement) WithPadding(padding int32) *ButtonElement {
	el.Padding = padding
	return el
}

func (el *ButtonElement) WithOutlineWidth(outlineWidth int32) *ButtonElement {
	el.OutlineWidth = outlineWidth
	return el
}

func (el *ButtonElement) WithForegroundColors(foregroundColors ColorSet) *ButtonElement {
	el.ForegroundColors = foregroundColors
	return el
}

func (el *ButtonElement) WithBackgroundColors(backgroundColors ColorSet) *ButtonElement {
	el.BackgroundColors = backgroundColors
	return el
}

func (el *ButtonElement) WithOutlineColors(outlineColors ColorSet) *ButtonElement {
	el.OutlineColors = outlineColors
	return el
}

func (el *ButtonElement) WithClick(click func()) *ButtonElement {
	el.Click = click
	return el
}

func DebugQuickActionModifierHeld() bool {
	return global.DebugAvailable && (rl.IsKeyDown(rl.KeyLeftShift) ||
		rl.IsKeyDown(rl.KeyRightShift) ||
		rl.IsKeyDown(rl.KeyLeftControl) ||
		rl.IsKeyDown(rl.KeyRightControl) ||
		rl.IsKeyDown(rl.KeyLeftSuper) ||
		rl.IsKeyDown(rl.KeyRightSuper))
}

type ButtonElement struct {
	BaseElement[*ButtonElement]
	Text                  string
	TextSize              int32
	Padding, OutlineWidth int32
	ForegroundColors      ColorSet
	BackgroundColors      ColorSet
	OutlineColors         ColorSet
	Click                 func()

	x, y, cx, cy, w, h, textWidth int32

	hovered, hoveredPrevious bool
	clicked, clickedPrevious bool
}

func (el *ButtonElement) prepare() {
	el.textWidth = rl.MeasureText(el.Text, el.TextSize)

	el.w, el.h = max(el.textWidth+el.Padding*2, el.Size().X), max(el.TextSize+el.Padding*2, el.Size().Y)

	pos := el.AbsolutePos()
	el.x, el.y, el.cx, el.cy = pos.X, pos.Y, pos.X+el.w/2, pos.Y+el.h/2
}

func (el *ButtonElement) update(deltaNano int64) {
	if !el.Enabled() || global.UIModalBlocksInput {
		el.hovered = false
		el.clicked = false
		el.hoveredPrevious = false
		el.clickedPrevious = false
		return
	}

	mouseX, mouseY := int32(global.MousePosition.X), int32(global.MousePosition.Y)
	el.hovered = mouseX > el.x &&
		mouseX < el.x+el.w &&
		mouseY > el.y &&
		mouseY < el.y+el.h

	if el.hovered {
		global.MouseCursorState = rl.MouseCursorPointingHand
		global.UIBlocksWorldInput = true
	}

	// Click state machine: track clicked across frames (clickedPrevious -> clicked)
	// so the button stays pressed while mouse is held, and fires Click() on release.
	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		if el.clickedPrevious {
			el.clicked = true
		} else {
			el.clicked = el.hovered

			// play clickdown sound
		}
	} else {
		el.clicked = false

		// Fire Click only if mouse was pressed on this button and released while still hovering
		if el.clickedPrevious && el.hovered && el.Click != nil {
			el.Click()
			//play clickup sound
		}

		// release anywhere
		//if b.clickedPrevious {
		//	b.Click()
		//  // play clickup sound
		//}
	}

	if el.clicked {
		global.UIBlocksWorldInput = true
	}

	// needs to be at the end of the update function
	el.hoveredPrevious, el.clickedPrevious = el.hovered, el.clicked
}

func (el *ButtonElement) draw() {
	btnWidthOuter, btnHeightOuter := el.w+el.OutlineWidth*2, el.h+el.OutlineWidth*2
	btnStartXOuter, btnStartYOuter := el.x-el.OutlineWidth, el.y-el.OutlineWidth

	state := StateDefault
	switch {
	case !el.Enabled():
		state = StateDisabled
	case el.clicked:
		state = StateClick
	case el.hovered:
		state = StateHover
	}

	oCol := el.OutlineColors.Color(state)
	bgCol := el.BackgroundColors.Color(state)
	fgCol := el.ForegroundColors.Color(state)

	opacity := el.Opacity()
	outlineColor := util.ColorOpacity(*oCol, opacity)
	backgroundColor := util.ColorOpacity(*bgCol, opacity)
	foregroundColor := util.ColorOpacity(*fgCol, opacity)

	rl.DrawRectangle(btnStartXOuter, btnStartYOuter, btnWidthOuter, btnHeightOuter, outlineColor)
	rl.DrawRectangle(el.x, el.y, el.w, el.h, backgroundColor)
	rl.DrawText(el.Text, el.cx-el.textWidth/2, el.cy-el.TextSize/2, el.TextSize, foregroundColor)
}
