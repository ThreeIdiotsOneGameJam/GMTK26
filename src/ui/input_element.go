package ui

import (
	"math"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

// Input currently supports keyboard selection, clipboard editing, Unicode-aware
// character and word navigation, and focus traversal. Mouse-based selection and
// fixed-width clipping remain future enhancements.

// Pos and Size do not account for the outline or shadow, which are rendered
// outside the element's layout box.

func Input() *InputElement {
	el := &InputElement{
		PlaceholderText: "Input",
		MaxTextLength:   math.MaxInt,
		TextSize:        48,
		Padding:         8,
		OutlineWidth:    4,
		ForegroundColors: ColorSet{
			Default:  ptrColor(PaletteText),
			Disabled: ptrColor(PaletteTextMuted),
		},
		PlaceholderColors: ColorSet{
			Default:  ptrColor(PaletteTextMuted),
			Disabled: ptrColor(PaletteBorder),
		},
		BackgroundColors: ColorSet{
			Default:  ptrColor(PaletteSurface),
			Click:    ptrColor(PaletteSurfaceUp),
			Disabled: ptrColor(PaletteBase),
		},
		OutlineColors: ColorSet{
			Default:  ptrColor(PaletteBorder),
			Disabled: ptrColor(PaletteBase),
		},
		Callback: func(text string) {},
	}
	el.DropShadowElement = NewDropShadowElement(el)

	return el.WithSizeDynamic(func(el *InputElement) vec.Vec2i {
		displayWidth := max(
			rl.MeasureText(el.Value(), el.TextSize),
			rl.MeasureText(el.PlaceholderText, el.TextSize),
		)

		return vec.Vec2i{
			X: displayWidth + el.Padding*2,
			Y: el.TextSize + el.Padding*2,
		}
	})
}

func (el *InputElement) WithText(text string) *InputElement {
	el.setValidatedText(text)
	return el
}

func (el *InputElement) SetText(text string) {
	el.WithText(text)
}

func (el *InputElement) Value() string {
	return el.editor.string()
}

func (el *InputElement) WithPlaceholderText(placeholderText string) *InputElement {
	el.PlaceholderText = placeholderText
	return el
}

// WithMaxTextLength sets the maximum amount of runes that the Text should be able to hold
func (el *InputElement) WithMaxTextLength(maxTextLength int) *InputElement {
	el.MaxTextLength = max(0, maxTextLength)
	el.revalidateText()
	return el
}

func (el *InputElement) WithCharset(charset string) *InputElement {
	el.Charset = util.NewRuneSetFromString(charset)
	el.revalidateText()
	return el
}

func (el *InputElement) WithInputTransformer(transformer InputTransformer) *InputElement {
	el.InputTransformer = transformer
	el.revalidateText()
	return el
}

func (el *InputElement) WithTextSize(textSize int32) *InputElement {
	el.TextSize = textSize
	return el
}

func (el *InputElement) WithPadding(padding int32) *InputElement {
	el.Padding = padding
	return el
}

func (el *InputElement) WithOutlineWidth(outlineWidth int32) *InputElement {
	el.OutlineWidth = outlineWidth
	return el
}

func (el *InputElement) WithForegroundColors(foregroundColors ColorSet) *InputElement {
	el.ForegroundColors = foregroundColors
	return el
}

func (el *InputElement) WithPlaceholderColors(placeholderColors ColorSet) *InputElement {
	el.PlaceholderColors = placeholderColors
	return el
}

func (el *InputElement) WithBackgroundColors(backgroundColors ColorSet) *InputElement {
	el.BackgroundColors = backgroundColors
	return el
}

func (el *InputElement) WithOutlineColors(outlineColors ColorSet) *InputElement {
	el.OutlineColors = outlineColors
	return el
}

func (el *InputElement) WithCallback(callback func(text string)) *InputElement {
	if callback == nil {
		callback = func(text string) {}
	}
	el.Callback = callback
	return el
}

func (el *InputElement) WithSubmit(submit func(text string)) *InputElement {
	el.Submit = submit
	return el
}

func (el *InputElement) WithDefaultText(text string) *InputElement {
	el.defaultText = text
	el.hasDefault = true
	return el
}

func (el *InputElement) HasDefault() bool {
	return el.hasDefault
}

func (el *InputElement) ResetToDefault() {
	if !el.hasDefault {
		return
	}

	el.applyEdit(func() {
		el.editor.selectAll()
		el.editor.replaceSelection(el.prepareInput(el.defaultText), el.MaxTextLength)
	})
	el.Focus()
}

func (el *InputElement) Focus() {
	setFocusedInput(el)
	el.editor.moveToEnd()
}

func (el *InputElement) Blur() {
	if el.Focused() {
		setFocusedInput(nil)
	}
}

type InputTransformer func(input string) string

type InputElement struct {
	DropShadowElement[*InputElement]
	PlaceholderText       string
	MaxTextLength         int // maximum rune count; complete graphemes are never split
	Charset               util.RuneSet
	InputTransformer      InputTransformer
	TextSize              int32
	Padding, OutlineWidth int32
	ForegroundColors      ColorSet
	PlaceholderColors     ColorSet
	BackgroundColors      ColorSet
	OutlineColors         ColorSet
	Callback              func(text string)
	Submit                func(text string)

	defaultText string
	hasDefault  bool

	x, y, cx, cy, w, h, textWidth int32

	hovered bool

	editor textEditorModel
}

func (el *InputElement) revalidateText() {
	el.setValidatedText(el.Value())
}

func (el *InputElement) setValidatedText(text string) {
	runes := el.prepareInput(text)
	runes = graphemeSafePrefix(runes, el.MaxTextLength)
	el.editor.syncText(string(runes))
}

func (el *InputElement) prepareInput(input string) []rune {
	if el.InputTransformer != nil {
		input = el.InputTransformer(input)
	}

	runes := []rune(input)
	filtered := make([]rune, 0, len(runes))

	for _, char := range runes {
		if el.isCharValid(char) {
			filtered = append(filtered, char)
		}
	}

	return filtered
}

func (el *InputElement) isCharValid(char rune) bool {
	if el.Charset == nil {
		return true
	}

	return el.Charset.Contains(char)
}

func (el *InputElement) insertInput(input string) {
	if input == "" {
		return
	}

	insert := el.prepareInput(input)
	if len(insert) == 0 {
		return
	}

	el.editor.replaceSelection(insert, el.MaxTextLength)
}

func (el *InputElement) applyEdit(edit func()) {
	oldText := el.Value()

	edit()
	text := el.Value()

	if text == oldText {
		return
	}

	if el.Callback != nil {
		el.Callback(text)
	}
}

var clipboardNewlineReplacer = strings.NewReplacer(
	"\r\n", " ",
	"\n", " ",
	"\r", " ",
)

func (el *InputElement) prepare() {
	el.textWidth = rl.MeasureText(el.Value(), el.TextSize)

	el.w, el.h = max(el.textWidth+el.Padding*2, el.Size().X), max(el.TextSize+el.Padding*2, el.Size().Y)

	pos := el.AbsolutePos()
	el.x, el.y, el.cx, el.cy = pos.X, pos.Y, pos.X+el.w/2, pos.Y+el.h/2
}

func (el *InputElement) update(deltaNano int64) {
	if !el.Enabled() || global.UIModalBlocksInput {
		el.hovered = false
		el.Blur()
		return
	}

	el.hovered = (elementRect{X: el.x, Y: el.y, Width: el.w, Height: el.h}).
		containsStrict(mousePosition())

	if el.hovered {
		claimPointer(rl.MouseCursorIBeam)
	}

	if rl.IsMouseButtonPressed(rl.MouseButtonRight) && el.hovered && el.hasDefault {
		el.ResetToDefault()
		return
	}

	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		if el.hovered {
			if !el.Focused() {
				el.Focus()
			}
		} else {
			el.Blur()
		}
	}

	if el.Focused() {
		global.UIBlocksWorldInput = true
		if claimKeyboardInput() {
			el.executeEditorCommand(pollEditorCommand(el.Submit != nil))
		}
	}
}

func (el *InputElement) executeEditorCommand(command editorCommand) {
	switch command.kind {
	case editorCommandSubmit:
		if el.Submit != nil {
			el.Submit(el.Value())
		}

	case editorCommandSelectAll:
		el.editor.selectAll()

	case editorCommandPaste:
		text := clipboardNewlineReplacer.Replace(rl.GetClipboardText())
		el.applyEdit(func() {
			el.insertInput(text)
		})

	case editorCommandCopy:
		if el.editor.hasSelection() {
			rl.SetClipboardText(el.editor.selectedText())
		}

	case editorCommandCut:
		if el.editor.hasSelection() {
			rl.SetClipboardText(el.editor.selectedText())
			el.applyEdit(func() {
				el.editor.replaceSelection(nil, el.MaxTextLength)
			})
		}

	case editorCommandMove:
		el.editor.move(command.direction, command.unit, command.extend)

	case editorCommandFocusNext:
		focusAdjacentInput(el, command.direction)

	case editorCommandDelete:
		el.applyEdit(func() {
			el.editor.delete(command.direction, command.unit)
		})

	case editorCommandInsert:
		el.applyEdit(func() {
			el.insertInput(command.text)
		})
	}
}

func (el *InputElement) caretOffsetAt(index int) int32 {
	if index <= 0 {
		return 0
	}

	runes := el.editor.runes()
	width := rl.MeasureText(string(runes[:index]), el.TextSize)

	if index < len(runes) {
		width += el.TextSize / 20
	}

	return width
}

func (el *InputElement) draw() {
	state := controlState(el.Enabled(), el.hovered, el.Focused())

	oCol := el.OutlineColors.Color(state)
	pCol := el.PlaceholderColors.Color(state)
	bgCol := el.BackgroundColors.Color(state)
	fgCol := el.ForegroundColors.Color(state)

	opacity := el.Opacity()
	outlineColor := util.ColorOpacity(*oCol, opacity)
	backgroundColor := util.ColorOpacity(*bgCol, opacity)
	foregroundColor := util.ColorOpacity(*fgCol, opacity)
	selectionColor := util.ColorOpacity(rl.SkyBlue, opacity)

	el.drawShadowedOutlinedRectangle(
		elementRect{X: el.x, Y: el.y, Width: el.w, Height: el.h},
		el.OutlineWidth,
		outlineColor,
		backgroundColor,
		opacity,
	)

	textY := el.cy - el.TextSize/2

	if el.Value() == "" {
		drawTextWithShadow(
			el.PlaceholderText,
			el.x+el.Padding,
			textY,
			el.TextSize,
			*pCol,
			el.ShadowColor,
			fontScaledShadowOffset(el.TextSize),
			opacity,
		)
	} else {
		if el.editor.hasSelection() {
			start, end := el.editor.selectionRange()

			startX := el.caretOffsetAt(start)
			endX := el.caretOffsetAt(end)

			rl.DrawRectangle(el.x+el.Padding+startX, textY, endX-startX, el.TextSize, selectionColor)
		}

		drawTextWithShadow(
			el.Value(),
			el.x+el.Padding,
			textY,
			el.TextSize,
			*fgCol,
			el.ShadowColor,
			fontScaledShadowOffset(el.TextSize),
			opacity,
		)
	}

	if el.Focused() && int(rl.GetTime()*2)%2 == 0 {
		cursorWidth := max(rl.MeasureText("|", el.TextSize)/2, 1)
		cursorX := el.caretOffsetAt(el.editor.caretPosition())

		rl.DrawRectangle(el.x+el.Padding+cursorX, textY-cursorWidth/2, cursorWidth, el.TextSize+cursorWidth, foregroundColor)
	}
}
