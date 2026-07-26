package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type editorCommandKind uint8

const (
	editorCommandNone editorCommandKind = iota
	editorCommandSubmit
	editorCommandSelectAll
	editorCommandPaste
	editorCommandCopy
	editorCommandCut
	editorCommandMove
	editorCommandFocusNext
	editorCommandDelete
	editorCommandInsert
)

type editorCommand struct {
	kind      editorCommandKind
	direction editorDirection
	unit      editorMoveUnit
	extend    bool
	text      string
}

type editorKeyState struct {
	ctrlOrCmd    bool
	wordModifier bool
	shift        bool

	left      bool
	right     bool
	backspace bool
	delete    bool
	tab       bool
	enter     bool
	keyA      bool
	keyC      bool
	keyV      bool
	keyX      bool
	text      string
}

func pollEditorCommand(canSubmit bool) editorCommand {
	ctrl := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	super := rl.IsKeyDown(rl.KeyLeftSuper) || rl.IsKeyDown(rl.KeyRightSuper)
	alt := rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt)

	var input strings.Builder
	for char := rl.GetCharPressed(); char != 0; char = rl.GetCharPressed() {
		input.WriteRune(char)
	}

	return editorCommandFromKeyState(editorKeyState{
		ctrlOrCmd:    ctrl || super,
		wordModifier: ctrl || alt,
		shift:        rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift),
		left:         keyPressedOrRepeated(rl.KeyLeft),
		right:        keyPressedOrRepeated(rl.KeyRight),
		backspace:    keyPressedOrRepeated(rl.KeyBackspace),
		delete:       keyPressedOrRepeated(rl.KeyDelete),
		tab:          rl.IsKeyPressed(rl.KeyTab),
		enter:        rl.IsKeyPressed(rl.KeyEnter),
		keyA:         rl.IsKeyPressed(rl.KeyA),
		keyC:         rl.IsKeyPressed(rl.KeyC),
		keyV:         rl.IsKeyPressed(rl.KeyV),
		keyX:         rl.IsKeyPressed(rl.KeyX),
		text:         input.String(),
	}, canSubmit)
}

func editorCommandFromKeyState(keys editorKeyState, canSubmit bool) editorCommand {
	unit := editorByGrapheme
	if keys.wordModifier {
		unit = editorByWord
	}

	switch {
	case canSubmit && keys.enter:
		return editorCommand{kind: editorCommandSubmit}
	case keys.ctrlOrCmd && keys.keyA:
		return editorCommand{kind: editorCommandSelectAll}
	case keys.ctrlOrCmd && keys.keyV:
		return editorCommand{kind: editorCommandPaste}
	case keys.ctrlOrCmd && keys.keyC:
		return editorCommand{kind: editorCommandCopy}
	case keys.ctrlOrCmd && keys.keyX:
		return editorCommand{kind: editorCommandCut}
	case keys.tab:
		direction := editorForward
		if keys.shift {
			direction = editorBackward
		}
		return editorCommand{
			kind:      editorCommandFocusNext,
			direction: direction,
		}
	case keys.left && !keys.right:
		return editorCommand{
			kind:      editorCommandMove,
			direction: editorBackward,
			unit:      unit,
			extend:    keys.shift,
		}
	case keys.right && !keys.left:
		return editorCommand{
			kind:      editorCommandMove,
			direction: editorForward,
			unit:      unit,
			extend:    keys.shift,
		}
	case keys.backspace:
		return editorCommand{
			kind:      editorCommandDelete,
			direction: editorBackward,
			unit:      unit,
		}
	case keys.delete:
		return editorCommand{
			kind:      editorCommandDelete,
			direction: editorForward,
			unit:      unit,
		}
	case keys.text != "":
		return editorCommand{kind: editorCommandInsert, text: keys.text}
	default:
		return editorCommand{}
	}
}

func keyPressedOrRepeated(key int32) bool {
	return rl.IsKeyPressed(key) || rl.IsKeyPressedRepeat(key)
}
