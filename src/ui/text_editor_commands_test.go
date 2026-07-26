package ui

import "testing"

func TestEditorCommandFromKeyState(t *testing.T) {
	tests := []struct {
		name      string
		keys      editorKeyState
		canSubmit bool
		want      editorCommand
	}{
		{
			name:      "submit takes priority",
			keys:      editorKeyState{enter: true, text: "ignored"},
			canSubmit: true,
			want:      editorCommand{kind: editorCommandSubmit},
		},
		{
			name: "word selection",
			keys: editorKeyState{
				wordModifier: true,
				shift:        true,
				left:         true,
			},
			want: editorCommand{
				kind:      editorCommandMove,
				direction: editorBackward,
				unit:      editorByWord,
				extend:    true,
			},
		},
		{
			name: "copy shortcut suppresses text",
			keys: editorKeyState{
				ctrlOrCmd: true,
				keyC:      true,
				text:      "c",
			},
			want: editorCommand{kind: editorCommandCopy},
		},
		{
			name: "word delete",
			keys: editorKeyState{
				wordModifier: true,
				backspace:    true,
			},
			want: editorCommand{
				kind:      editorCommandDelete,
				direction: editorBackward,
				unit:      editorByWord,
			},
		},
		{
			name: "reverse focus traversal",
			keys: editorKeyState{
				shift: true,
				tab:   true,
			},
			want: editorCommand{
				kind:      editorCommandFocusNext,
				direction: editorBackward,
			},
		},
		{
			name: "text insertion",
			keys: editorKeyState{text: "hello"},
			want: editorCommand{kind: editorCommandInsert, text: "hello"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := editorCommandFromKeyState(test.keys, test.canSubmit); got != test.want {
				t.Fatalf("command = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestInputExecutesGraphemeDeleteCommand(t *testing.T) {
	callbackText := ""
	input := Input().
		WithText("a\u0301b").
		WithCallback(func(text string) {
			callbackText = text
		})
	t.Cleanup(func() {
		setFocusedInput(nil)
	})
	input.Focus()

	input.executeEditorCommand(editorCommand{
		kind:      editorCommandDelete,
		direction: editorBackward,
		unit:      editorByGrapheme,
	})

	if got, want := input.Value(), "a\u0301"; got != want {
		t.Fatalf("input value = %q, want %q", got, want)
	}
	if got, want := callbackText, "a\u0301"; got != want {
		t.Fatalf("callback text = %q, want %q", got, want)
	}
}
