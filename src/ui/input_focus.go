package ui

var focusedInput *InputElement

func (el *InputElement) Focused() bool {
	return focusedInput == el
}

func setFocusedInput(next *InputElement) {
	if focusedInput == next {
		return
	}

	if focusedInput != nil {
		focusedInput.editor.clearSelection()
	}

	focusedInput = next
}

func clearFocusedInputWithin(root Element) {
	if focusedInput == nil || focusedInput.RootElement() != root {
		return
	}

	setFocusedInput(nil)
}

func focusAdjacentInput(current *InputElement, direction editorDirection) {
	if current == nil {
		return
	}

	inputs := collectFocusableInputs(current.RootElement(), nil)
	if len(inputs) == 0 {
		current.Blur()
		return
	}

	index := 0
	for candidateIndex, candidate := range inputs {
		if candidate == current {
			index = candidateIndex
			break
		}
	}

	if direction == editorBackward {
		index = (index - 1 + len(inputs)) % len(inputs)
	} else {
		index = (index + 1) % len(inputs)
	}
	inputs[index].Focus()
}

func collectFocusableInputs(element Element, inputs []*InputElement) []*InputElement {
	if element == nil || !element.Visible() || !element.Enabled() {
		return inputs
	}

	if input, ok := element.(*InputElement); ok {
		inputs = append(inputs, input)
	}

	for _, child := range element.Base().Children {
		inputs = collectFocusableInputs(child, inputs)
	}
	return inputs
}
