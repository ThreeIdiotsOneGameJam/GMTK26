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
var escShowingCountdownSettings bool

func HideEscScreen() {
	escShowingSettings = false
	escShowingCountdownSettings = false
	if escScreen != nil {
		escScreen.WithVisible(false)
	}
	if escBlur != nil {
		escBlur.Release()
	}
}

func NewEscScreen(previousScreen *ui.ScreenElement) *ui.ScreenElement {
	escShowingSettings = false
	escShowingCountdownSettings = false

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
			escShowingCountdownSettings = false
		})).
		AddChild(uiutil.MenuButton("Leave Game", menuStartY+btnStride*2, func() {
			LeaveCurrentGame()
			GoToPreviousScreen(previousScreen)
		}))

	textShadow := color.RGBA{R: 0, G: 0, B: 0, A: 210}

	settingsPanel := ui.Screen().
		WithBackgroundColor(color.RGBA{}).
		WithVisibleDynamic(func(_ *ui.ScreenElement) bool {
			return escShowingSettings && !escShowingCountdownSettings
		}).
		AddChild(
			ui.Text().
				WithText("Settings").
				WithTextSize(64).
				WithTextColor(uiutil.MenuHeaderColor).
				WithTextShadow(textShadow, vec.Vec2i{X: 3, Y: 3}).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{Y: -230}),
		)

	const (
		sliderWidth = 160
		rowStartY   = -130
		rowStrideY  = 60
	)

	saveSettings := func() {
		if err := settings.Save(); err != nil {
			fmt.Printf("failed to save settings: %v\n", err)
		}
	}
	countdownPreview, preview := newCountdownSettingsPreview(func() bool {
		return escShowingSettings && escShowingCountdownSettings
	})

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
	addToggleRowStyled(settingsPanel, "Reduced Motion", rowStartY+rowStrideY*3, uiutil.MenuHeaderColor, uiutil.MenuMutedColor, &textShadow,
		func() bool { return settings.Current.ReducedMotion },
		func(v bool) { settings.Current.ReducedMotion = v },
		saveSettings,
	)
	addButtonRowStyled(
		settingsPanel,
		"Countdown Overlay",
		"Configure",
		rowStartY+rowStrideY*4,
		uiutil.MenuHeaderColor,
		&textShadow,
		func() {
			escShowingCountdownSettings = true
		},
	)
	settingsPanel.AddChild(
		uiutil.MenuButton("Back", 245, func() {
			escShowingSettings = false
		}),
	)

	countdownPanel := ui.Screen().
		WithBackgroundColor(color.RGBA{}).
		WithVisibleDynamic(func(_ *ui.ScreenElement) bool {
			return escShowingSettings && escShowingCountdownSettings
		}).
		AddChild(
			ui.Text().
				WithText("Countdown Overlay").
				WithTextSize(64).
				WithTextColor(uiutil.MenuHeaderColor).
				WithTextShadow(textShadow, vec.Vec2i{X: 3, Y: 3}).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{Y: -230}),
		)

	addSliderRowStyled(countdownPanel, "Size", -100, sliderWidth,
		uiutil.MenuHeaderColor, uiutil.MenuMutedColor, &textShadow,
		func() float32 { return settings.Current.CountdownScale },
		func(v float32) {
			settings.Current.CountdownScale = v
			preview.Play()
		},
		func() {
			saveSettings()
			preview.Restart()
		},
		settings.MinCountdownScale,
		settings.MaxCountdownScale,
		settings.DefaultCountdownScale,
		func(v float32) string {
			return fmt.Sprintf("%d%%", int(v*100+0.5))
		},
	)
	addCountdownAnchorRowStyled(countdownPanel, "Position", 50,
		uiutil.MenuHeaderColor, &textShadow,
		func() settings.CountdownAnchor { return settings.Current.CountdownAnchor },
		func(v settings.CountdownAnchor) {
			settings.Current.CountdownAnchor = v
			preview.Restart()
		},
		saveSettings,
	)
	countdownPanel.AddChild(
		uiutil.MenuButton("Back", 245, func() {
			escShowingCountdownSettings = false
		}),
	)

	return ui.Screen().
		WithBackgroundColor(color.RGBA{}).
		WithVisible(false).
		AddChild(escBlur).
		AddChild(menuPanel).
		AddChild(settingsPanel).
		AddChild(countdownPanel).
		AddChild(countdownPreview)
}
