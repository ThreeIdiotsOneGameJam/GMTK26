package ui

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestElementRectContainment(t *testing.T) {
	rect := elementRect{X: 10, Y: 20, Width: 30, Height: 40}

	if !rect.contains(vec.Vec2i{X: 10, Y: 20}) {
		t.Fatal("inclusive bounds should contain their top-left edge")
	}
	if rect.containsStrict(vec.Vec2i{X: 10, Y: 20}) {
		t.Fatal("strict bounds should exclude their top-left edge")
	}
	if rect.contains(vec.Vec2i{X: 41, Y: 20}) {
		t.Fatal("bounds should exclude points past the right edge")
	}
}

func TestControlStatePriority(t *testing.T) {
	tests := []struct {
		name                     string
		enabled, hovered, active bool
		want                     UIState
	}{
		{name: "default", enabled: true, want: StateDefault},
		{name: "hover", enabled: true, hovered: true, want: StateHover},
		{name: "active", enabled: true, hovered: true, active: true, want: StateClick},
		{name: "disabled wins", hovered: true, active: true, want: StateDisabled},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := controlState(test.enabled, test.hovered, test.active); got != test.want {
				t.Fatalf("controlState() = %v, want %v", got, test.want)
			}
		})
	}
}
