package uiutil

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func MenuButton(text string, y int32, click func()) *ui.ButtonElement {
	return ui.Button().
		WithText(text).
		WithSize(vec.Vec2i{X: 340, Y: 62}).
		WithAnchors(anchor.Center, anchor.Center).
		WithRelativePos(vec.Vec2i{Y: y}).
		WithClick(click)
}
