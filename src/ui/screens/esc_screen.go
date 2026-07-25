package screens

import (
	"image/color"

	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func HideEscScreen() {
	EscScreen.WithVisible(false)
	escBlur.Release()
}

var (
	escBlur = ui.BlurBackdrop().
		WithTint(color.RGBA{R: 100, G: 100, B: 100, A: 100}).
		WithOffset(0.1)
	escResumeButton = ui.Button().
			WithText("Resume").
			WithAnchors(anchor.Center, anchor.Center)
	escMainMenuButton = ui.Button().
				WithText("Leave Game").
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{Y: 80})
	EscScreen = ui.Screen().
			WithBackgroundColor(color.RGBA{}).
			WithVisible(false).
			AddChild(escBlur).
			AddChild(escResumeButton).
			AddChild(escMainMenuButton)
)

func init() {
	escResumeButton.WithClick(HideEscScreen)
	escMainMenuButton.WithClick(func() {
		LeaveCurrentGame()
		HideEscScreen()
		SetActiveScreen(MainScreenID)
	})
}
