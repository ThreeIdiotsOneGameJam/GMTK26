package uiutil

import (
	"image/color"

	"github.com/threeidiotsonegamejam/gmtk26/src/render/shaders"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

// MenuScreenBackground matches the menu_background.frag base color for
// fade transitions between menu screens.
var MenuScreenBackground = color.RGBA{R: 49, G: 44, B: 88, A: 255}

var (
	MenuHeaderColor = ui.PaletteText
	MenuMutedColor  = ui.PaletteTextSecondary
)

func MenuBackdrop() *ui.ShaderElement {
	return ui.Shader().WithShader(&shaders.MenuBackground)
}

// MenuScreen creates the shared backdrop for full-screen menus. Add content,
// then add MenuVignette last so it stays above the menu.
func MenuScreen() *ui.ScreenElement {
	return ui.Screen().
		WithBackgroundColor(MenuScreenBackground).
		AddChild(MenuBackdrop())
}

func MenuVignette() *ui.VignetteElement {
	return ui.Vignette()
}

func MenuTitle(text string, textSize, y int32) *ui.TextElement {
	return ui.Text().
		WithText(text).
		WithTextSize(textSize).
		WithTextColor(MenuHeaderColor).
		WithAnchors(anchor.Center, anchor.Top).
		WithRelativePos(vec.Vec2i{Y: y})
}
