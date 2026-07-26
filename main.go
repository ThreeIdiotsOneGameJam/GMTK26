package main

import (
	"flag"
	"fmt"
	// "math"
	"os"

	"strconv"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
	"github.com/threeidiotsonegamejam/gmtk26/src/audio"
	"github.com/threeidiotsonegamejam/gmtk26/src/constants"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/render/shaders"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
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
		screens.HandleEscape()
	}

	if rl.IsKeyPressed(rl.KeyF11) {
		rl.ToggleFullscreen()
	}

	if global.DebugAvailable && rl.IsKeyPressed(rl.KeyF3) {
		global.DebugEnabled = !global.DebugEnabled
	}

	if rl.IsKeyPressed(rl.KeySpace) {
		screens.HandleTownHallShortcut()
	}

	net.DrainEvents(handleServerPacket)
}

func handleServerPacket(packet packets.S2CPacket) {
	switch p := packet.(type) {
	case *packets.S2CConnectAcceptedPacket:
		fmt.Printf("connected as %s (persistent: %t)\n", p.ClientID, p.Persistent)
		screens.SetLocalClientID(p.ClientID)
		screens.ResetGameSession()
	case *packets.S2CGameJoinedPacket:
		fmt.Printf("joined game %d with code %s\n", p.Game.GameID, p.Game.GameCode)
		screens.ClearGameCodeInput()
		screens.EnterGame(p.Game)
	case *packets.S2CMatchmakingWaitingPacket:
		fmt.Printf("matchmaking queue position %d\n", p.QueuePosition)
		screens.ApplyMatchmakingWaiting(p.QueuePosition)
	case *packets.S2CGameUpdatePacket:
		screens.ApplyGameUpdate(p.Game)
	case *packets.S2CGameRejectedPacket:
		fmt.Printf("game %s rejected: %s\n", p.Operation, p.Message)
		switch p.Operation {
		case "create":
			screens.RejectGameCreation(p.Message)
		case "join":
			screens.RejectGameJoin(p.Message)
		case "start":
			screens.RejectGameStart(p.Message)
		}
	case *packets.S2CGameClosedPacket:
		fmt.Printf("game %d closed: %s\n", p.GameID, p.Reason)
		screens.CloseGame(p.GameID)
	case *packets.S2CGameStartPacket:
		net.LocalGameState.ApplyStartPacket(p)
		screens.ApplyServerGameStart(p)
		fmt.Printf("game started as faction %d, round %d\n", p.FactionIdx, p.Round)
	case *packets.S2CGameStatePacket:
		net.LocalGameState.ApplyStatePacket(p)
		screens.ApplyServerGameState(p)
	case *packets.S2CGameEndPacket:
		net.LocalGameState.ApplyEndPacket()
		screens.ApplyServerGameEnd(p)
		fmt.Printf("game ended! winner: %s (faction %d)\n", p.WinnerName, p.WinnerFaction)
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
	global.UIBlocksWorldInput = false
	global.UIModalBlocksInput = false

	screens.Update(deltaTime.Nanoseconds())
	screens.Draw()

	if global.DebugEnabled {
		util.DrawTextSimple("Render Size: "+fmt.Sprintf("%d x %d", rl.GetRenderWidth(), rl.GetRenderHeight()), 10, 10)
		util.DrawTextSimple("FPS: "+strconv.FormatFloat(fps, 'f', 2, 64), 10, 20)
		util.DrawTextSimple("Runtime: "+time.Now().Sub(startTime).Round(time.Second).String(), 10, 30)
		util.DrawTextSimple("WS: "+net.State().String(), 10, 40)
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
	guest := flag.Bool("guest", false, "use a temporary guest identity (not saved)")
	uuidFlag := flag.String("uuid", "", `session client ID: "random" or a canonical UUID (not saved)`)
	debug := flag.Bool("debug", false, "allow toggling the debug overlay with F3")
	flag.Parse()

	if *guest && *uuidFlag != "" {
		fmt.Fprintln(os.Stderr, "flags --guest and --uuid are mutually exclusive")
		os.Exit(2)
	}

	global.DebugAvailable = *debug

	if err := settings.Load(); err != nil {
		fmt.Printf("failed to load settings, using defaults: %v\n", err)
	}

	switch {
	case *guest:
		game.UseGuestIdentity()
	case *uuidFlag != "":
		id, err := parseLaunchUUID(*uuidFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		game.UseEphemeralUUID(id)
	default:
		if err := game.LoadOrCreatePlayerData(); err != nil {
			fmt.Println("failed to load or create player data, using ephemeral values")
		}
	}

	var configFlags uint32 = rl.FlagWindowResizable

	if fpsTarget == 0 {
		configFlags |= rl.FlagVsyncHint
	}

	rl.SetConfigFlags(configFlags)

	rl.InitWindow(constants.WindowWidth, constants.WindowHeight, constants.GameName)
	defer rl.CloseWindow()
	rl.SetWindowMinSize(768, 576)
	icon := rl.LoadImage(global.AssetDir + "/textures/icon.png")
	if icon != nil {
		rl.SetWindowIcon(*icon)
		defer rl.UnloadImage(icon)
	}

	shaders.Load()
	defer shaders.Unload()

	go net.Connect(constants.DefaultServerAddrs()...)
	defer net.Close()

	rl.SetExitKey(rl.KeyNull)

	audio.Init()
	defer audio.Terminate()
	defer screens.Shutdown()

	mainLoop()
}

func parseLaunchUUID(value string) (uuid.UUID, error) {
	if value == "random" {
		id, err := uuid.NewRandom()
		if err != nil {
			return uuid.Nil, fmt.Errorf("generate random uuid: %w", err)
		}
		return id, nil
	}

	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil || id.String() != value {
		return uuid.Nil, fmt.Errorf(`--uuid must be "random" or a canonical non-nil UUID`)
	}
	return id, nil
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
