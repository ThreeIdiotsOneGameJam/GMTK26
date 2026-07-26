package ui

import (
	"slices"

	"github.com/go-text/typesetting/segmenter"
)

type editorDirection int

const (
	editorBackward editorDirection = -1
	editorForward  editorDirection = 1
)

type editorMoveUnit uint8

const (
	editorByGrapheme editorMoveUnit = iota
	editorByWord
)

// textEditorModel owns text-editing state without depending on raylib input or
// rendering. Positions are rune offsets so they match go-text's segmenter API.
type textEditorModel struct {
	text      []rune
	caret     int
	anchor    int
	selecting bool

	graphemeStops []int
	words         []textRange
	segmenter     segmenter.Segmenter
}

type textRange struct {
	start int
	end   int
}

func (editor *textEditorModel) syncText(text string) {
	runes := []rune(text)
	if slices.Equal(editor.text, runes) {
		return
	}

	editor.text = runes
	editor.rebuildBoundaries()
	editor.caret = editor.graphemeAtOrBefore(clampEditorPosition(editor.caret, len(editor.text)))
	editor.anchor = editor.graphemeAtOrBefore(clampEditorPosition(editor.anchor, len(editor.text)))

	if editor.selecting && editor.anchor == editor.caret {
		editor.clearSelection()
	}
}

func (editor *textEditorModel) string() string {
	return string(editor.text)
}

func (editor *textEditorModel) runes() []rune {
	return editor.text
}

func (editor *textEditorModel) caretPosition() int {
	return editor.caret
}

func (editor *textEditorModel) hasSelection() bool {
	start, end := editor.selectionRange()
	return start != end
}

func (editor *textEditorModel) selectionRange() (start, end int) {
	start, end = editor.caret, editor.caret
	if editor.selecting {
		start = editor.anchor
		if start > end {
			start, end = end, start
		}
	}

	return start, end
}

func (editor *textEditorModel) selectedText() string {
	start, end := editor.selectionRange()
	return string(editor.text[start:end])
}

func (editor *textEditorModel) clearSelection() {
	editor.selecting = false
	editor.anchor = 0
}

func (editor *textEditorModel) selectAll() {
	if len(editor.text) == 0 {
		editor.clearSelection()
		return
	}

	editor.anchor = 0
	editor.caret = len(editor.text)
	editor.selecting = true
}

func (editor *textEditorModel) moveToEnd() {
	editor.caret = len(editor.text)
	editor.clearSelection()
}

func (editor *textEditorModel) move(
	direction editorDirection,
	unit editorMoveUnit,
	extendSelection bool,
) {
	start, end := editor.selectionRange()
	if !extendSelection && start != end {
		if direction == editorBackward {
			editor.caret = start
		} else {
			editor.caret = end
		}
		editor.clearSelection()
		return
	}

	if extendSelection && !editor.selecting {
		editor.anchor = editor.caret
		editor.selecting = true
	}

	editor.caret = editor.destination(direction, unit)
	if editor.selecting && editor.anchor == editor.caret {
		editor.clearSelection()
	}
}

func (editor *textEditorModel) delete(direction editorDirection, unit editorMoveUnit) bool {
	start, end := editor.selectionRange()
	if start == end {
		target := editor.destination(direction, unit)
		if direction == editorBackward {
			start = target
		} else {
			end = target
		}
	}

	if start == end {
		return false
	}

	return editor.replaceRange(start, end, nil)
}

func (editor *textEditorModel) replaceSelection(insert []rune, maxLength int) bool {
	start, end := editor.selectionRange()
	remainingLength := len(editor.text) - (end - start)
	available := max(0, maxLength-remainingLength)
	insert = graphemeSafePrefix(insert, available)

	return editor.replaceRange(start, end, insert)
}

func (editor *textEditorModel) replaceRange(start, end int, insert []rune) bool {
	start = clampEditorPosition(start, len(editor.text))
	end = clampEditorPosition(end, len(editor.text))
	if start > end {
		start, end = end, start
	}

	next := make([]rune, 0, len(editor.text)-(end-start)+len(insert))
	next = append(next, editor.text[:start]...)
	next = append(next, insert...)
	next = append(next, editor.text[end:]...)

	changed := !slices.Equal(editor.text, next)
	editor.text = next
	editor.rebuildBoundaries()
	editor.caret = editor.graphemeAtOrAfter(start + len(insert))
	editor.clearSelection()
	return changed
}

func (editor *textEditorModel) destination(
	direction editorDirection,
	unit editorMoveUnit,
) int {
	switch unit {
	case editorByWord:
		if direction == editorBackward {
			for index := len(editor.words) - 1; index >= 0; index-- {
				if editor.words[index].start < editor.caret {
					return editor.words[index].start
				}
			}
			return 0
		}

		for _, word := range editor.words {
			if word.end > editor.caret {
				return word.end
			}
		}
		return len(editor.text)

	default:
		if direction == editorBackward {
			return editor.graphemeBefore(editor.caret)
		}
		return editor.graphemeAfter(editor.caret)
	}
}

func (editor *textEditorModel) rebuildBoundaries() {
	editor.segmenter.Init(editor.text)

	editor.graphemeStops = append(editor.graphemeStops[:0], 0)
	graphemes := editor.segmenter.GraphemeIterator()
	for graphemes.Next() {
		grapheme := graphemes.Grapheme()
		end := grapheme.Offset + len(grapheme.Text)
		if end > editor.graphemeStops[len(editor.graphemeStops)-1] {
			editor.graphemeStops = append(editor.graphemeStops, end)
		}
	}

	if last := editor.graphemeStops[len(editor.graphemeStops)-1]; last != len(editor.text) {
		editor.graphemeStops = append(editor.graphemeStops, len(editor.text))
	}

	editor.words = editor.words[:0]
	words := editor.segmenter.WordIterator()
	for words.Next() {
		word := words.Word()
		editor.words = append(editor.words, textRange{
			start: word.Offset,
			end:   word.Offset + len(word.Text),
		})
	}
}

func (editor *textEditorModel) graphemeBefore(position int) int {
	index, _ := slices.BinarySearch(editor.graphemeStops, position)
	if index == 0 {
		return 0
	}
	return editor.graphemeStops[index-1]
}

func (editor *textEditorModel) graphemeAfter(position int) int {
	index, found := slices.BinarySearch(editor.graphemeStops, position)
	if found {
		index++
	}
	if index >= len(editor.graphemeStops) {
		return len(editor.text)
	}
	return editor.graphemeStops[index]
}

func (editor *textEditorModel) graphemeAtOrBefore(position int) int {
	index, found := slices.BinarySearch(editor.graphemeStops, position)
	if found {
		return editor.graphemeStops[index]
	}
	if index == 0 {
		return 0
	}
	return editor.graphemeStops[index-1]
}

func (editor *textEditorModel) graphemeAtOrAfter(position int) int {
	index, _ := slices.BinarySearch(editor.graphemeStops, position)
	if index >= len(editor.graphemeStops) {
		return len(editor.text)
	}
	return editor.graphemeStops[index]
}

func graphemeSafePrefix(runes []rune, maxRunes int) []rune {
	if len(runes) <= maxRunes {
		return runes
	}
	if maxRunes <= 0 {
		return nil
	}

	var seg segmenter.Segmenter
	seg.Init(runes)

	end := 0
	graphemes := seg.GraphemeIterator()
	for graphemes.Next() {
		grapheme := graphemes.Grapheme()
		nextEnd := grapheme.Offset + len(grapheme.Text)
		if nextEnd > maxRunes {
			break
		}
		end = nextEnd
	}

	return runes[:end]
}

func clampEditorPosition(position, textLength int) int {
	return min(max(position, 0), textLength)
}
