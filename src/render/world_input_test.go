package render

import (
	"testing"

	v "github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestLeftGestureQuickPressClicks(t *testing.T) {
	r := WorldRenderer{MousePosition: v.Vec2{X: 100, Y: 100}}

	clicked, panning := r.updateLeftGesture(0.01, true, true, false)
	if clicked || panning {
		t.Fatalf("press = clicked %v, panning %v; want false, false", clicked, panning)
	}

	r.MousePosition = v.Vec2{X: 102, Y: 101}
	clicked, panning = r.updateLeftGesture(0.05, false, false, false)
	if !clicked || panning {
		t.Fatalf("release = clicked %v, panning %v; want true, false", clicked, panning)
	}
}

func TestLeftGestureHoldPansWithoutClicking(t *testing.T) {
	r := WorldRenderer{MousePosition: v.Vec2{X: 100, Y: 100}}

	r.updateLeftGesture(0.01, true, true, false)
	clicked, panning := r.updateLeftGesture(leftPanHoldSeconds, false, true, false)
	if clicked || !panning {
		t.Fatalf("hold = clicked %v, panning %v; want false, true", clicked, panning)
	}

	clicked, panning = r.updateLeftGesture(0.01, false, false, false)
	if clicked || panning {
		t.Fatalf("release = clicked %v, panning %v; want false, false", clicked, panning)
	}
}

func TestLeftGestureDragPansBeforeHoldDelay(t *testing.T) {
	start := v.Vec2{X: 100, Y: 100}
	r := WorldRenderer{MousePosition: start}

	r.updateLeftGesture(0.01, true, true, false)
	r.MousePosition.X += leftPanDragThreshold
	clicked, panning := r.updateLeftGesture(0.01, false, true, false)
	if clicked || !panning {
		t.Fatalf("drag = clicked %v, panning %v; want false, true", clicked, panning)
	}
	if r.PanStart != start {
		t.Fatalf("PanStart = %+v; want %+v", r.PanStart, start)
	}
}

func TestLeftGestureBlockedPressNeverCapturesWorld(t *testing.T) {
	r := WorldRenderer{MousePosition: v.Vec2{X: 100, Y: 100}}

	clicked, panning := r.updateLeftGesture(0.01, true, true, true)
	if clicked || panning {
		t.Fatalf("blocked press = clicked %v, panning %v; want false, false", clicked, panning)
	}

	clicked, panning = r.updateLeftGesture(0.01, false, false, false)
	if clicked || panning {
		t.Fatalf("release = clicked %v, panning %v; want false, false", clicked, panning)
	}
}

func TestLeftGestureReleaseOverUIDoesNotClickWorld(t *testing.T) {
	r := WorldRenderer{MousePosition: v.Vec2{X: 100, Y: 100}}

	r.updateLeftGesture(0.01, true, true, false)
	clicked, panning := r.updateLeftGesture(0.05, false, false, true)
	if clicked || panning {
		t.Fatalf("blocked release = clicked %v, panning %v; want false, false", clicked, panning)
	}
}
