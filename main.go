package main

import (
	"strconv"
	"time"

	"github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/screens"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
)

func update() {
	tick()
	frame()
}

func tick() {
	if rl.IsKeyPressed(rl.KeyEscape) {
		screens.SetActiveScreen(screens.MainScreenID)
	}
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

	global.MouseCursorState = rl.MouseCursorDefault

	screens.GetActiveScreen().Update(deltaTime.Nanoseconds())
	screens.GetActiveScreen().Draw()

	rl.SetMouseCursor(global.MouseCursorState)

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

	rl.SetExitKey(rl.KeyNull)

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
