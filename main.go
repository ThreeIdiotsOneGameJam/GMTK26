package main

import (
	"fmt"

	"github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/math"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/world"
)

func update() {
	tick()
	frame()
}

func tick() {

}

var mainScreen = ui.Screen{
	BackgroundColor: rl.RayWhite,
	Elements: []ui.Element{
		&ui.TextElement{
			Pos: func(renderWidth, renderHeight int32) math.Vec2i {
				return math.Vec2i{
					X: renderWidth / 2,
					Y: renderHeight / 4,
				}
			},
			Text:      "Game",
			TextSize:  96,
			TextColor: rl.Black,
		},
		&ui.ButtonElement{
			Pos: func(renderWidth, renderHeight int32) math.Vec2i {
				return math.Vec2i{
					X: renderWidth / 2,
					Y: renderHeight / 2,
				}
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
			Click: func() {
				fmt.Println("clicked")
			},
		},
	},
}

func frame() {
	rl.BeginDrawing()

	mainScreen.Update(0)
	mainScreen.Draw()

	world.Draw()

	rl.EndDrawing()
}

var updateFunc = update

func main() {
	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagWindowResizable)
	rl.InitWindow(1200, 675, "Game")
	defer rl.CloseWindow()

	world.Init()

	mainLoop()
}
