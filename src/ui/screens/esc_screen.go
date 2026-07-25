package screens

import (
	"image/color"

	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var escBlur *ui.BlurBackdropElement

func HideEscScreen() {
	if escScreen != nil {
		escScreen.WithVisible(false)
	}
	if escBlur != nil {
		escBlur.Release()
	}
}

func NewEscScreen(previousScreen *ui.ScreenElement) *ui.ScreenElement {
	escBlur = ui.BlurBackdrop().
		WithTint(color.RGBA{R: 100, G: 100, B: 100, A: 100}).
		WithOffset(0.1)

	resumeButton := ui.Button().
		WithText("Resume").
		WithAnchors(anchor.Center, anchor.Center).
		WithClick(HideEscScreen)

	leaveButton := ui.Button().
		WithText("Leave Game").
		WithAnchors(anchor.Center, anchor.Center).
		WithRelativePos(vec.Vec2i{Y: 80}).
		WithClick(func() {
			LeaveCurrentGame()
			HideEscScreen()
			GoToPreviousScreen(previousScreen)
		})

	return ui.Screen().
		WithBackgroundColor(color.RGBA{}).
		WithVisible(false).
		AddChild(escBlur).
		AddChild(resumeButton).
		AddChild(leaveButton)
}
