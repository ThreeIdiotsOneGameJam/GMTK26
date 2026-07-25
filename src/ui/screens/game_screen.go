package screens

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/audio"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	gameNet "github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var gameSeedInput = ui.Input()
var gameWorld = ui.World()
var gameRegenerateButton = ui.Button()
var currentGame *game.Game

// localClientID is the server-assigned identity from the connect handshake,
// used to decide whether the local player hosts the current game.
var localClientID game.ClientID

// serverGameActive is true between S2CGameStartPacket and S2CGameEndPacket.
var serverGameActive bool
var serverRound int32
var serverCoins int32
var serverPoints int32
var serverResources game.Resources
var serverDeadline int64
var gameOverMessage string
var gameActionError string

func init() {
	// In a running multiplayer game, building placement goes through the
	// server instead of mutating the local map; the next state broadcast
	// carries the result.
	gameWorld.Renderer.OnPlaceBuilding = func(hex game.Hex, building game.BuildingType) bool {
		if !serverGameActive {
			return false
		}
		if err := gameNet.SendBuildAction(serverRound, hex, building); err != nil {
			fmt.Printf("failed to send build action: %v\n", err)
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
	if currentGame == nil || !currentGame.Multiplayer {
		return
	}
	serverGameActive = true
	gameOverMessage = ""
	applyServerRound(p.Round, p.Deadline, p.Map, p.Coins, p.Points, p.Resources)
}

func ApplyServerGameState(p *packets.S2CGameStatePacket) {
	if !serverGameActive {
		return
	}
	applyServerRound(p.Round, p.Deadline, p.Map, p.Coins, p.Points, p.Resources)
}

func ApplyServerGameEnd(p *packets.S2CGameEndPacket) {
	if !serverGameActive {
		return
	}
	serverGameActive = false
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
	if currentGame == nil || !currentGame.Multiplayer {
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
	applyGameState(state)
	SetActiveScreen(GameScreenID)
}

func RejectGameStart(message string) {
	gameActionError = message
}

func RejectGameJoin(message string) {
	clearMatchmaking()
	SetPlayError(message)
	SetActiveScreen(PlayScreenID)
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
	clearCurrentGame()
	SetActiveScreen(PlayScreenID)
}

func ResetGameSession() {
	wasMatchmaking := matchmakingActive
	clearMatchmaking()
	if currentGame == nil || !currentGame.Multiplayer {
		if wasMatchmaking {
			SetActiveScreen(PlayScreenID)
		}
		return
	}
	clearCurrentGame()
	SetActiveScreen(PlayScreenID)
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
	if currentGame.Multiplayer {
		if err := gameNet.Send(&packets.C2SLeaveGamePacket{}); err != nil {
			println("failed to leave game:", err.Error())
		}
	}
	clearCurrentGame()
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
	serverGameActive = false
	gameOverMessage = ""
	gameActionError = ""
	serverResources = nil
}

func setBuildingClick(building game.BuildingType) func() {
	return func() {
		gameWorld.Renderer.BuildingToPlace = building
	}
}

var GameScreen = ui.Screen().
	WithEnter(func() {
		EscScreen.WithVisible(false)

		audio.StartMusic()
		audio.StartAmbience()
	}).
	WithExit(func() {
		audio.StopMusic()
		audio.StopAmbience()
	}).
	AddChild(
		gameWorld,
	).
	AddChild(
		ui.Text().
			WithTextDynamic(func() string {
				if currentGame == nil || !currentGame.Multiplayer {
					return ""
				}
				return "Game code: " + currentGame.GameCode
			}).
			WithTextSize(28).
			WithTextColor(rl.Black).
			WithAnchors(anchor.TopRight, anchor.TopRight).
			WithRelativePos(vec.Vec2i{X: -20, Y: 20}),
	).
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
					gameActionError = err.Error()
					fmt.Printf("failed to request game start: %v\n", err)
				}
			}),
	).
	AddChild(
		ui.Text().
			WithTextDynamic(func() string { return gameActionError }).
			WithTextSize(22).
			WithTextColor(rl.Maroon).
			WithAnchors(anchor.Top, anchor.Top).
			WithRelativePos(vec.Vec2i{X: 0, Y: 110}).
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
			// Local map tools would desync a server-run game.
			WithVisibleDynamic(func(el *ui.GroupElement) bool {
				return !serverGameActive
			}).
			AddChild(
				gameSeedInput.
					WithPadding(8).
					WithTextSize(24).
					WithSize(vec.Vec2i{X: 320, Y: 0}).
					WithPlaceholderText("Seed"),
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
		EscScreen).
	AddChild(
		ui.Vignette().WithAlpha(120),
	)
