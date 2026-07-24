package screens

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"

	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var PlayScreen = ui.Screen().
	AddChild(
		ui.Text().
			WithText("Play").
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
	AddChild(
		ui.Button().
			WithText("Singleplayer").
			WithAnchors(anchor.Bottom, anchor.Center).
			WithRelativePos(vec.Vec2i{X: 0, Y: -16}).
			WithClick(func() {
				SetActiveScreen(GameScreenID)
			}),
	).
	AddChild(
		ui.Input().
			WithPlaceholderText("Placeholder").
			WithSize(vec.Vec2i{X: 600, Y: 0}).
			WithAnchors(anchor.Top, anchor.Center),
	).
	AddChild(
		ui.Input().
			WithPlaceholderText("Placeholder").
			WithText("Text").
			WithSize(vec.Vec2i{X: 600, Y: 0}).
			WithAnchors(anchor.Top, anchor.Center).
			WithRelativePos(vec.Vec2i{X: 0, Y: 80}),
	).
	AddChild(
		ui.Button().
			WithText("Back").
			WithTextSize(48).
			WithPadding(8).
			WithOutlineWidth(4).
			WithForegroundColors(util.ColorSet{
				Default: &rl.DarkGray,
			}).
			WithBackgroundColors(util.ColorSet{
				Default: &rl.LightGray,
				Hover:   util.ColorAdd(rl.LightGray, 25),
				Click:   util.ColorAdd(rl.LightGray, 40),
			}).
			WithOutlineColors(util.ColorSet{
				Default: &rl.Gray,
			}).
			WithAnchors(anchor.BottomLeft, anchor.BottomLeft).
			WithRelativePos(vec.Vec2i{X: 20, Y: -20}).
			WithClick(func() {
				SetActiveScreen(MainScreenID)
			}),
	)
