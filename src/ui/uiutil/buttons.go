package uiutil

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var MenuButtonStyle = ui.DefaultButtonStyle

func MenuAction(text string, click func()) *ui.ButtonElement {
	return ui.Button().
		WithStyle(MenuButtonStyle).
		WithText(text).
		WithSize(vec.Vec2i{X: 340, Y: 62}).
		WithClick(click)
}

func BackButton(click func()) *ui.ButtonElement {
	return ui.Button().
		WithText("Back").
		WithTextSize(40).
		WithPadding(8).
		WithOutlineWidth(4).
		WithAnchors(anchor.BottomLeft, anchor.BottomLeft).
		WithRelativePos(vec.Vec2i{X: 20, Y: -20}).
		WithClick(click)
}
