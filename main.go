package main

import (
	"github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/math"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
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

	btn := ui.ButtonElement{
		CenterPos: math.Vec2i{
			X: midX,
			Y: midY,
		},
		Text:         "Play",
		TextSize:     48,
		Padding:      8,
		OutlineWidth: 4,
		Colors: ui.ButtonColors{
			Outline:         rl.Gray,
			Background:      rl.LightGray,
			BackgroundHover: math.ColorAdd(rl.LightGray, 25),
			BackgroundClick: math.ColorAdd(rl.LightGray, 40),
			Foreground:      rl.DarkGray,
		},
		Click: func() {},
	}
	btn.Draw()

	rl.EndDrawing()
}

var updateFunc = update

func main() {
	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagWindowResizable)
	rl.InitWindow(1200, 675, "Game")
	defer rl.CloseWindow()

	mainLoop()
}
