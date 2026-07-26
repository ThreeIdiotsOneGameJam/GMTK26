package screens

import (
	"runtime"

	"github.com/threeidiotsonegamejam/gmtk26/src/constants"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/uiutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func NewMainScreen() *ui.ScreenElement {
	var screen *ui.ScreenElement
	screen = ui.Screen().
		WithBackgroundColor(uiutil.MenuScreenBackground).
		AddChild(uiutil.MenuBackdrop()).
		AddChild(
			ui.Text().
				WithText(constants.GameName).
				WithTextSize(96).
				WithTextColor(uiutil.MenuHeaderColor).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{
						X: 0,
						Y: el.Parent.Size().Y / 4,
					}
				}),
		).
		AddChild(uiutil.MenuButton("Play", 0, func() {
			if ui.DebugQuickActionModifierHeld() {
				StartSoloWithDefaults()
				return
			}
			SetActiveScreen(NewPlayScreen(screen))
		})).
		AddChild(uiutil.MenuButton("Settings", 80, func() {
			SetActiveScreen(NewSettingsScreen(screen))
		}))
	if runtime.GOOS != "js" {
		screen.AddChild(uiutil.MenuButton("Exit", 160, global.CloseWindow))
	}
	screen.AddChild(ui.Vignette())
	return screen
}
