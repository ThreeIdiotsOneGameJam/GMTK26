package main

import (
	"fmt"

	"github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/math"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
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
			ForegroundColors: util.ColorSet{
				Default: &rl.DarkGray,
			},
			BackgroundColors: util.ColorSet{
				Default: &rl.LightGray,
				Hover:   math.ColorAdd(rl.LightGray, 25),
				Click:   math.ColorAdd(rl.LightGray, 40),
			},
			OutlineColors: util.ColorSet{
				Default: &rl.Gray,
			},
			Click: func() {
				fmt.Println("clicked")
			},
		},
	},
}

func frame() {
	rl.BeginDrawing()

	rl.SetMouseCursor(rl.MouseCursorDefault)
	mainScreen.Update(0)
	mainScreen.Draw()

	rl.EndDrawing()
}

var updateFunc = update

func main() {
	rl.SetConfigFlags(rl.FlagVsyncHint | rl.FlagWindowResizable)
	rl.InitWindow(1200, 675, "Game")
	defer rl.CloseWindow()

	mainLoop()
}
