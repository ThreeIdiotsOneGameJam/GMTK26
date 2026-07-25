package screens

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/uiutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var escBlur *ui.BlurBackdropElement
var escShowingSettings bool

func HideEscScreen() {
	escShowingSettings = false
	if escScreen != nil {
		escScreen.WithVisible(false)
	}
	if escBlur != nil {
		escBlur.Release()
	}
}

func NewEscScreen(previousScreen *ui.ScreenElement) *ui.ScreenElement {
	escShowingSettings = false

	escBlur = ui.BlurBackdrop().
		WithTint(color.RGBA{R: 100, G: 100, B: 100, A: 100}).
		WithOffset(0.1)

	const (
		btnStride  = 80
		menuStartY = -80
	)

	menuPanel := ui.Group().
		WithSizeDynamic(func(_ *ui.GroupElement) vec.Vec2i {
			return vec.Vec2i{X: int32(rl.GetRenderWidth()), Y: int32(rl.GetRenderHeight())}
		}).
		WithVisibleDynamic(func(_ *ui.GroupElement) bool {
			return !escShowingSettings
		}).
		AddChild(uiutil.MenuButton("Resume", menuStartY, HideEscScreen)).
		AddChild(uiutil.MenuButton("Settings", menuStartY+btnStride, func() {
			escShowingSettings = true
		})).
		AddChild(uiutil.MenuButton("Leave Game", menuStartY+btnStride*2, func() {
			LeaveCurrentGame()
			HideEscScreen()
			GoToPreviousScreen(previousScreen)
		}))

	textShadow := color.RGBA{R: 0, G: 0, B: 0, A: 210}

	settingsPanel := ui.Screen().
		WithBackgroundColor(color.RGBA{}).
		WithVisibleDynamic(func(_ *ui.ScreenElement) bool {
			return escShowingSettings
		}).
		AddChild(
			ui.Text().
				WithText("Settings").
				WithTextSize(64).
				WithTextColor(uiutil.MenuHeaderColor).
				WithTextShadow(textShadow, vec.Vec2i{X: 3, Y: 3}).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{Y: -180}),
		)

	const (
		sliderWidth = 160
		rowStartY   = -40
		rowStrideY  = 68
	)

	saveSettings := func() {
		if err := settings.Save(); err != nil {
			fmt.Printf("failed to save settings: %v\n", err)
		}
	}

	addVolumeRowStyled(settingsPanel, "Music", rowStartY, sliderWidth, uiutil.MenuHeaderColor, uiutil.MenuMutedColor, &textShadow,
		func() float32 { return settings.Current.MusicVolume },
		func(v float32) { settings.Current.MusicVolume = v },
		saveSettings,
	)
	addVolumeRowStyled(settingsPanel, "SFX", rowStartY+rowStrideY, sliderWidth, uiutil.MenuHeaderColor, uiutil.MenuMutedColor, &textShadow,
		func() float32 { return settings.Current.SFXVolume },
		func(v float32) { settings.Current.SFXVolume = v },
		saveSettings,
	)
	addVolumeRowStyled(settingsPanel, "Ambience", rowStartY+rowStrideY*2, sliderWidth, uiutil.MenuHeaderColor, uiutil.MenuMutedColor, &textShadow,
		func() float32 { return settings.Current.AmbienceVolume },
		func(v float32) { settings.Current.AmbienceVolume = v },
		saveSettings,
	)

	settingsPanel.AddChild(
		uiutil.MenuButton("Back", 200, func() {
			escShowingSettings = false
		}),
	)

	return ui.Screen().
		WithBackgroundColor(color.RGBA{}).
		WithVisible(false).
		AddChild(escBlur).
		AddChild(menuPanel).
		AddChild(settingsPanel)
}
