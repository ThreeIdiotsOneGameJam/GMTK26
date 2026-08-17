package screens

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestTooltipPositionStaysInsideViewport(t *testing.T) {
	viewport := vec.Vec2i{X: 768, Y: 576}
	tooltip := vec.Vec2i{X: 300, Y: 52}
	tests := []struct {
		name  string
		mouse vec.Vec2i
		want  vec.Vec2i
	}{
		{
			name:  "centered above cursor",
			mouse: vec.Vec2i{X: 384, Y: 300},
			want:  vec.Vec2i{X: 234, Y: 238},
		},
		{
			name:  "left edge",
			mouse: vec.Vec2i{X: 10, Y: 300},
			want:  vec.Vec2i{X: 0, Y: 238},
		},
		{
			name:  "right edge",
			mouse: vec.Vec2i{X: 760, Y: 300},
			want:  vec.Vec2i{X: 468, Y: 238},
		},
		{
			name:  "top edge flips below",
			mouse: vec.Vec2i{X: 384, Y: 5},
			want:  vec.Vec2i{X: 234, Y: 15},
		},
		{
			name:  "bottom edge clamps",
			mouse: vec.Vec2i{X: 384, Y: 575},
			want:  vec.Vec2i{X: 234, Y: 513},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := tooltipPosition(test.mouse, viewport, tooltip); got != test.want {
				t.Fatalf("tooltipPosition() = %+v, want %+v", got, test.want)
			}
		})
	}
}
