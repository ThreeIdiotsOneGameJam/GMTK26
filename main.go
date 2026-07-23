package main

import (
	"math"
	"strconv"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/mathutil/vec"
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
var viewport rl.RenderTexture2D

func frame() {
	currentTime := time.Now()
	deltaTime := currentTime.Sub(lastFrameTime)
	lastFrameTime = currentTime

	fps := 0.0

	if deltaTime > 0 {
		fps = float64(time.Second) / float64(deltaTime)
	}

	rl.BeginTextureMode(viewport)

	rl.ClearBackground(rl.RayWhite)

	global.MouseCursorState = rl.MouseCursorDefault

	screen := screens.GetActiveScreen()
	screen.Update(deltaTime.Nanoseconds())
	screen.Draw()

	util.DrawTextSimple("FPS: "+strconv.FormatFloat(fps, 'f', 2, 64), 10, 10)
	util.DrawTextSimple("Runtime: "+time.Now().Sub(startTime).Round(time.Second).String(), 10, 20)

	rl.EndTextureMode()

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	if viewport.Texture.Width != global.ViewportSize.X {
		rl.UnloadRenderTexture(viewport)
		viewport = rl.LoadRenderTexture(global.ViewportSize.X, global.ViewportSize.Y)
	}

	screenW := float32(rl.GetScreenWidth())
	screenH := float32(rl.GetScreenHeight())

	viewW := float32(viewport.Texture.Width)
	viewH := float32(viewport.Texture.Height)

	srcRect := rl.Rectangle{
		X:      0.0,
		Y:      0.0,
		Width:  viewW,
		Height: -viewH,
	}
	ratio := float32(int32(math.Min(
		float64(screenW/viewW),
		float64(screenH/viewH),
	))) + 1
	dstRect := rl.Rectangle{
		X:      (screenW - viewW*ratio) / 2.0,
		Y:      (screenH - viewH*ratio) / 2.0,
		Width:  viewW * ratio,
		Height: viewH * ratio,
	}
	rl.DrawTexturePro(viewport.Texture, srcRect, dstRect, rl.Vector2{}, 0.0, rl.White)
	rl.EndDrawing()

	mouse := rl.GetMousePosition()
	global.MousePosition = vec.Vec2{
		X: (mouse.X - dstRect.X) * (srcRect.Width / dstRect.Width),
		Y: (mouse.Y - dstRect.Y) * (-srcRect.Height / dstRect.Height),
	}

	rl.SetMouseCursor(global.MouseCursorState)

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

	viewport = rl.LoadRenderTexture(global.ViewportSize.X, global.ViewportSize.Y)
	defer rl.UnloadRenderTexture(viewport)

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
