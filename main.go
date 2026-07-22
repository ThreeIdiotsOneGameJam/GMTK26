package main

import (
	"fmt"
	"strconv"
	"time"

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

var startTime = time.Now()

// fpsTarget = -1 -> unlimited, fpsTarget = 0 -> vsync
var fpsTarget float64 = 0

var lastFrameTime = startTime
var frameCount uint64 = 0

func frame() {
	currentTime := time.Now()
	deltaTime := currentTime.Sub(lastFrameTime)
	lastFrameTime = currentTime

	fps := 0.0

	if deltaTime > 0 {
		fps = float64(time.Second) / float64(deltaTime)
	}

	util.DrawTextSimple("FPS: "+strconv.FormatFloat(fps, 'f', 2, 64), 10, 10)
	util.DrawTextSimple("Runtime: "+time.Now().Sub(startTime).Round(time.Second).String(), 10, 20)

	rl.BeginDrawing()

	rl.SetMouseCursor(rl.MouseCursorDefault)
	mainScreen.Update(deltaTime.Nanoseconds())
	mainScreen.Draw()

	rl.EndDrawing()

	frameCount++

	if fpsTarget > 0 {
		targetFrameTime := time.Duration(float64(time.Second) / fpsTarget)
		elapsed := time.Since(currentTime)

		if remaining := targetFrameTime - elapsed; remaining > 0 {
			time.Sleep(remaining)
		}
	}
}

var updateFunc = update

func main() {
	var configFlags uint32 = rl.FlagWindowResizable

	if fpsTarget == 0 {
		configFlags |= rl.FlagVsyncHint
	}

	rl.SetConfigFlags(configFlags)

	rl.InitWindow(1200, 675, "Game")
	defer rl.CloseWindow()

	mainLoop()
}

func toggleVsync() {
	toggleWindowState(rl.FlagVsyncHint)
}

func toggleWindowState(flag uint32) {
	if rl.IsWindowState(flag) {
		rl.ClearWindowState(flag)
	} else {
		rl.SetWindowState(flag)
	}
}
