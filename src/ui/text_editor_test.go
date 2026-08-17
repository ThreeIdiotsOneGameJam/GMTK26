package ui

import "testing"

func TestTextEditorMovesAndDeletesByGrapheme(t *testing.T) {
	var editor textEditorModel
	editor.syncText("a\u0301b")

	editor.move(editorForward, editorByGrapheme, false)
	if got, want := editor.caretPosition(), 2; got != want {
		t.Fatalf("caret after first grapheme = %d, want %d", got, want)
	}

	if changed := editor.delete(editorBackward, editorByGrapheme); !changed {
		t.Fatal("backspace did not report a text change")
	}
	if got, want := editor.string(), "b"; got != want {
		t.Fatalf("text after grapheme backspace = %q, want %q", got, want)
	}
}

func TestTextEditorSelectionUsesWholeGraphemes(t *testing.T) {
	var editor textEditorModel
	editor.syncText("A👨‍👩‍👧‍👦B")

	editor.move(editorForward, editorByGrapheme, false)
	editor.move(editorForward, editorByGrapheme, true)

	if got, want := editor.selectedText(), "👨‍👩‍👧‍👦"; got != want {
		t.Fatalf("selected text = %q, want %q", got, want)
	}
}

func TestTextEditorMovesAndDeletesByWord(t *testing.T) {
	var editor textEditorModel
	editor.syncText("hello, world")

	editor.move(editorForward, editorByWord, false)
	if got, want := editor.caretPosition(), 5; got != want {
		t.Fatalf("caret after first word = %d, want %d", got, want)
	}

	editor.move(editorForward, editorByWord, false)
	if got, want := editor.caretPosition(), len([]rune("hello, world")); got != want {
		t.Fatalf("caret after second word = %d, want %d", got, want)
	}

	if changed := editor.delete(editorBackward, editorByWord); !changed {
		t.Fatal("word backspace did not report a text change")
	}
	if got, want := editor.string(), "hello, "; got != want {
		t.Fatalf("text after word backspace = %q, want %q", got, want)
	}
}

func TestTextEditorReplacementHonorsLimitWithoutSplittingGrapheme(t *testing.T) {
	var editor textEditorModel

	editor.replaceSelection([]rune("A👨‍👩‍👧‍👦B"), 3)
	if got, want := editor.string(), "A"; got != want {
		t.Fatalf("limited text = %q, want %q", got, want)
	}
}
