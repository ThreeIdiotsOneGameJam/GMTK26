package screens

import (
	"fmt"
	"image/color"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/uiutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func NewSettingsScreen(previousScreen *ui.ScreenElement) *ui.ScreenElement {
	const (
		sliderWidth = 180
		rowStartY   = -105
		rowStrideY  = 60
	)

	showCountdownSettings := false
	saveSettings := func() {
		if err := settings.Save(); err != nil {
			fmt.Printf("failed to save settings: %v\n", err)
		}
	}
	goBack := func() {
		if showCountdownSettings {
			showCountdownSettings = false
			return
		}
		GoToPreviousScreen(previousScreen)
	}
	countdownPreview, preview := newCountdownSettingsPreview(func() bool {
		return showCountdownSettings
	})

	screen := uiutil.MenuScreen().
		WithBack(goBack)

	settingsPanel := ui.Screen().
		WithBackgroundColor(color.RGBA{}).
		WithVisibleDynamic(func(_ *ui.ScreenElement) bool {
			return !showCountdownSettings
		}).
		AddChild(
			ui.Text().
				WithText("Settings").
				WithTextSize(80).
				WithTextColor(uiutil.MenuHeaderColor).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{X: 0, Y: el.Parent.Size().Y / 8}
				}),
		)

	addVolumeRow(settingsPanel, "Music", rowStartY, sliderWidth,
		func() float32 { return settings.Current.MusicVolume },
		func(v float32) { settings.Current.MusicVolume = v },
		saveSettings,
	)
	addVolumeRow(settingsPanel, "SFX", rowStartY+rowStrideY, sliderWidth,
		func() float32 { return settings.Current.SFXVolume },
		func(v float32) { settings.Current.SFXVolume = v },
		saveSettings,
	)
	addVolumeRow(settingsPanel, "Ambience", rowStartY+rowStrideY*2, sliderWidth,
		func() float32 { return settings.Current.AmbienceVolume },
		func(v float32) { settings.Current.AmbienceVolume = v },
		saveSettings,
	)
	addToggleRow(settingsPanel, "Reduced Motion", rowStartY+rowStrideY*3,
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
		nil,
		func() {
			showCountdownSettings = true
		},
	)
	settingsPanel.AddChild(
		uiutil.BackButton(goBack).
			WithTextSize(48).
			WithPadding(8),
	)

	countdownPanel := ui.Screen().
		WithBackgroundColor(color.RGBA{}).
		WithVisibleDynamic(func(_ *ui.ScreenElement) bool {
			return showCountdownSettings
		}).
		AddChild(
			ui.Text().
				WithText("Countdown Overlay").
				WithTextSize(72).
				WithTextColor(uiutil.MenuHeaderColor).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{Y: el.Parent.Size().Y / 8}
				}),
		)

	addSliderRowStyled(countdownPanel, "Size", -90, sliderWidth,
		uiutil.MenuHeaderColor, uiutil.MenuMutedColor, nil,
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
	addCountdownAnchorRowStyled(countdownPanel, "Position", 60,
		uiutil.MenuHeaderColor, nil,
		func() settings.CountdownAnchor { return settings.Current.CountdownAnchor },
		func(v settings.CountdownAnchor) {
			settings.Current.CountdownAnchor = v
			preview.Restart()
		},
		saveSettings,
	)
	countdownPanel.AddChild(
		uiutil.BackButton(func() {
			showCountdownSettings = false
		}).
			WithTextSize(48).
			WithPadding(8),
	)

	return screen.
		AddChild(settingsPanel).
		AddChild(countdownPanel).
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
				WithAnchors(anchor.Bottom, anchor.Bottom).
				WithRelativePos(vec.Vec2i{X: 0, Y: -20}).
				WithVisibleDynamic(func(*ui.TextElement) bool {
					return global.DebugEnabled
				}),
		).
		AddChild(uiutil.MenuVignette()).
		AddChild(
			countdownPreview,
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
	addVolumeRowStyled(
		screen,
		label,
		centerY,
		sliderWidth,
		uiutil.MenuHeaderColor,
		uiutil.MenuMutedColor,
		nil,
		get,
		set,
		commit,
	)
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
	addSliderRowStyled(
		screen,
		label,
		centerY,
		sliderWidth,
		labelColor,
		valueColor,
		shadow,
		get,
		set,
		commit,
		0,
		1,
		1,
		func(v float32) string {
			return fmt.Sprintf("%d%%", int(v*100+0.5))
		},
	)
}

func addSliderRowStyled(
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
	minValue float32,
	maxValue float32,
	defaultValue float32,
	formatValue func(float32) string,
) {
	const splitGap int32 = 24

	labelText := ui.Text().
		WithText(label).
		WithTextSize(32).
		WithTextColor(labelColor).
		WithAnchors(anchor.Right, anchor.Center).
		WithRelativePos(vec.Vec2i{X: -splitGap, Y: centerY})
	valueText := ui.Text().
		WithTextDynamic(func() string {
			return formatValue(get())
		}).
		WithTextSize(28).
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
				WithRange(minValue, maxValue).
				WithValueDynamic(get).
				WithDefaultValue(defaultValue).
				WithCallback(set).
				WithCommit(func(float32) { commit() }).
				WithAnchors(anchor.Left, anchor.Center).
				WithRelativePos(vec.Vec2i{X: splitGap, Y: centerY}),
		).
		AddChild(valueText)
}

func addToggleRow(
	screen *ui.ScreenElement,
	label string,
	centerY int32,
	get func() bool,
	set func(bool),
	commit func(),
) {
	addToggleRowStyled(
		screen,
		label,
		centerY,
		uiutil.MenuHeaderColor,
		uiutil.MenuMutedColor,
		nil,
		get,
		set,
		commit,
	)
}

func addToggleRowStyled(
	screen *ui.ScreenElement,
	label string,
	centerY int32,
	labelColor color.RGBA,
	valueColor color.RGBA,
	shadow *color.RGBA,
	get func() bool,
	set func(bool),
	commit func(),
) {
	const (
		splitGap    int32 = 24
		toggleWidth int32 = 64
	)

	labelText := ui.Text().
		WithText(label).
		WithTextSize(32).
		WithTextColor(labelColor).
		WithAnchors(anchor.Right, anchor.Center).
		WithRelativePos(vec.Vec2i{X: -splitGap, Y: centerY})
	valueText := ui.Text().
		WithTextDynamic(func() string {
			if get() {
				return "On"
			}
			return "Off"
		}).
		WithTextSize(28).
		WithTextColor(valueColor).
		WithAnchors(anchor.Left, anchor.Center).
		WithRelativePos(vec.Vec2i{
			X: splitGap + toggleWidth + 24,
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
			ui.Toggle().
				WithSize(vec.Vec2i{X: toggleWidth, Y: 36}).
				WithValueDynamic(get).
				WithDefaultValue(false).
				WithCallback(set).
				WithCommit(func(bool) { commit() }).
				WithAnchors(anchor.Left, anchor.Center).
				WithRelativePos(vec.Vec2i{X: splitGap, Y: centerY}),
		).
		AddChild(valueText)
}

func addButtonRowStyled(
	screen *ui.ScreenElement,
	label string,
	buttonText string,
	centerY int32,
	labelColor color.RGBA,
	shadow *color.RGBA,
	click func(),
) {
	const splitGap int32 = 24

	labelText := ui.Text().
		WithText(label).
		WithTextSize(32).
		WithTextColor(labelColor).
		WithAnchors(anchor.Right, anchor.Center).
		WithRelativePos(vec.Vec2i{X: -splitGap, Y: centerY})
	button := ui.Button().
		WithText(buttonText).
		WithTextSize(24).
		WithPadding(4).
		WithOutlineWidth(2).
		WithSize(vec.Vec2i{X: 150, Y: 38}).
		WithAnchors(anchor.Left, anchor.Center).
		WithRelativePos(vec.Vec2i{X: splitGap, Y: centerY}).
		WithClick(click)
	if shadow != nil {
		offset := vec.Vec2i{X: 2, Y: 2}
		labelText.WithTextShadow(*shadow, offset)
		button.WithShadow(*shadow, offset)
	}

	screen.
		AddChild(labelText).
		AddChild(button)
}

func addCountdownAnchorRowStyled(
	screen *ui.ScreenElement,
	label string,
	centerY int32,
	labelColor color.RGBA,
	shadow *color.RGBA,
	get func() settings.CountdownAnchor,
	set func(settings.CountdownAnchor),
	commit func(),
) {
	const (
		splitGap int32 = 24
		cellSize int32 = 28
		cellGap  int32 = 8
	)

	labelText := ui.Text().
		WithText(label).
		WithTextSize(32).
		WithTextColor(labelColor).
		WithAnchors(anchor.Right, anchor.Center).
		WithRelativePos(vec.Vec2i{X: -splitGap, Y: centerY})
	if shadow != nil {
		labelText.WithTextShadow(*shadow, vec.Vec2i{X: 2, Y: 2})
	}

	type anchorChoice struct {
		value  settings.CountdownAnchor
		column int32
		row    int32
		button *ui.ButtonElement
	}
	choices := make(
		[]anchorChoice,
		0,
		settings.CountdownGridSize*settings.CountdownGridSize,
	)
	for row := int32(0); row < settings.CountdownGridSize; row++ {
		for column := int32(0); column < settings.CountdownGridSize; column++ {
			choices = append(choices, anchorChoice{
				value:  settings.CountdownAnchorAt(column, row),
				column: column,
				row:    row,
			})
		}
	}

	gridSize := cellSize*settings.CountdownGridSize +
		cellGap*(settings.CountdownGridSize-1)
	grid := ui.Group().
		WithSize(vec.Vec2i{X: gridSize, Y: gridSize}).
		WithAnchors(anchor.Left, anchor.Center).
		WithRelativePos(vec.Vec2i{X: splitGap, Y: centerY})

	for i := range choices {
		choice := &choices[i]
		value := choice.value
		choice.button = ui.Button().
			WithText("").
			WithTextSize(14).
			WithPadding(0).
			WithOutlineWidth(2).
			WithoutShadow().
			WithSize(vec.Vec2i{X: cellSize, Y: cellSize}).
			WithAnchors(anchor.TopLeft, anchor.TopLeft).
			WithRelativePos(vec.Vec2i{
				X: choice.column * (cellSize + cellGap),
				Y: choice.row * (cellSize + cellGap),
			}).
			WithClick(func() {
				set(value)
				commit()
			})
		grid.AddChild(choice.button)
	}

	refresh := func() {
		for i := range choices {
			choice := &choices[i]
			if choice.value == get() {
				choice.button.
					WithForegroundColors(ui.ColorSet{Default: &ui.PaletteText}).
					WithBackgroundColors(ui.ColorSet{
						Default: &ui.PaletteIndigo,
						Hover:   &ui.PaletteIndigoHover,
						Click:   &ui.PaletteIndigoPress,
					}).
					WithOutlineColors(ui.ColorSet{Default: &ui.PaletteViolet})
				continue
			}
			choice.button.
				WithForegroundColors(ui.ColorSet{Default: &ui.PaletteTextSecondary}).
				WithBackgroundColors(ui.ColorSet{
					Default: &ui.PaletteSurfaceUp,
					Hover:   &ui.PaletteIndigoDim,
					Click:   &ui.PaletteIndigoPress,
				}).
				WithOutlineColors(ui.ColorSet{Default: &ui.PaletteBorder})
		}
	}
	refresh()
	grid.WithUpdate(func(_ int64) {
		refresh()
	})

	screen.
		AddChild(labelText).
		AddChild(grid)
}

type countdownSettingsPreview struct {
	startedAt time.Time
}

func (preview *countdownSettingsPreview) Play() {
	_, _, visible := countdownSettingsPreviewState(preview.startedAt, time.Now())
	if !visible {
		preview.Restart()
	}
}

func (preview *countdownSettingsPreview) Restart() {
	preview.startedAt = time.Now()
}

func countdownSettingsPreviewState(
	startedAt time.Time,
	now time.Time,
) (digit int, progress float64, visible bool) {
	if startedAt.IsZero() || now.Before(startedAt) {
		return 0, 0, false
	}
	return roundCountdownState(roundCountdownDuration - now.Sub(startedAt))
}

func newCountdownSettingsPreview(
	visibleWhen func() bool,
) (*ui.GroupElement, *countdownSettingsPreview) {
	preview := &countdownSettingsPreview{}
	displayText := ""
	visible := false

	text := ui.Text().
		WithTextDynamic(func() string {
			return displayText
		}).
		WithTextSize(roundCountdownTextSize(
			0,
			int32(rl.GetRenderHeight()),
			settings.Current.CountdownScale,
		)).
		WithTextColor(color.RGBA{R: 255, G: 244, B: 194, A: 255}).
		WithTextShadow(color.RGBA{R: 14, G: 12, B: 25, A: 220}, vec.Vec2i{X: 7, Y: 9}).
		WithAnchors(anchor.Center, anchor.Center).
		WithVisibleDynamic(func(_ *ui.TextElement) bool {
			return visible
		})

	group := ui.Group().
		WithAnchors(anchor.Center, anchor.TopLeft).
		WithRelativePosDynamic(func(el *ui.GroupElement) vec.Vec2i {
			return roundCountdownPosition(
				settings.Current.CountdownAnchor,
				el.Parent.Size(),
			)
		}).
		WithUpdate(func(_ int64) {
			digit, progress, previewVisible := countdownSettingsPreviewState(
				preview.startedAt,
				time.Now(),
			)
			if !previewVisible {
				visible = false
				return
			}

			displayText = fmt.Sprintf("%d", digit)
			visible = true
			text.TextSize = roundCountdownTextSize(
				progress,
				int32(rl.GetRenderHeight()),
				settings.Current.CountdownScale,
			)
		}).
		AddChild(text)
	if visibleWhen != nil {
		group.WithVisibleDynamic(func(_ *ui.GroupElement) bool {
			return visibleWhen()
		})
	}

	return group, preview
}
