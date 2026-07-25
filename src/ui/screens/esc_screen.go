package screens

import (
	"image/color"

	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var EscScreen = ui.Screen().
	WithBackgroundColor(color.RGBA{R: 100, G: 100, B: 100, A: 100}).
	WithVisible(false).
	AddChild(
		ui.Button().WithText("TestESC").WithAnchors(anchor.Center, anchor.Center).WithClick(func() { println("TEST ESC") })).
	AddChild(
		ui.Button().WithText("Main Menu").WithAnchors(anchor.Center, anchor.Center).WithRelativePos(vec.Vec2i{Y: 80}).WithClick(func() {
			LeaveCurrentGame()
			SetActiveScreen(MainScreenID)
		}))
