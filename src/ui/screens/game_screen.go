package screens

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
)

var GameScreen = ui.Screen().
	WithEnter(func() { EscScreen.WithVisible(false) }).
	AddChild(
		ui.World(),
	).
	AddChild(
		EscScreen)
