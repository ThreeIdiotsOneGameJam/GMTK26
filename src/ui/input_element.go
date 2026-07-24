package ui

import (
	"math"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

/*
TODO: this motherfucker needs so much functionality that i decided the just leave it half unimplemented
 as noone is gonna notice anyway and we dont have time but heres a list of most stuff i could think of that is missing:

- overlapping inputs will probably fuck shit up
- no mouse-based selection or cursorPos init from mouse position
- more keyboard shortcuts (ctrl/option+arrow, cmd+arrow/home/end, combinations with shift)
- double-click word select, triple-click line select
- tab focus traversal
- fixed-width horizontal scrolling + clipping (currently just expands infinitely lol)
- disabled state (readonly)

FIXME: setting charset does not reverify existing text
FIXME: setting max length does not reverify existing text
FIXME: WithText can bypass all checks

FIXME: geometry/layout are not recalculated after processing text input, leaving them one frame behind edits - needs to be calculated again at end of update
*/

// Pos and Size do not account for the outline, which is rendered outside this

func Input() *InputElement {
	el := &InputElement{
		Text:            "",
		PlaceholderText: "Input",
		MaxTextLength:   math.MaxInt, // MaxTextLength refers to the maximum amount of runes that the Text should be able to hold
		TextSize:        48,
		Padding:         8,
		OutlineWidth:    4,
		ForegroundColors: util.ColorSet{
			Default: &rl.DarkGray,
		},
		PlaceholderColors: util.ColorSet{
			Default: util.ColorSub(rl.LightGray, 10),
		},
		BackgroundColors: util.NewColorSetClick(util.SimpleGrayscaleColor(220), util.SimpleGrayscaleColor(240)),
		OutlineColors: util.ColorSet{
			Default: util.ColorAdd(rl.Gray, 40),
		},
		Callback: func(text string) {},
	}
	el.BaseElement = NewBaseElement(el)

	return el.WithSizeDynamic(func(el *InputElement) vec.Vec2i {
		displayWidth := max(
			rl.MeasureText(el.Text, el.TextSize),
			rl.MeasureText(el.PlaceholderText, el.TextSize),
		)

		return vec.Vec2i{
			X: displayWidth + el.Padding*2,
			Y: el.TextSize + el.Padding*2,
		}
	})
}

func (el *InputElement) WithText(text string) *InputElement {
	el.Text = text
	return el
}

func (el *InputElement) WithPlaceholderText(placeholderText string) *InputElement {
	el.PlaceholderText = placeholderText
	return el
}

// WithMaxTextLength sets the maximum amount of runes that the Text should be able to hold
func (el *InputElement) WithMaxTextLength(maxTextLength int) *InputElement {
	el.MaxTextLength = max(0, maxTextLength)
	return el
}

func (el *InputElement) WithCharset(charset string) *InputElement {
	el.Charset = util.NewRuneSetFromString(charset)
	return el
}

func (el *InputElement) WithInputTransformer(transformer InputTransformer) *InputElement {
	el.InputTransformer = transformer
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

func (el *InputElement) WithForegroundColors(foregroundColors util.ColorSet) *InputElement {
	el.ForegroundColors = foregroundColors
	return el
}

func (el *InputElement) WithPlaceholderColors(placeholderColors util.ColorSet) *InputElement {
	el.PlaceholderColors = placeholderColors
	return el
}

func (el *InputElement) WithBackgroundColors(backgroundColors util.ColorSet) *InputElement {
	el.BackgroundColors = backgroundColors
	return el
}

func (el *InputElement) WithOutlineColors(outlineColors util.ColorSet) *InputElement {
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

type InputTransformer func(input string) string

type InputElement struct {
	BaseElement[*InputElement]
	Text, PlaceholderText string
	MaxTextLength         int // in runes, NOT []byte len
	Charset               util.RuneSet
	InputTransformer      InputTransformer
	TextSize              int32
	Padding, OutlineWidth int32
	ForegroundColors      util.ColorSet
	PlaceholderColors     util.ColorSet
	BackgroundColors      util.ColorSet
	OutlineColors         util.ColorSet
	Callback              func(text string)

	x, y, cx, cy, w, h, textWidth int32

	hovered bool
	clicked bool

	cursorPos        int // []rune index, NOT string []byte pos
	selectionStart   int // []rune index, NOT string []byte pos
	selectionStarted bool

	runes                                []rune
	textBefore, textSelection, textAfter []rune
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

func (el *InputElement) recalculateTextSplit() {
	el.runes = []rune(el.Text)

	el.cursorPos = util.Clamp(el.cursorPos, 0, len(el.runes))
	el.selectionStart = util.Clamp(el.selectionStart, 0, len(el.runes))

	start, end := el.cursorPos, el.cursorPos

	if el.selectionStarted {
		start = el.selectionStart
		if start > end {
			start, end = end, start
		}
	}

	el.textBefore = el.runes[:start]
	el.textSelection = el.runes[start:end]
	el.textAfter = el.runes[end:]
}

func (el *InputElement) selectionRange() (start, end int) {
	start, end = el.cursorPos, el.cursorPos

	if el.selectionStarted {
		start = el.selectionStart
		if start > end {
			start, end = end, start
		}
	}

	return start, end
}

func (el *InputElement) hasSelection() bool {
	start, end := el.selectionRange()
	return start != end
}

func (el *InputElement) clearSelection() {
	el.selectionStarted = false
	el.selectionStart = 0
}

func (el *InputElement) isCharValid(char rune) bool {
	if el.Charset == nil {
		return true
	}

	return el.Charset.Contains(char)
}

// replaceRange is the only method that should assign el.Text during editing.
// insert is assumed to have already been charset-filtered.
func (el *InputElement) replaceRange(start, end int, insert []rune) {
	start = util.Clamp(start, 0, len(el.runes))
	end = util.Clamp(end, start, len(el.runes))

	remainingLength := len(el.runes) - (end - start)
	available := max(0, el.MaxTextLength-remainingLength)

	if len(insert) > available {
		insert = insert[:available]
	}

	next := make([]rune, 0, remainingLength+len(insert))
	next = append(next, el.runes[:start]...)
	next = append(next, insert...)
	next = append(next, el.runes[end:]...)

	el.Text = string(next)
	el.cursorPos = start + len(insert)
	el.clearSelection()
}

func (el *InputElement) replaceSelection(insert []rune) {
	start, end := el.selectionRange()
	el.replaceRange(start, end, insert)
}

func (el *InputElement) insertInput(input string) {
	if input == "" {
		return
	}

	insert := el.prepareInput(input)
	if len(insert) == 0 {
		return
	}

	el.replaceSelection(insert)
}

func (el *InputElement) applyEdit(edit func()) {
	oldText := el.Text

	edit()
	el.recalculateTextSplit()

	if el.Text == oldText {
		return
	}

	if el.Callback != nil {
		el.Callback(el.Text)

		// Protect against callbacks that modify Text.
		el.recalculateTextSplit()
	}
}

func (el *InputElement) deleteBackward() {
	start, end := el.selectionRange()

	if start != end {
		el.replaceRange(start, end, nil)
		return
	}

	if start > 0 {
		el.replaceRange(start-1, start, nil)
	}
}

func (el *InputElement) deleteForward() {
	start, end := el.selectionRange()

	if start != end {
		el.replaceRange(start, end, nil)
		return
	}

	if end < len(el.runes) {
		el.replaceRange(end, end+1, nil)
	}
}

func (el *InputElement) moveCursor(delta int, extendSelection bool) {
	start, end := el.selectionRange()

	if !extendSelection && start != end {
		if delta < 0 {
			el.cursorPos = start
		} else {
			el.cursorPos = end
		}

		el.clearSelection()
		el.recalculateTextSplit()
		return
	}

	if extendSelection && !el.selectionStarted {
		el.selectionStart = el.cursorPos
		el.selectionStarted = true
	}

	el.cursorPos = util.Clamp(
		el.cursorPos+delta,
		0,
		len(el.runes),
	)

	if el.selectionStarted && el.selectionStart == el.cursorPos {
		el.clearSelection()
	}

	el.recalculateTextSplit()
}

var clipboardNewlineReplacer = strings.NewReplacer(
	"\r\n", " ",
	"\n", " ",
	"\r", " ",
)

func (el *InputElement) update(deltaNano int64) {
	el.recalculateTextSplit()

	el.textWidth = rl.MeasureText(el.Text, el.TextSize)

	el.w, el.h = max(el.textWidth+el.Padding*2, el.Size().X), max(el.TextSize+el.Padding*2, el.Size().Y)

	pos := el.AbsolutePos()
	el.x, el.y, el.cx, el.cy = pos.X, pos.Y, pos.X+el.w/2, pos.Y+el.h/2

	mouseX, mouseY := int32(global.MousePosition.X), int32(global.MousePosition.Y)
	el.hovered = mouseX > el.x &&
		mouseX < el.x+el.w &&
		mouseY > el.y &&
		mouseY < el.y+el.h

	if el.hovered {
		global.MouseCursorState = rl.MouseCursorIBeam
	}

	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		prevClicked := el.clicked
		el.clicked = el.hovered
		if !prevClicked && el.clicked {
			el.cursorPos = len(el.runes)
		} else if prevClicked && !el.clicked {
			el.clearSelection()
		}
	}

	if el.clicked {
		ctrlOrCmd := rl.IsKeyDown(rl.KeyLeftControl) ||
			rl.IsKeyDown(rl.KeyRightControl) ||
			rl.IsKeyDown(rl.KeyLeftSuper) ||
			rl.IsKeyDown(rl.KeyRightSuper)

		shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)

		left := rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressedRepeat(rl.KeyLeft)
		right := rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressedRepeat(rl.KeyRight)

		backspace := rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyPressedRepeat(rl.KeyBackspace)

		del := rl.IsKeyPressed(rl.KeyDelete) || rl.IsKeyPressedRepeat(rl.KeyDelete)

		ctrlA := ctrlOrCmd && rl.IsKeyPressed(rl.KeyA)
		ctrlV := ctrlOrCmd && rl.IsKeyPressed(rl.KeyV)
		ctrlC, ctrlX := ctrlOrCmd && rl.IsKeyPressed(rl.KeyC), ctrlOrCmd && rl.IsKeyPressed(rl.KeyX)

		switch {
		case ctrlA:
			if len(el.runes) == 0 {
				el.clearSelection()
			} else {
				el.selectionStarted = true
				el.selectionStart = 0
				el.cursorPos = len(el.runes)
				el.recalculateTextSplit()
			}

		case ctrlV:
			text := clipboardNewlineReplacer.Replace(rl.GetClipboardText())
			el.applyEdit(func() {
				el.insertInput(text)
			})

		case ctrlC && el.hasSelection():
			rl.SetClipboardText(string(el.textSelection))

		case ctrlX && el.hasSelection():
			rl.SetClipboardText(string(el.textSelection))
			el.applyEdit(func() {
				el.replaceSelection(nil)
			})

		case left && !right:
			el.moveCursor(-1, shift)

		case right && !left:
			el.moveCursor(1, shift)

		case backspace:
			el.applyEdit(el.deleteBackward)

		case del:
			el.applyEdit(el.deleteForward)

		default:
			var input strings.Builder

			for char := rl.GetCharPressed(); char != 0; char = rl.GetCharPressed() {
				input.WriteRune(char)
			}

			if input.Len() > 0 {
				el.applyEdit(func() {
					el.insertInput(input.String())
				})
			}
		}
	}
}

func (el *InputElement) caretOffsetAt(index int) int32 {
	if index <= 0 {
		return 0
	}

	width := rl.MeasureText(string(el.runes[:index]), el.TextSize)

	if index < len(el.runes) {
		width += el.TextSize / 20
	}

	return width
}

func (el *InputElement) draw() {
	btnWidthOuter, btnHeightOuter := el.w+el.OutlineWidth*2, el.h+el.OutlineWidth*2
	btnStartXOuter, btnStartYOuter := el.x-el.OutlineWidth, el.y-el.OutlineWidth

	oCol := el.OutlineColors.Color(util.DefaultState)
	pCol := el.PlaceholderColors.Color(util.DefaultState)
	bgCol := el.BackgroundColors.Color(util.DefaultState)
	fgCol := el.ForegroundColors.Color(util.DefaultState)

	if el.hovered {
		oCol = el.OutlineColors.Color(util.HoverState)
		pCol = el.PlaceholderColors.Color(util.HoverState)
		bgCol = el.BackgroundColors.Color(util.HoverState)
		fgCol = el.ForegroundColors.Color(util.HoverState)
	}

	if el.clicked {
		oCol = el.OutlineColors.Color(util.ClickState)
		pCol = el.PlaceholderColors.Color(util.ClickState)
		bgCol = el.BackgroundColors.Color(util.ClickState)
		fgCol = el.ForegroundColors.Color(util.ClickState)
	}

	rl.DrawRectangle(btnStartXOuter, btnStartYOuter, btnWidthOuter, btnHeightOuter, *oCol)

	rl.DrawRectangle(el.x, el.y, el.w, el.h, *bgCol)

	textY := el.cy - el.TextSize/2

	if el.Text == "" {
		rl.DrawText(el.PlaceholderText, el.x+el.Padding, textY, el.TextSize, *pCol)
	} else {
		if el.selectionStarted && el.selectionStart != el.cursorPos {
			start := min(el.selectionStart, el.cursorPos)
			end := max(el.selectionStart, el.cursorPos)

			startX := el.caretOffsetAt(start)
			endX := el.caretOffsetAt(end)

			rl.DrawRectangle(el.x+el.Padding+startX, textY, endX-startX, el.TextSize, rl.SkyBlue)
		}

		rl.DrawText(el.Text, el.x+el.Padding, textY, el.TextSize, *fgCol)
	}

	if el.clicked && int(rl.GetTime()*2)%2 == 0 {
		cursorWidth := max(rl.MeasureText("|", el.TextSize)/2, 1)
		cursorX := el.caretOffsetAt(el.cursorPos)

		rl.DrawRectangle(el.x+el.Padding+cursorX, textY-cursorWidth/2, cursorWidth, el.TextSize+cursorWidth, *fgCol)
	}
}
