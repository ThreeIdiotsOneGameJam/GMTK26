package ui

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type Element interface {
	update(deltaNano int64)
	draw()

	updateTree(deltaNano int64)
	drawTree()

	Base() *ElementBase

	RootElement() Element
	RelativePos() vec.Vec2i
	Size() vec.Vec2i
	AbsolutePos() vec.Vec2i
}

type ElementBase struct {
	Parent   Element
	Children []Element

	SelfAnchor   anchor.Anchor
	ParentAnchor anchor.Anchor
}

type BaseElement[T Element] struct {
	ElementBase

	self T

	// RelativePos is the offset from the parent anchor.
	RelativePosProvider func(el T) vec.Vec2i
	SizeProvider        func(el T) vec.Vec2i
}

// NewBaseElement initializes the self-reference required by fluent methods.
func NewBaseElement[T Element](self T) BaseElement[T] {
	return BaseElement[T]{
		self: self,
	}
}

func (el *BaseElement[T]) update(deltaNano int64) {}

func (el *BaseElement[T]) draw() {}

func (el *BaseElement[T]) updateTree(deltaNano int64) {
	el.self.update(deltaNano)

	for _, child := range el.Children {
		child.updateTree(deltaNano)
	}
}

func (el *BaseElement[T]) drawTree() {
	el.self.draw()

	for _, child := range el.Children {
		child.drawTree()
	}
}

func (el *BaseElement[T]) Base() *ElementBase {
	return &el.ElementBase
}

func (el *BaseElement[T]) WithRelativePos(
	relativePos vec.Vec2i,
) T {
	el.RelativePosProvider = func(T) vec.Vec2i {
		return relativePos
	}

	return el.self
}

func (el *BaseElement[T]) WithRelativePosDynamic(
	relativePosProvider func(el T) vec.Vec2i,
) T {
	el.RelativePosProvider = relativePosProvider
	return el.self
}

func (el *BaseElement[T]) WithSize(
	size vec.Vec2i,
) T {
	el.SizeProvider = func(T) vec.Vec2i {
		return size
	}

	return el.self
}

func (el *BaseElement[T]) WithSizeDynamic(
	sizeProvider func(el T) vec.Vec2i,
) T {
	el.SizeProvider = sizeProvider
	return el.self
}

func (el *BaseElement[T]) WithSelfAnchor(
	selfAnchor anchor.Anchor,
) T {
	el.SelfAnchor = selfAnchor
	return el.self
}

func (el *BaseElement[T]) WithParentAnchor(
	parentAnchor anchor.Anchor,
) T {
	el.ParentAnchor = parentAnchor
	return el.self
}

func (el *BaseElement[T]) WithAnchors(
	selfAnchor anchor.Anchor,
	parentAnchor anchor.Anchor,
) T {
	el.SelfAnchor = selfAnchor
	el.ParentAnchor = parentAnchor

	return el.self
}

func (el *BaseElement[T]) RootElement() Element {
	if el == nil {
		return nil
	}

	if el.Parent == nil {
		return el.self
	}

	return el.Parent.RootElement()
}

func (el *BaseElement[T]) RelativePos() vec.Vec2i {
	if el == nil || el.RelativePosProvider == nil {
		return vec.Vec2i{}
	}

	return el.RelativePosProvider(el.self)
}

func (el *BaseElement[T]) Size() vec.Vec2i {
	if el == nil || el.SizeProvider == nil {
		return vec.Vec2i{}
	}

	return el.SizeProvider(el.self)
}

func (el *BaseElement[T]) AbsolutePos() vec.Vec2i {
	if el == nil {
		return vec.Vec2i{}
	}

	relativePos := el.RelativePos()

	if el.Parent == nil {
		return relativePos
	}

	parentSize := el.Parent.Size()
	selfSize := el.Size()

	parentAnchorOffset := parentSize.
		Vec2().
		Mul(anchor.AnchorCoords[el.ParentAnchor]).
		RoundToInt()

	selfAnchorOffset := selfSize.
		Vec2().
		Mul(anchor.AnchorCoords[el.SelfAnchor]).
		RoundToInt()

	return el.Parent.
		AbsolutePos().
		Add(parentAnchorOffset).
		Add(relativePos).
		Sub(selfAnchorOffset)
}

// removeChildBase rebuilds base's Children slice without any instances of childBase.
func (base *ElementBase) removeChildBase(childBase *ElementBase) (removed bool) {
	removed = false
	children := base.Children[:0]

	for _, child := range base.Children {
		if child.Base() == childBase {
			removed = true
			continue
		}

		children = append(children, child)
	}

	clear(base.Children[len(children):])
	base.Children = children

	return removed
}

func (el *BaseElement[T]) AddChild(child Element) T {
	if child == nil {
		return el.self
	}

	childBase := child.Base()

	// prevent children becoming parents of themselves or their ancestors
	for current := Element(el.self); current != nil; current = current.Base().Parent {
		if current.Base() == childBase {
			panic("ui: adding child would create an element cycle")
		}
	}

	// prevent duplicating if already attached
	if oldParent := childBase.Parent; oldParent != nil &&
		oldParent.Base() == el.Base() {
		return el.self
	}

	// detach from the previous parent before reparenting
	if oldParent := childBase.Parent; oldParent != nil {
		oldParent.Base().removeChildBase(childBase)
	}

	childBase.Parent = el.self
	el.Children = append(el.Children, child)

	return el.self
}

func (el *BaseElement[T]) RemoveChild(child Element) T {
	if child == nil {
		return el.self
	}

	childBase := child.Base()

	if el.removeChildBase(childBase) &&
		childBase.Parent != nil &&
		childBase.Parent.Base() == el.Base() {
		childBase.Parent = nil
	}

	return el.self
}
