package screens

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/math"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
)

var SettingsScreen = ui.Screen{
	BackgroundColor: rl.RayWhite,
	Elements: []ui.Element{
		&ui.TextElement{
			Pos: func(renderWidth, renderHeight int32) math.Vec2i {
				return math.Vec2i{
					X: renderWidth / 2,
					Y: renderHeight / 6,
				}
			},
			Text:      "Settings",
			TextSize:  96,
			TextColor: rl.Black,
		},
		&ui.ButtonElement{
			Pos: func(renderWidth, renderHeight int32) math.Vec2i {
				return math.Vec2i{
					X: 20 + 8 + 4/2 + rl.MeasureText("Back", 48)/2,
					Y: renderHeight - 20 - 8 - 4/2 - 48/2,
				}
			},
			Text:         "Back",
			TextSize:     48,
			Padding:      8,
			OutlineWidth: 4,
			ForegroundColors: util.ColorSet{
				Default: &rl.DarkGray,
			},
			BackgroundColors: util.ColorSet{
				Default: &rl.LightGray,
				Hover:   math.ColorAdd(rl.LightGray, 25),
				Click:   math.ColorAdd(rl.LightGray, 40),
			},
			OutlineColors: util.ColorSet{
				Default: &rl.Gray,
			},
			Click: func() {
				SetActiveScreen(MainScreenID)
			},
		},
	},
}
