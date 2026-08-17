package screens

import (
	"image/color"

	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/uiutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type gameControlHint struct {
	control string
	action  string
}

var gameControlHintColumns = [][]gameControlHint{
	{
		{control: "L-CLICK", action: "Select / command / place"},
		{control: "DRAG L/R", action: "Pan map"},
		{control: "R-CLICK", action: "Cancel queued action"},
		{control: "M-CLICK", action: "Clear selection / action"},
		{control: "WHEEL", action: "Zoom"},
	},
	{
		{control: "WASD", action: "Move camera"},
		{control: "SPACE", action: "Focus Town Hall"},
		{control: "ESC", action: "Pause / back"},
		{control: "F11", action: "Toggle fullscreen"},
		{control: "F3", action: "Debug overlay"},
	},
}

func NewControlsScreen(previousScreen *ui.ScreenElement) *ui.ScreenElement {
	goBack := func() {
		GoToPreviousScreen(previousScreen)
	}

	screen := uiutil.MenuScreen().
		WithBack(goBack).
		AddChild(
			ui.Text().
				WithText("Controls").
				WithTextSize(80).
				WithTextColor(uiutil.MenuHeaderColor).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{Y: max(int32(44), el.Parent.Size().Y/8)}
				}),
		).
		AddChild(newControlsScreenGrid()).
		AddChild(uiutil.BackButton(goBack))

	screen.AddChild(uiutil.MenuVignette())
	return screen
}

func newControlsScreenGrid() *ui.GroupElement {
	const (
		controlTextSize int32 = 20
		actionTextSize  int32 = 20
		controlGap      int32 = 30
		columnGap       int32 = 72
		rowStride       int32 = 46
	)

	leftWidth := controlsColumnWidth(
		gameControlHintColumns[0],
		controlTextSize,
		actionTextSize,
		controlGap,
		true,
	)
	rightWidth := controlsColumnWidth(
		gameControlHintColumns[1],
		controlTextSize,
		actionTextSize,
		controlGap,
		true,
	)
	grid := ui.Group().
		WithSize(vec.Vec2i{
			X: leftWidth + columnGap + rightWidth,
			Y: int32(len(gameControlHintColumns[0])) * rowStride,
		}).
		WithAnchors(anchor.Center, anchor.Center).
		WithRelativePos(vec.Vec2i{Y: 28})

	shadow := color.RGBA{R: 0, G: 0, B: 0, A: 210}
	columnX := []int32{0, leftWidth + columnGap}
	for columnIndex, hints := range gameControlHintColumns {
		for rowIndex, hint := range hints {
			visible := func(*ui.TextElement) bool {
				return hint.control != "F3" || global.DebugAvailable
			}
			x := columnX[columnIndex]
			y := int32(rowIndex) * rowStride

			grid.
				AddChild(
					ui.Text().
						WithText(hint.control).
						WithTextSize(controlTextSize).
						WithTextColor(uiutil.MenuMutedColor).
						WithTextShadow(shadow, vec.Vec2i{X: 2, Y: 2}).
						WithRelativePos(vec.Vec2i{X: x, Y: y}).
						WithVisibleDynamic(visible),
				).
				AddChild(
					ui.Text().
						WithText(hint.action).
						WithTextSize(actionTextSize).
						WithTextColor(uiutil.MenuHeaderColor).
						WithTextShadow(shadow, vec.Vec2i{X: 2, Y: 2}).
						WithRelativePos(vec.Vec2i{
							X: x + ui.MeasureText(hint.control, controlTextSize).X + controlGap,
							Y: y,
						}).
						WithVisibleDynamic(visible),
				)
		}
	}

	return grid
}

func controlsColumnWidth(
	hints []gameControlHint,
	controlTextSize int32,
	actionTextSize int32,
	gap int32,
	inline bool,
) int32 {
	width := int32(0)
	for _, hint := range hints {
		if inline {
			width = max(
				width,
				ui.MeasureText(hint.control, controlTextSize).X+
					gap+
					ui.MeasureText(hint.action, actionTextSize).X,
			)
			continue
		}
		width = max(
			width,
			max(
				ui.MeasureText(hint.control, controlTextSize).X,
				ui.MeasureText(hint.action, actionTextSize).X,
			),
		)
	}
	return width
}
