package screens

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/mathutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"

	"github.com/threeidiotsonegamejam/gmtk26/src/mathutil/vec"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
)

var MainScreen = ui.Screen().
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
	AddChild(
		ui.Button().
			WithText("Play").
			WithTextSize(48).
			WithPadding(8).
			WithOutlineWidth(4).
			WithForegroundColors(util.ColorSet{
				Default: &rl.DarkGray,
			}).
			WithBackgroundColors(util.ColorSet{
				Default: &rl.LightGray,
				Hover:   mathutil.ColorAdd(rl.LightGray, 25),
				Click:   mathutil.ColorAdd(rl.LightGray, 40),
			}).
			WithOutlineColors(util.ColorSet{
				Default: &rl.Gray,
			}).
			WithAnchors(anchor.Center, anchor.Center).
			WithRelativePos(vec.Vec2i{}).
			WithClick(func() {
				SetActiveScreen(GameScreenID)
			}),
	).
	AddChild(
		ui.Button().
			WithText("Settings").
			WithTextSize(48).
			WithPadding(8).
			WithOutlineWidth(4).
			WithForegroundColors(util.ColorSet{
				Default: &rl.DarkGray,
			}).
			WithBackgroundColors(util.ColorSet{
				Default: &rl.LightGray,
				Hover:   mathutil.ColorAdd(rl.LightGray, 25),
				Click:   mathutil.ColorAdd(rl.LightGray, 40),
			}).
			WithOutlineColors(util.ColorSet{
				Default: &rl.Gray,
			}).
			WithAnchors(anchor.Center, anchor.Center).
			WithRelativePos(vec.Vec2i{
				X: 0,
				Y: 80,
			}).
			WithClick(func() {
				SetActiveScreen(SettingsScreenID)
			}),
	)
