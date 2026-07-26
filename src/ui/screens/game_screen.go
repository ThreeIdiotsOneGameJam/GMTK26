package screens

import (
	"fmt"
	"image/color"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/audio"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	gameNet "github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

const gameCodeCopyFeedbackDuration = 1500 * time.Millisecond

var gameSeedInput = ui.Input()
var gameWorld = ui.GameWorld()
var gameRegenerateButton = ui.Button()
var currentGame *game.Game
var gamePreviousScreen *ui.ScreenElement
var gameLeaveTransition bool

// localClientID is the server-assigned identity from the connect handshake,
// used to decide whether the local player hosts the current game.
var localClientID game.ClientID

// serverGameActive is true between S2CGameStartPacket and S2CGameEndPacket
// for both remote multiplayer and the in-process solo server.
var serverGameActive bool
var serverRound int32
var serverCoins int32
var serverPoints int32
var serverResources game.Resources
var serverDeadline int64
var gameOverMessage string
var gameActionError string

func init() {
	// In a running authoritative game, building placement goes through the
	// server instead of mutating the local map; the next state broadcast
	// carries the result.
	gameWorld.Renderer.OnPlaceBuilding = func(hex game.Hex, building game.BuildingType) bool {
		if !serverGameActive {
			return currentGame != nil &&
				(currentGame.Multiplayer || gameNet.LocalGameActive())
		}
		if err := gameNet.SendBuildAction(serverRound, hex, building); err != nil {
			fmt.Printf("failed to send build action: %v\n", err)
		} else {
			// Pending actions are intentionally client-only: opponents should
			// not see a building until the server resolves the round.
			gameWorld.Renderer.QueueBuilding(hex, building)
		}
		return true
	}
}

func SetLocalClientID(id game.ClientID) {
	localClientID = id
}

func isLocalHost() bool {
	return currentGame != nil && currentGame.HostID == localClientID
}

// ApplyServerGameStart switches the pre-game session into the running game
// using the authoritative state from the server.
func ApplyServerGameStart(p *packets.S2CGameStartPacket) {
	if currentGame == nil {
		return
	}
	serverGameActive = true
	gameOverMessage = ""
	gameWorld.Renderer.ClearQueuedBuilding()
	applyServerRound(p.Round, p.Deadline, p.Map, p.Coins, p.Points, p.Resources)
	focusOnTownhall()
}

func ApplyServerGameState(p *packets.S2CGameStatePacket) {
	if !serverGameActive {
		return
	}
	if p.Round != serverRound {
		gameWorld.Renderer.ClearQueuedBuilding()
	}
	applyServerRound(p.Round, p.Deadline, p.Map, p.Coins, p.Points, p.Resources)
}

func ApplyServerGameEnd(p *packets.S2CGameEndPacket) {
	if !serverGameActive {
		return
	}
	serverGameActive = false
	gameWorld.Renderer.ClearQueuedBuilding()
	if p.WinnerName != "" {
		gameOverMessage = "Game over! Winner: " + p.WinnerName
	} else {
		gameOverMessage = "Game over!"
	}
}

func applyServerRound(round int32, deadline int64, m game.Map, coins, points int32, resources game.Resources) {
	serverRound = round
	serverDeadline = deadline
	serverCoins = coins
	serverPoints = points
	serverResources = resources
	currentGame.Round = round
	currentGame.Map = m
	gameWorld.Map = m
	gameSeedInput.Text = strconv.FormatInt(m.Seed, 10)
}

func multiplayerStatusText() string {
	if currentGame == nil {
		return ""
	}
	if gameOverMessage != "" {
		return gameOverMessage
	}
	if serverGameActive {
		remaining := max(time.Until(time.Unix(0, serverDeadline)), 0)
		return fmt.Sprintf(
			"Round %d | Coins: %d | Points: %d | Next round: %ds",
			serverRound,
			serverCoins,
			serverPoints,
			int(remaining.Seconds()),
		)
	}
	if !currentGame.Multiplayer {
		return "Starting solo game..."
	}

	players := 0
	for i := range currentGame.Factions {
		if currentGame.Factions[i].Player != nil {
			players++
		}
	}
	if isLocalHost() {
		return fmt.Sprintf("%d/%d players joined", players, currentGame.MaxPlayers)
	}
	return fmt.Sprintf(
		"%d/%d players | Waiting for the host to start",
		players,
		currentGame.MaxPlayers,
	)
}

func EnterGame(state game.Game) {
	clearMatchmaking()
	serverGameActive = false
	gameOverMessage = ""
	gameActionError = ""
	gameWorld.Renderer.ClearQueuedBuilding()
	applyGameState(state)
	gameWorld.Renderer.ResetCamera(&gameWorld.Map)
	if screenIsActiveOrPending(gameScreen) {
		return
	}
	// Return to the play menu on close, not intermediate creation/matchmaking screens.
	SetActiveScreen(NewGameScreen(playScreenOrNew()))
}

func RejectGameStart(message string) {
	gameActionError = capitalizeSentence(message)
}

func capitalizeSentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

func RejectGameJoin(message string) {
	clearMatchmaking()
	SetPlayError(message)
	SetActiveScreen(playScreenOrNew())
}

func ApplyGameUpdate(state game.Game) {
	if currentGame == nil || currentGame.GameID != state.GameID {
		return
	}
	applyGameState(state)
}

func CloseGame(gameID uint64) {
	if currentGame == nil || currentGame.GameID != gameID {
		return
	}
	if gameLeaveTransition && activeScreen == gameScreen && pendingScreen != nil {
		// The local leave action already owns the transition and destination.
		// Keep its outgoing game frame intact until the crossfade completes.
		return
	}
	clearCurrentGame()
	GoToPreviousScreen(gamePreviousScreen)
}

func ResetGameSession() {
	wasMatchmaking := matchmakingActive
	clearMatchmaking()
	if currentGame == nil || !currentGame.Multiplayer {
		if wasMatchmaking {
			SetActiveScreen(playScreenOrNew())
		}
		return
	}
	clearCurrentGame()
	GoToPreviousScreen(gamePreviousScreen)
}

func LeaveCurrentGame() {
	if matchmakingActive {
		if err := gameNet.Send(&packets.C2SLeaveGamePacket{}); err != nil {
			fmt.Printf("failed to cancel matchmaking: %v\n", err)
		}
		clearMatchmaking()
	}
	if currentGame == nil {
		return
	}
	gameLeaveTransition = true
	if gameNet.LocalGameActive() {
		gameNet.StopLocalGame()
	} else if currentGame.Multiplayer {
		if err := gameNet.Send(&packets.C2SLeaveGamePacket{}); err != nil {
			println("failed to leave game:", err.Error())
		}
	}
}

func applyGameState(state game.Game) {
	nextMap := state.Map
	if currentGame != nil &&
		currentGame.GameID == state.GameID &&
		gameWorld.Map.Seed == nextMap.Seed &&
		nextMap.Grid == nil {
		nextMap = gameWorld.Map
	}
	if nextMap.Grid == nil {
		nextMap.Generate()
	}

	state.Map = nextMap
	gameWorld.Map = nextMap
	gameSeedInput.Text = strconv.FormatInt(nextMap.Seed, 10)
	currentGame = &state
}

func clearCurrentGame() {
	currentGame = nil
	gameWorld.Map = game.Map{}
	gameWorld.Renderer.ClearQueuedBuilding()
	gameLeaveTransition = false
	serverGameActive = false
	gameOverMessage = ""
	gameActionError = ""
	serverResources = nil
	gameNet.LocalGameState.Reset()
}

func setBuildingClick(building game.BuildingType) func() {
	return func() {
		gameWorld.Renderer.BuildingToPlace = building
	}
}

func focusOnTownhall() {
	for x := range gameWorld.Map.Grid {
		for y := range gameWorld.Map.Grid[x] {
			cell := &gameWorld.Map.Grid[x][y]
			if cell.Owner == int8(gameNet.LocalGameState.FactionIdx) && cell.Building == game.BuildingTownhall {
				gameWorld.Renderer.FocusOnHex(game.NewHex(int32(x), int32(y)))
				return
			}
		}
	}
}

func NewGameScreen(previousScreen *ui.ScreenElement) *ui.ScreenElement {
	gamePreviousScreen = previousScreen
	// nil previousScreen → Leave Game creates a fresh main screen.
	escScreen = NewEscScreen(nil)

	var copiedGameCode string
	var gameCodeCopiedUntil time.Time

	gameCodeButtonText := func() string {
		if currentGame == nil || !currentGame.Multiplayer || currentGame.GameCode == "" {
			return ""
		}
		if currentGame.GameCode == copiedGameCode && time.Now().Before(gameCodeCopiedUntil) {
			return "Copied!"
		}
		return "Game code: " + currentGame.GameCode
	}

	transparent := color.RGBA{}
	gameCodeButton := ui.Button().
		WithText(gameCodeButtonText()).
		WithTextSize(28).
		WithPadding(0).
		WithOutlineWidth(0).
		WithForegroundColors(ui.ColorSet{
			Default: &rl.Black,
			Hover:   &ui.PaletteIndigo,
			Click:   &ui.PaletteIndigoPress,
		}).
		WithBackgroundColors(ui.ColorSet{Default: &transparent}).
		WithOutlineColors(ui.ColorSet{Default: &transparent}).
		WithAnchors(anchor.TopRight, anchor.TopRight).
		WithRelativePos(vec.Vec2i{X: -20, Y: 20}).
		WithVisibleDynamic(func(el *ui.ButtonElement) bool {
			return currentGame != nil && currentGame.Multiplayer && currentGame.GameCode != ""
		}).
		WithClick(func() {
			if currentGame == nil || !currentGame.Multiplayer || currentGame.GameCode == "" {
				return
			}

			copiedGameCode = currentGame.GameCode
			gameCodeCopiedUntil = time.Now().Add(gameCodeCopyFeedbackDuration)
			rl.SetClipboardText(copiedGameCode)
		})

	screen := ui.Screen().
		WithEnter(func() {
			HideEscScreen()
			gameWorld.Renderer.ResetCamera(&gameWorld.Map)

			audio.StartMusic()
			audio.StartAmbience()
		}).
		WithExit(func() {
			HideEscScreen()
			gameWorld.Renderer.ResetCamera(&gameWorld.Map)
			audio.StopMusic()
			audio.StopAmbience()
			clearCurrentGame()
		}).
		AddChild(
			gameWorld,
		).
		AddChild(
			ui.Group().WithUpdate(func(_ int64) {
				gameCodeButton.Text = gameCodeButtonText()
			}),
		).
		AddChild(gameCodeButton).
		AddChild(
			ui.Text().
				WithTextDynamic(multiplayerStatusText).
				WithTextSize(26).
				WithTextColor(rl.Black).
				WithAnchors(anchor.Top, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 20}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return multiplayerStatusText() != ""
				}),
		).
		AddChild(
			ui.Button().
				WithText("Start Game").
				WithTextSize(30).
				WithPadding(10).
				WithOutlineWidth(4).
				WithAnchors(anchor.Top, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 58}).
				WithVisibleDynamic(func(el *ui.ButtonElement) bool {
					return currentGame != nil && currentGame.Multiplayer &&
						!serverGameActive && gameOverMessage == "" && isLocalHost()
				}).
				WithClick(func() {
					gameActionError = ""
					if err := gameNet.Send(&packets.C2SStartGamePacket{}); err != nil {
						gameActionError = capitalizeSentence(err.Error())
						fmt.Printf("failed to request game start: %v\n", err)
					}
				}),
		).
		AddChild(
			ui.Text().
				WithTextDynamic(func() string { return gameActionError }).
				WithTextSize(22).
				WithTextColor(color.RGBA{R: 255, G: 96, B: 96, A: 255}).
				WithTextShadow(color.RGBA{R: 0, G: 0, B: 0, A: 200}, vec.Vec2i{X: 2, Y: 2}).
				WithAnchors(anchor.Top, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 122}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return gameActionError != ""
				}),
		).
		AddChild(
			ui.Group().
				WithAnchors(anchor.BottomLeft, anchor.BottomLeft).
				WithRelativePos(vec.Vec2i{X: 8, Y: -48}).
				AddChild(
					ui.Text().
						WithTextSize(24).
						WithRelativePos(vec.Vec2i{X: 0, Y: -32}).
						WithTextColor(rl.Black).
						WithText("Buildings:"),
				).
				AddChild(
					ui.Button().
						WithPadding(8).
						WithTextSize(24).
						WithText("None").
						WithClick(setBuildingClick(game.BuildingUnknown)),
				).
				AddChild(
					ui.Button().
						WithPadding(8).
						WithTextSize(24).
						WithText("Barracks").
						WithRelativePos(vec.Vec2i{X: 84, Y: 0}).
						WithClick(setBuildingClick(game.BuildingBarracks)),
				).
				AddChild(
					ui.Button().
						WithPadding(8).
						WithTextSize(24).
						WithText("Farm").
						WithRelativePos(vec.Vec2i{X: 224, Y: 0}).
						WithClick(setBuildingClick(game.BuildingFarm)),
				).
				AddChild(
					ui.Button().
						WithPadding(8).
						WithTextSize(24).
						WithText("Mine").
						WithRelativePos(vec.Vec2i{X: 308, Y: 0}).
						WithClick(setBuildingClick(game.BuildingMine)),
				).
				AddChild(
					ui.Button().
						WithPadding(8).
						WithTextSize(24).
						WithText("Forester").
						WithRelativePos(vec.Vec2i{X: 384, Y: 0}).
						WithClick(setBuildingClick(game.BuildingForester)),
				),
		).
		AddChild(
			ui.Group().
				WithAnchors(anchor.TopLeft, anchor.TopLeft).
				WithRelativePos(vec.Vec2i{X: 8, Y: 8}).
				WithVisibleDynamic(func(el *ui.GroupElement) bool {
					return !serverGameActive && global.DebugEnabled
				}).
				AddChild(
					gameSeedInput.
						WithPadding(8).
						WithTextSize(24).
						WithSize(vec.Vec2i{X: 320, Y: 0}).
						WithPlaceholderText("Seed").
						WithDefaultText(""),
				).
				AddChild(
					gameRegenerateButton.
						WithPadding(8).
						WithTextSize(24).
						WithRelativePos(vec.Vec2i{X: 0, Y: 52}).
						WithText("Regenerate").
						WithClick(func() {
							gameWorld.Map.Seed = gameSeedFromText(gameSeedInput.Text)
							gameWorld.Map.Generate()
						}),
				).
				AddChild(
					ui.Button().
						WithPadding(8).
						WithTextSize(24).
						WithRelativePos(vec.Vec2i{X: 0, Y: 104}).
						WithText("Random").
						WithClick(func() {
							gameSeedInput.Text = strconv.FormatInt(rand.Int63(), 10)
							gameRegenerateButton.Click()
						}),
				),
		).
		AddChild(
			ui.Group().
				WithAnchors(anchor.TopLeft, anchor.TopLeft).
				WithRelativePos(vec.Vec2i{X: 8, Y: 8}).
				WithVisibleDynamic(func(el *ui.GroupElement) bool {
					return serverGameActive
				}).
				AddChild(
					ui.Text().
						WithTextDynamic(func() string {
							return fmt.Sprintf("Wood: %d", serverResources[game.ResourceWood])
						}).
						WithTextSize(20).
						WithTextColor(rl.Black).
						WithRelativePos(vec.Vec2i{X: 0, Y: 24}),
				).
				AddChild(
					ui.Text().
						WithTextDynamic(func() string {
							return fmt.Sprintf("Stone: %d", serverResources[game.ResourceStone])
						}).
						WithTextSize(20).
						WithTextColor(rl.Black).
						WithRelativePos(vec.Vec2i{X: 0, Y: 48}),
				).
				AddChild(
					ui.Text().
						WithTextDynamic(func() string {
							return fmt.Sprintf("Coal: %d", serverResources[game.ResourceCoal])
						}).
						WithTextSize(20).
						WithTextColor(rl.Black).
						WithRelativePos(vec.Vec2i{X: 0, Y: 72}),
				).
				AddChild(
					ui.Text().
						WithTextDynamic(func() string {
							return fmt.Sprintf("Iron: %d", serverResources[game.ResourceIron])
						}).
						WithTextSize(20).
						WithTextColor(rl.Black).
						WithRelativePos(vec.Vec2i{X: 0, Y: 96}),
				).
				AddChild(
					ui.Text().
						WithTextDynamic(func() string {
							return fmt.Sprintf("Steel: %d", serverResources[game.ResourceSteel])
						}).
						WithTextSize(20).
						WithTextColor(rl.Black).
						WithRelativePos(vec.Vec2i{X: 0, Y: 120}),
				).
				AddChild(
					ui.Text().
						WithTextDynamic(func() string {
							return fmt.Sprintf("Gold: %d", serverResources[game.ResourceGold])
						}).
						WithTextSize(20).
						WithTextColor(rl.Black).
						WithRelativePos(vec.Vec2i{X: 0, Y: 144}),
				),
		).
		AddChild(
			ui.Button().
				WithText("Focus").
				WithTextSize(18).
				WithPadding(6).
				WithOutlineWidth(2).
				WithAnchors(anchor.BottomRight, anchor.BottomRight).
				WithRelativePos(vec.Vec2i{X: -20, Y: -20}).
				WithVisibleDynamic(func(el *ui.ButtonElement) bool {
					return serverGameActive
				}).
				WithClick(focusOnTownhall),
		).
		AddChild(
			ui.GameBuildingDetailsPanel().
				WithWorld(gameWorld),
		).
		AddChild(
			ui.Vignette().WithAlpha(120),
		).
		AddChild(
			escScreen,
		)

	gameScreen = screen
	return screen
}
