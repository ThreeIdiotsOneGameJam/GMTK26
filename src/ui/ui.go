package ui

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type Element interface {
	prepare()
	update(deltaNano int64)
	draw()

	prepareTree()
	updateTree(deltaNano int64)
	drawTree()

	Base() *ElementBase

	RootElement() Element
	RelativePos() vec.Vec2i
	Size() vec.Vec2i
	AbsolutePos() vec.Vec2i
	Visible() bool
	Enabled() bool
	Opacity() float32
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
	VisibleProvider     func(el T) bool
	EnabledProvider     func(el T) bool
	OpacityProvider     func(el T) float32
}

// NewBaseElement initializes the self-reference required by fluent methods.
func NewBaseElement[T Element](self T) BaseElement[T] {
	return BaseElement[T]{
		self: self,
	}
}

func (el *BaseElement[T]) update(deltaNano int64) {}

func (el *BaseElement[T]) draw() {}

func (el *BaseElement[T]) prepare() {}

func (el *BaseElement[T]) prepareTree() {
	if !el.Visible() {
		return
	}

	el.self.prepare()

	for _, child := range el.Children {
		child.prepareTree()
	}
}

func (el *BaseElement[T]) updateTree(deltaNano int64) {
	if !el.Visible() {
		return
	}

	el.self.prepare()
	el.self.update(deltaNano)

	for _, child := range el.Children {
		child.updateTree(deltaNano)
	}
}

func (el *BaseElement[T]) drawTree() {
	if !el.Visible() || el.Opacity() <= 0 {
		return
	}

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

func (el *BaseElement[T]) WithVisible(visible bool) T {
	el.VisibleProvider = func(T) bool {
		return visible
	}
	return el.self
}

func (el *BaseElement[T]) WithVisibleDynamic(visibleProvider func(el T) bool) T {
	el.VisibleProvider = visibleProvider
	return el.self
}

func (el *BaseElement[T]) WithEnabled(enabled bool) T {
	el.EnabledProvider = func(T) bool {
		return enabled
	}
	return el.self
}

func (el *BaseElement[T]) WithEnabledDynamic(enabledProvider func(el T) bool) T {
	el.EnabledProvider = enabledProvider
	return el.self
}

func (el *BaseElement[T]) WithOpacity(opacity float32) T {
	el.OpacityProvider = func(T) float32 {
		return opacity
	}
	return el.self
}

func (el *BaseElement[T]) WithOpacityDynamic(opacityProvider func(el T) float32) T {
	el.OpacityProvider = opacityProvider
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

func (el *BaseElement[T]) Visible() bool {
	if el == nil {
		return false
	}

	if el.Parent != nil && !el.Parent.Visible() {
		return false
	}

	return el.VisibleProvider == nil || el.VisibleProvider(el.self)
}

func (el *BaseElement[T]) Enabled() bool {
	if el == nil {
		return false
	}

	if el.Parent != nil && !el.Parent.Enabled() {
		return false
	}

	return el.EnabledProvider == nil || el.EnabledProvider(el.self)
}

func (el *BaseElement[T]) Opacity() float32 {
	if el == nil {
		return 0
	}

	opacity := float32(1)
	if el.OpacityProvider != nil {
		opacity = el.OpacityProvider(el.self)
	}
	opacity = max(float32(0), min(opacity, float32(1)))

	if el.Parent != nil {
		opacity *= el.Parent.Opacity()
	}

	return opacity
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
