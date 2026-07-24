package screens

import (
	"image/color"

	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
)

var EscScreen = ui.Screen().
	WithBackgroundColor(color.RGBA{R: 100, G: 100, B: 100, A: 100}).
	WithVisible(false).
	AddChild(
		ui.Button().WithText("TestESC").WithAnchors(anchor.Center, anchor.Center).WithClick(func() { println("TEST ESC") }))
