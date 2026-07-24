package screens

import (
	"hash/fnv"
	"math/rand"
	"strconv"

	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var gameSeedInput = ui.Input()
var gameWorld = ui.World()
var gameRegenerateButton = ui.Button()

var GameScreen = ui.Screen().
	WithEnter(func() { EscScreen.WithVisible(false) }).
	AddChild(
		gameWorld,
	).
	AddChild(
		ui.Group().
			WithAnchors(anchor.TopLeft, anchor.TopLeft).
			WithRelativePos(vec.Vec2i{X: 8, Y: 8}).
			AddChild(
				gameSeedInput.
					WithPadding(8).
					WithPlaceholderText("Seed"),
			).
			AddChild(
				gameRegenerateButton.
					WithPadding(8).
					WithRelativePos(vec.Vec2i{X: 0, Y: 80}).
					WithText("Regenerate").
					WithClick(func() {
						if gameSeedInput.Text == "0" || gameSeedInput.Text == "" {
							gameWorld.World.Seed = 0
						} else {	
							h := fnv.New64a()
							h.Write([]byte(gameSeedInput.Text))
							gameWorld.World.Seed = int64(h.Sum64())
						}

						gameWorld.World.Generate()
					}),
		).
		AddChild(
			ui.Button().
				WithPadding(8).
				WithRelativePos(vec.Vec2i{X: 0, Y: 160}).
				WithText("Random").
				WithClick(func() {
					gameSeedInput.Text = strconv.FormatInt(rand.Int63(), 10)
					gameRegenerateButton.Click()
				}),
		),
	).
	AddChild(
		EscScreen)
