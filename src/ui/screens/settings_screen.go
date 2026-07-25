package screens

import (
	"fmt"
	"image/color"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/uiutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func NewSettingsScreen(previousScreen *ui.ScreenElement) *ui.ScreenElement {
	const (
		sliderWidth = 180
		rowStartY   = -40
		rowStrideY  = 72
	)

	saveSettings := func() {
		if err := settings.Save(); err != nil {
			fmt.Printf("failed to save settings: %v\n", err)
		}
	}

	goBack := func() {
		GoToPreviousScreen(previousScreen)
	}

	screen := ui.Screen().
		WithBackgroundColor(uiutil.MenuScreenBackground).
		WithBack(goBack).
		AddChild(uiutil.MenuBackdrop()).
		AddChild(
			ui.Text().
				WithText("Settings").
				WithTextSize(96).
				WithTextColor(uiutil.MenuHeaderColor).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{X: 0, Y: el.Parent.Size().Y / 8}
				}),
		)

	addVolumeRow(screen, "Music", rowStartY, sliderWidth,
		func() float32 { return settings.Current.MusicVolume },
		func(v float32) { settings.Current.MusicVolume = v },
		saveSettings,
	)
	addVolumeRow(screen, "SFX", rowStartY+rowStrideY, sliderWidth,
		func() float32 { return settings.Current.SFXVolume },
		func(v float32) { settings.Current.SFXVolume = v },
		saveSettings,
	)
	addVolumeRow(screen, "Ambience", rowStartY+rowStrideY*2, sliderWidth,
		func() float32 { return settings.Current.AmbienceVolume },
		func(v float32) { settings.Current.AmbienceVolume = v },
		saveSettings,
	)

	return screen.
		AddChild(
			ui.Text().
				WithTextDynamic(func() string {
					p := game.PlayerData
					return fmt.Sprintf(
						"Player Data:\n  ClientID: %s\n  PlayerName: %s\n  Color: %d,%d,%d",
						p.ClientID, p.PlayerName, p.Color[0], p.Color[1], p.Color[2],
					)
				}).
				WithTextSize(24).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{X: 0, Y: 220}),
		).
		AddChild(
			ui.Button().
				WithText("Back").
				WithTextSize(48).
				WithPadding(8).
				WithOutlineWidth(4).
				WithAnchors(anchor.BottomLeft, anchor.BottomLeft).
				WithRelativePos(vec.Vec2i{
					X: 20,
					Y: -20,
				}).
				WithClick(goBack),
		).
		AddChild(
			ui.Vignette(),
		)
}

func addVolumeRow(
	screen *ui.ScreenElement,
	label string,
	centerY int32,
	sliderWidth int32,
	get func() float32,
	set func(float32),
	commit func(),
) {
	addVolumeRowStyled(screen, label, centerY, sliderWidth, uiutil.MenuHeaderColor, uiutil.MenuMutedColor, nil, get, set, commit)
}

func addVolumeRowStyled(
	screen *ui.ScreenElement,
	label string,
	centerY int32,
	sliderWidth int32,
	labelColor color.RGBA,
	valueColor color.RGBA,
	shadow *color.RGBA,
	get func() float32,
	set func(float32),
	commit func(),
) {
	const splitGap int32 = 24

	labelText := ui.Text().
		WithText(label).
		WithTextSize(36).
		WithTextColor(labelColor).
		WithAnchors(anchor.Right, anchor.Center).
		WithRelativePos(vec.Vec2i{
			X: -splitGap,
			Y: centerY,
		})
	valueText := ui.Text().
		WithTextDynamic(func() string {
			return fmt.Sprintf("%d%%", int(get()*100+0.5))
		}).
		WithTextSize(32).
		WithTextColor(valueColor).
		WithAnchors(anchor.Left, anchor.Center).
		WithRelativePos(vec.Vec2i{
			X: splitGap + sliderWidth + 24,
			Y: centerY,
		})
	if shadow != nil {
		offset := vec.Vec2i{X: 2, Y: 2}
		labelText.WithTextShadow(*shadow, offset)
		valueText.WithTextShadow(*shadow, offset)
	}

	screen.
		AddChild(labelText).
		AddChild(
			ui.Slider().
				WithSize(vec.Vec2i{X: sliderWidth, Y: 36}).
				WithValueDynamic(get).
				WithDefaultValue(1).
				WithCallback(set).
				WithCommit(func(float32) { commit() }).
				WithAnchors(anchor.Left, anchor.Center).
				WithRelativePos(vec.Vec2i{
					X: splitGap,
					Y: centerY,
				}),
		).
		AddChild(valueText)
}
