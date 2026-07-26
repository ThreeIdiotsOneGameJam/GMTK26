package ui

import "testing"

func TestInputFocusIsExclusive(t *testing.T) {
	first := Input()
	second := Input()
	t.Cleanup(func() {
		setFocusedInput(nil)
	})

	first.Focus()
	if !first.Focused() {
		t.Fatal("first input did not receive focus")
	}

	second.Focus()
	if first.Focused() {
		t.Fatal("first input retained focus")
	}
	if !second.Focused() {
		t.Fatal("second input did not receive focus")
	}
}

func TestInputFocusTraversalSkipsDisabledInputs(t *testing.T) {
	screen := Screen()
	first := Input()
	disabled := Input().WithEnabled(false)
	third := Input()
	screen.AddChildren(first, disabled, third)
	t.Cleanup(func() {
		setFocusedInput(nil)
	})

	first.Focus()
	focusAdjacentInput(first, editorForward)
	if !third.Focused() {
		t.Fatal("forward traversal did not skip disabled input")
	}

	focusAdjacentInput(third, editorBackward)
	if !first.Focused() {
		t.Fatal("backward traversal did not return to first input")
	}
}

func TestScreenExitClearsOnlyItsInputFocus(t *testing.T) {
	firstScreen := Screen()
	first := Input()
	firstScreen.AddChild(first)

	secondScreen := Screen()
	second := Input()
	secondScreen.AddChild(second)
	t.Cleanup(func() {
		setFocusedInput(nil)
	})

	second.Focus()
	firstScreen.Exit()
	if !second.Focused() {
		t.Fatal("exiting another screen cleared focus")
	}

	secondScreen.Exit()
	if second.Focused() {
		t.Fatal("exiting focused screen retained focus")
	}
}
