package screens

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/render/shaders"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/uiutil"

	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func NewMainScreen() *ui.ScreenElement {
	var screen *ui.ScreenElement
	screen = ui.Screen().
		AddChild(ui.Shader().WithShader(&shaders.MenuBackground)).
		AddChild(
			ui.Text().
				WithText("Game").
				WithTextSize(96).
				WithTextColor(rl.Black).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{
						X: 0,
						Y: el.Parent.Size().Y / 4,
					}
				}),
		).
		AddChild(uiutil.MenuButton("Play", 0, func() {
			SetActiveScreen(NewPlayScreen(screen))
		})).
		AddChild(uiutil.MenuButton("Settings", 80, func() {
			SetActiveScreen(NewSettingsScreen(screen))
		})).
		AddChild(uiutil.MenuButton("Exit", 160, global.CloseWindow)).
		AddChild(
			ui.Vignette(),
		)
	return screen
}
