package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/math"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
)

type ButtonElement struct {
	// Pos and Size do not account for the outline, which is rendered outside this
	// If Size is 0 it will be calculated from TextSize
	Pos                   func(renderWidth, renderHeight int32) math.Vec2i
	Size                  math.Vec2i
	Text                  string
	TextSize              int32
	Padding, OutlineWidth int32
	ForegroundColors      util.ColorSet
	BackgroundColors      util.ColorSet
	OutlineColors         util.ColorSet
	Click                 func()

	x, y, cx, cy, w, h, textWidth int32

	hovered, hoveredPrevious bool
	clicked, clickedPrevious bool
}

func (b *ButtonElement) update(delta float32) {
	b.textWidth = rl.MeasureText(b.Text, b.TextSize)

	pos := b.Pos(int32(rl.GetRenderWidth()), int32(rl.GetRenderHeight()))
	b.x, b.y, b.cx, b.cy = pos.X-b.w/2, pos.Y-b.h/2, pos.X, pos.Y

	b.w, b.h = math.Maxi(b.textWidth+b.Padding*2, b.Size.X), math.Maxi(b.TextSize+b.Padding*2, b.Size.Y)

	mouseX, mouseY := rl.GetMouseX(), rl.GetMouseY()
	b.hovered = mouseX > b.x &&
		mouseX < b.x+b.w &&
		mouseY > b.y &&
		mouseY < b.y+b.h
}

func (b *ButtonElement) draw() {
	btnWidthOuter, btnHeightOuter := b.w+b.OutlineWidth*2, b.h+b.OutlineWidth*2
	btnStartXOuter, btnStartYOuter := b.x-b.OutlineWidth, b.y-b.OutlineWidth

	oCol := fallbackColor(b.OutlineColors.Default, &color.RGBA{})
	bgCol := fallbackColor(b.BackgroundColors.Default, &color.RGBA{})
	fgCol := fallbackColor(b.ForegroundColors.Default, &color.RGBA{})

	if b.hovered {
		oCol = fallbackColor(b.OutlineColors.Hover, oCol)
		bgCol = fallbackColor(b.BackgroundColors.Hover, bgCol)
		fgCol = fallbackColor(b.ForegroundColors.Hover, fgCol)
	}

	if b.clicked {
		oCol = fallbackColor(b.OutlineColors.Click, oCol)
		bgCol = fallbackColor(b.BackgroundColors.Click, bgCol)
		fgCol = fallbackColor(b.ForegroundColors.Click, fgCol)
	}

	rl.DrawRectangle(btnStartXOuter, btnStartYOuter, btnWidthOuter, btnHeightOuter, *oCol)

	rl.DrawRectangle(b.x, b.y, b.w, b.h, *bgCol)

	rl.DrawText(b.Text, b.cx-b.textWidth/2, b.cy-b.TextSize/2, b.TextSize, *fgCol)
}

func fallbackColor(override, fallback *color.RGBA) *color.RGBA {
	if override != nil {
		return override
	}
	return fallback
}
