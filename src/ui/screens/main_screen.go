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
	actions := []ui.Element{
		uiutil.MenuAction("Play", func() {
			if ui.DebugQuickActionModifierHeld() {
				StartSoloWithDefaults()
				return
			}
			SetActiveScreen(NewPlayScreen(screen))
		}),
		uiutil.MenuAction("Settings", func() {
			SetActiveScreen(NewSettingsScreen(screen))
		}),
	}
	if runtime.GOOS != "js" {
		actions = append(actions, uiutil.MenuAction("Exit", global.CloseWindow))
	}

	screen = uiutil.MenuScreen().
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
		AddChild(
			ui.VStack(18, actions...).
				WithAlignment(ui.StackCenter).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{Y: int32(len(actions)-1) * 40}),
		)
	screen.AddChild(uiutil.MenuVignette())
	return screen
}
