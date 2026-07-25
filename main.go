package main

import (
	"fmt"
	// "math"
	"strconv"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/audio"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/screens"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/rlvec"
)

func update() {
	tick()
	frame()
}

func tick() {
	if rl.IsKeyPressed(rl.KeyEscape) {
		//screens.SetActiveScreen(screens.MainScreenID)
		screens.ToggleEscScreen()
	}

	if rl.IsKeyPressed(rl.KeyF11) {
		rl.ToggleFullscreen()
	}

	if rl.IsKeyPressed(rl.KeyF3) {
		global.DebugEnabled = !global.DebugEnabled
	}

	net.DrainEvents(handleServerPacket)
}

func handleServerPacket(packet packets.S2CPacket) {
	switch p := packet.(type) {
	case *packets.S2CConnectAcceptedPacket:
		fmt.Printf("connected as %s (persistent: %t)\n", p.ClientID, p.Persistent)
		screens.ResetGameSession()
	case *packets.S2CGameJoinedPacket:
		fmt.Printf("joined game %d with code %s\n", p.Game.GameID, p.Game.GameCode)
		screens.ClearGameCodeInput()
		screens.EnterGame(p.Game)
	case *packets.S2CGameUpdatePacket:
		screens.ApplyGameUpdate(p.Game)
	case *packets.S2CGameRejectedPacket:
		fmt.Printf("game %s rejected: %s\n", p.Operation, p.Message)
		if p.Operation == "create" {
			screens.RejectGameCreation(p.Message)
		}
	case *packets.S2CGameClosedPacket:
		fmt.Printf("game %d closed: %s\n", p.GameID, p.Reason)
		screens.CloseGame(p.GameID)
	default:
		fmt.Printf("received unhandled packet type %T\n", packet)
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

	audio.Update()

	rl.BeginDrawing()

	rl.ClearBackground(rl.RayWhite)

	global.MouseCursorState = rl.MouseCursorDefault

	screens.Update(deltaTime.Nanoseconds())
	screens.Draw()

	if global.DebugEnabled {
		util.DrawTextSimple("FPS: "+strconv.FormatFloat(fps, 'f', 2, 64), 10, 10)
		util.DrawTextSimple("Runtime: "+time.Now().Sub(startTime).Round(time.Second).String(), 10, 20)
		util.DrawTextSimple("WS: "+net.State().String(), 10, 30)
	}

	rl.EndDrawing()

	global.MousePosition = rlvec.FromRL(rl.GetMousePosition())

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
	err := game.LoadOrCreatePlayerData()
	if err != nil {
		fmt.Println("failed to load or create player data, using ephemeral values")
	}

	var configFlags uint32 = rl.FlagWindowResizable

	if fpsTarget == 0 {
		configFlags |= rl.FlagVsyncHint
	}

	rl.SetConfigFlags(configFlags)

	rl.InitWindow(1200, 675, "Game")
	defer rl.CloseWindow()

	go net.Connect("localhost:58008")
	defer net.Close()

	rl.SetExitKey(rl.KeyNull)

	audio.Init()
	defer audio.Terminate()

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
