package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/math"
)

type ButtonElement struct {
	// CenterPos and Size do not account for the outline, which is rendered outside this
	// If Size is 0 it will be calculated from TextSize
	CenterPos, Size       math.Vec2i
	Text                  string
	TextSize              int32
	Padding, OutlineWidth int32
	Colors                ButtonColors
	Click                 func()
}

type ButtonColors struct {
	Outline      color.RGBA
	OutlineHover color.RGBA
	OutlineClick color.RGBA

	Background      color.RGBA
	BackgroundHover color.RGBA
	BackgroundClick color.RGBA

	Foreground      color.RGBA
	ForegroundHover color.RGBA
	ForegroundClick color.RGBA
}

func (b *ButtonElement) Update(delta float32) {}
func (b *ButtonElement) Draw() {
	textWidth := rl.MeasureText(b.Text, b.TextSize)

	btnWidthInner, btnHeightInner := textWidth+b.Padding*2, b.TextSize+b.Padding*2
	w, h := math.Maxi(btnWidthInner, b.Size.X), math.Maxi(btnHeightInner, b.Size.Y)
	btnStartXInner, btnStartYInner := b.CenterPos.X-w/2, b.CenterPos.Y-h/2

	btnWidthOuter, btnHeightOuter := w+b.OutlineWidth*2, h+b.OutlineWidth*2
	btnStartXOuter, btnStartYOuter := b.CenterPos.X-w/2-b.OutlineWidth, b.CenterPos.Y-h/2-b.OutlineWidth

	btnHovered := rl.GetMouseX() > btnStartXInner &&
		rl.GetMouseX() < btnStartXInner+w &&
		rl.GetMouseY() > btnStartYInner &&
		rl.GetMouseY() < btnStartYInner+h

	oCol := b.Colors.Outline
	bgCol := b.Colors.Background
	fgCol := b.Colors.Foreground

	if btnHovered {
		rl.SetMouseCursor(rl.MouseCursorPointingHand)

		oCol = b.Colors.OutlineHover
		bgCol = b.Colors.BackgroundHover
		fgCol = b.Colors.ForegroundHover

		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			oCol = b.Colors.OutlineClick
			bgCol = b.Colors.BackgroundClick
			fgCol = b.Colors.ForegroundClick
		}
	} else {
		rl.SetMouseCursor(rl.MouseCursorDefault)
	}

	rl.DrawRectangle(btnStartXOuter, btnStartYOuter, btnWidthOuter, btnHeightOuter, oCol)

	rl.DrawRectangle(btnStartXInner, btnStartYInner, w, h, bgCol)

	rl.DrawText(b.Text, b.CenterPos.X-textWidth/2, b.CenterPos.Y-b.TextSize/2, b.TextSize, fgCol)
}
