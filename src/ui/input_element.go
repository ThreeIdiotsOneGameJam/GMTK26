package ui

import (
	"math"
	"slices"
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
- most stuff could be refactored to be merged into just a few methods

FIXME: cases which trigger callbacks without any actual changes:
 del at end/backspace at start
 pasting empty/filtered text without a selection
 cutting zero-length selection
 replacing text with identical text

can be handled by storing oldText and only triggering callback on el.Text != oldText

FIXME: ctrlA on empty input creates zero width selection

FIXME: setting charset does not reverify existing text
FIXME: setting max length does not reverify existing text
FIXME: WithText can bypass all checks

FIXME: geometry/layout are not recalculated after processing text input, leaving them one frame behind edits - needs to be calculated again at end of update

FIXME: if callback modifies Text everything falls apart - need to recalculateTextSplit after
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

type InputElement struct {
	BaseElement[*InputElement]
	Text, PlaceholderText string
	MaxTextLength         int // in runes, NOT []byte len
	Charset               util.RuneSet
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

		paste := ""
		if ctrlV {
			paste = rl.GetClipboardText()
		}

		switch {
		case ctrlA:
			el.selectionStarted = true
			el.selectionStart = 0
			el.cursorPos = len(el.runes)
			el.recalculateTextSplit()
		case ctrlV && paste != "":
			paste = strings.NewReplacer(
				"\r\n", " ",
				"\n", " ",
				"\r", " ",
			).Replace(paste)
			insert := []rune(paste)

			filtered := insert[:0]
			for _, char := range insert {
				if el.isCharValid(char) {
					filtered = append(filtered, char)
				}
			}
			insert = filtered

			available := el.MaxTextLength -
				(len(el.runes) - len(el.textSelection))

			if available <= 0 {
				return
			}

			insert = insert[:min(len(insert), available)]

			if len(insert) == 0 {
				break
			}

			el.Text = string(slices.Concat(el.textBefore, insert, el.textAfter))
			el.cursorPos = len(el.textBefore) + len(insert)
			el.clearSelection()
			el.recalculateTextSplit()
			el.Callback(el.Text)
		case (ctrlC || ctrlX) && el.selectionStarted:
			rl.SetClipboardText(string(el.textSelection))
			if ctrlX {
				el.cursorPos = len(el.textBefore)
				el.Text = string(slices.Concat(el.textBefore, el.textAfter))
				el.clearSelection()
				el.recalculateTextSplit()
				el.Callback(el.Text)
			}
		case left && !right:
			if !shift && el.selectionStarted {
				el.cursorPos = min(el.cursorPos, el.selectionStart)
				el.clearSelection()
			} else if el.cursorPos > 0 {
				if shift && !el.selectionStarted {
					el.selectionStart = el.cursorPos
					el.selectionStarted = true
				}
				el.cursorPos--
			}
			if el.selectionStarted && el.selectionStart == el.cursorPos {
				el.clearSelection()
			}
			el.recalculateTextSplit()
		case right && !left:
			if !shift && el.selectionStarted {
				el.cursorPos = max(el.cursorPos, el.selectionStart)
				el.clearSelection()
			} else if el.cursorPos < len(el.runes) {
				if shift && !el.selectionStarted {
					el.selectionStart = el.cursorPos
					el.selectionStarted = true
				}
				el.cursorPos++
			}
			if el.selectionStarted && el.selectionStart == el.cursorPos {
				el.clearSelection()
			}
			el.recalculateTextSplit()
		case backspace:
			if el.selectionStarted && len(el.textSelection) > 0 {
				el.cursorPos = len(el.textBefore)
				el.Text = string(slices.Concat(el.textBefore, el.textAfter))
				el.clearSelection()
			} else if len(el.textBefore) > 0 {
				el.Text = string(slices.Concat(el.textBefore[:len(el.textBefore)-1], el.textAfter))
				el.cursorPos--
			}
			el.recalculateTextSplit()
			el.Callback(el.Text)
		case del:
			if el.selectionStarted && len(el.textSelection) > 0 {
				el.cursorPos = len(el.textBefore)
				el.Text = string(slices.Concat(el.textBefore, el.textAfter))
				el.clearSelection()
			} else if len(el.textAfter) > 0 {
				el.Text = string(slices.Concat(el.textBefore, el.textAfter[1:]))
			}
			el.recalculateTextSplit()
			el.Callback(el.Text)
		default:
			for char := rl.GetCharPressed(); char != 0; char = rl.GetCharPressed() {
				if !el.isCharValid(char) {
					continue
				}
				insert := []rune{char}
				available := el.MaxTextLength -
					(len(el.runes) - len(el.textSelection))

				if available <= 0 {
					continue
				}

				insert = insert[:min(len(insert), available)]
				el.cursorPos = len(el.textBefore)
				el.Text = string(slices.Concat(el.textBefore, insert, el.textAfter))
				el.cursorPos++
				el.clearSelection()
				el.recalculateTextSplit()
				el.Callback(el.Text)
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
