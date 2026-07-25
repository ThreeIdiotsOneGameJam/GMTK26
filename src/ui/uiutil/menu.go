package uiutil

import (
	"image/color"

	"github.com/threeidiotsonegamejam/gmtk26/src/render/shaders"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
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
