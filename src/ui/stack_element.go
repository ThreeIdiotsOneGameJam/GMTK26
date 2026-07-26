package ui

import "github.com/threeidiotsonegamejam/gmtk26/src/util/vec"

type StackDirection uint8

const (
	StackVertical StackDirection = iota
	StackHorizontal
)

type StackAlignment uint8

const (
	StackStart StackAlignment = iota
	StackCenter
	StackEnd
)

func VStack(gap int32, children ...Element) *StackElement {
	return newStack(StackVertical, gap, children...)
}

func HStack(gap int32, children ...Element) *StackElement {
	return newStack(StackHorizontal, gap, children...)
}

func newStack(direction StackDirection, gap int32, children ...Element) *StackElement {
	el := &StackElement{
		Direction: direction,
		Alignment: StackStart,
		Gap:       max(gap, 0),
	}
	el.BaseElement = NewBaseElement(el)
	el.WithSizeDynamic(func(el *StackElement) vec.Vec2i {
		return el.contentSize()
	})
	return el.AddChildren(children...)
}

type StackElement struct {
	BaseElement[*StackElement]

	Direction StackDirection
	Alignment StackAlignment
	Gap       int32
	Padding   int32
}

func (el *StackElement) WithAlignment(alignment StackAlignment) *StackElement {
	el.Alignment = alignment
	return el
}

func (el *StackElement) WithGap(gap int32) *StackElement {
	el.Gap = max(gap, 0)
	return el
}

func (el *StackElement) WithPadding(padding int32) *StackElement {
	el.Padding = max(padding, 0)
	return el
}

func (el *StackElement) contentSize() vec.Vec2i {
	mainSize := int32(0)
	crossSize := int32(0)
	visibleChildren := int32(0)

	for _, child := range el.Children {
		if !child.Visible() {
			continue
		}

		size := child.Size()
		if el.Direction == StackVertical {
			mainSize += size.Y
			crossSize = max(crossSize, size.X)
		} else {
			mainSize += size.X
			crossSize = max(crossSize, size.Y)
		}
		visibleChildren++
	}

	if visibleChildren > 1 {
		mainSize += (visibleChildren - 1) * el.Gap
	}
	mainSize += el.Padding * 2
	crossSize += el.Padding * 2

	if el.Direction == StackVertical {
		return vec.Vec2i{X: crossSize, Y: mainSize}
	}
	return vec.Vec2i{X: mainSize, Y: crossSize}
}

func (el *StackElement) ChildOffset(target Element) vec.Vec2i {
	cursor := el.Padding
	stackSize := el.Size()

	for _, child := range el.Children {
		if !child.Visible() {
			continue
		}

		childSize := child.Size()
		if child == target {
			crossOffset := el.Padding
			if el.Direction == StackVertical {
				crossOffset = alignedOffset(stackSize.X, childSize.X, el.Padding, el.Alignment)
				return vec.Vec2i{X: crossOffset, Y: cursor}
			}

			crossOffset = alignedOffset(stackSize.Y, childSize.Y, el.Padding, el.Alignment)
			return vec.Vec2i{X: cursor, Y: crossOffset}
		}

		if el.Direction == StackVertical {
			cursor += childSize.Y + el.Gap
		} else {
			cursor += childSize.X + el.Gap
		}
	}

	return vec.Vec2i{}
}

func alignedOffset(container, child, padding int32, alignment StackAlignment) int32 {
	switch alignment {
	case StackCenter:
		return (container - child) / 2
	case StackEnd:
		return container - child - padding
	default:
		return padding
	}
}

func (el *StackElement) draw() {}
