package screens

import (
	"hash/fnv"
	"math/rand"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var gameSeedInput = ui.Input()
var gameWorld = ui.World()
var gameRegenerateButton = ui.Button()
var gameMouseBuilding = ui.Building(gameWorld).
	WithType(game.BuildingMine).
	WithPosProvider(func(el *ui.BuildingElement) vec.Vec2 {
		return el.World.Renderer.MousePosition
	})

func setBuildingClick(building game.BuildingType) func() {
	return func() {
		gameMouseBuilding.Type = building
	}
}

var GameScreen = ui.Screen().
	WithEnter(func() { EscScreen.WithVisible(false) }).
	AddChild(
		gameWorld,
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
		gameMouseBuilding,
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
						if gameSeedInput.Text == "0" || gameSeedInput.Text == "" {
							gameWorld.Map.Seed = 0
						} else {
							h := fnv.New64a()
							h.Write([]byte(gameSeedInput.Text))
							gameWorld.Map.Seed = int64(h.Sum64())
						}

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
