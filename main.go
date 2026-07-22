package main

import (
	"github.com/gen2brain/raylib-go/raylib"
)

func update() {
	tick()
	frame()
}

func tick() {

}

func frame() {
	rl.BeginDrawing()

	rl.ClearBackground(rl.RayWhite)

	text, textHeight := "Game", int32(96)
	textWidth := rl.MeasureText(text, textHeight)
	midX, midY := int32(rl.GetRenderWidth()/2), int32(rl.GetRenderHeight()/2)

	rl.DrawText(text, midX-(textWidth/2), int32(rl.GetRenderHeight()/4), textHeight, rl.Black)

	btnText, btnTextHeight, btnPadding, btnOutlineWidth := "Play", int32(48), int32(8), int32(4)
	btnTextWidth := rl.MeasureText(btnText, btnTextHeight)

	btnWidthInner, btnHeightInner := btnTextWidth+btnPadding*2, btnTextHeight+btnPadding*2
	btnStartXInner, btnStartYInner := midX-btnWidthInner/2, midY-btnHeightInner/2

	btnWidthOuter, btnHeightOuter := btnWidthInner+btnOutlineWidth*2, btnHeightInner+btnOutlineWidth*2
	btnStartXOuter, btnStartYOuter := midX-btnWidthInner/2-btnOutlineWidth, midY-btnHeightInner/2-btnOutlineWidth

	btnHovered := rl.GetMouseX() > btnStartXInner &&
		rl.GetMouseX() < btnStartXInner+btnWidthInner &&
		rl.GetMouseY() > btnStartYInner &&
		rl.GetMouseY() < btnStartYInner+btnHeightInner

	btnColorInner := rl.LightGray
	if btnHovered {
		rl.SetMouseCursor(rl.MouseCursorPointingHand)

		colorDelta := uint8(25)

		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			colorDelta += 15
		}

		btnColorInner.R += colorDelta
		btnColorInner.G += colorDelta
		btnColorInner.B += colorDelta
	} else {
		rl.SetMouseCursor(rl.MouseCursorDefault)
	}

	rl.DrawRectangle(btnStartXOuter, btnStartYOuter, btnWidthOuter, btnHeightOuter, rl.Gray)

	rl.DrawRectangle(btnStartXInner, btnStartYInner, btnWidthInner, btnHeightInner, btnColorInner)

	rl.DrawText(btnText, midX-btnTextWidth/2, midY-btnTextHeight/2, btnTextHeight, rl.DarkGray)

	rl.EndDrawing()
}

var updateFunc = update

func main() {
	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagWindowResizable)
	rl.InitWindow(1200, 675, "Game")
	defer rl.CloseWindow()

	mainLoop()
}

func DrawTextSimple(text string, x int32, y int32) {
	rl.DrawText(text, x, y, 10, rl.Black)
}
