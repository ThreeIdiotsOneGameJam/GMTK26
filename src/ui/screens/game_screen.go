package screens

import (
	"math/rand"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
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

func EnterGame(state game.Game) {
	applyGameState(state)
	SetActiveScreen(GameScreenID)
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
	if currentGame == nil || !currentGame.Multiplayer {
		return
	}
	clearCurrentGame()
	SetActiveScreen(PlayScreenID)
}

func LeaveCurrentGame() {
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
}

func setBuildingClick(building game.BuildingType) func() {
	return func() {
		gameWorld.Renderer.BuildingToPlace = building
	}
}

var GameScreen = ui.Screen().
	WithEnter(func() { EscScreen.WithVisible(false) }).
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
		EscScreen).
	AddChild(
		ui.Vignette().WithAlpha(120),
	)
