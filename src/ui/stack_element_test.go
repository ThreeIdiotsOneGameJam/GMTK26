package ui

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestVStackMeasuresAndPositionsChildren(t *testing.T) {
	first := Group().WithSize(vec.Vec2i{X: 20, Y: 10})
	second := Group().WithSize(vec.Vec2i{X: 10, Y: 20})
	stack := VStack(5, first, second).
		WithAlignment(StackCenter).
		WithPadding(2)

	if got, want := stack.Size(), (vec.Vec2i{X: 24, Y: 39}); got != want {
		t.Fatalf("stack size = %v, want %v", got, want)
	}
	if got, want := first.AbsolutePos(), (vec.Vec2i{X: 2, Y: 2}); got != want {
		t.Fatalf("first position = %v, want %v", got, want)
	}
	if got, want := second.AbsolutePos(), (vec.Vec2i{X: 7, Y: 17}); got != want {
		t.Fatalf("second position = %v, want %v", got, want)
	}
}
