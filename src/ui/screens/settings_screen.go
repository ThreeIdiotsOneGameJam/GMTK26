package screens

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/mathutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/mathutil/vec"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
)

var SettingsScreen = ui.Screen().
	AddChild(
		ui.Text().
			WithText("Settings").
			WithTextSize(96).
			WithTextColor(rl.Black).
			WithAnchors(anchor.Center, anchor.Top).
			WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
				return vec.Vec2i{X: 0, Y: el.Parent.Size().Y / 6}
			}),
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
				Hover:   mathutil.ColorAdd(rl.LightGray, 25),
				Click:   mathutil.ColorAdd(rl.LightGray, 40),
			}).
			WithOutlineColors(util.ColorSet{
				Default: &rl.Gray,
			}).
			WithAnchors(anchor.BottomLeft, anchor.BottomLeft).
			WithRelativePos(vec.Vec2i{
				X: 20,
				Y: -20,
			}).
			WithClick(func() {
				SetActiveScreen(MainScreenID)
			}),
	)
