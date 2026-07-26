package render

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	v "github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestLeftGestureQuickPressClicks(t *testing.T) {
	r := WorldRenderer{MousePosition: v.Vec2{X: 100, Y: 100}}

	clicked, panning, released := r.updateLeftGesture(0.01, true, true, false)
	if clicked || panning || released {
		t.Fatalf("press = clicked %v, panning %v, released %v; want false, false, false", clicked, panning, released)
	}

	r.MousePosition = v.Vec2{X: 102, Y: 101}
	clicked, panning, released = r.updateLeftGesture(0.05, false, false, false)
	if !clicked || panning || released {
		t.Fatalf("release = clicked %v, panning %v, released %v; want true, false, false", clicked, panning, released)
	}
}

func TestLeftGestureHoldPansWithoutClicking(t *testing.T) {
	r := WorldRenderer{MousePosition: v.Vec2{X: 100, Y: 100}}

	r.updateLeftGesture(0.01, true, true, false)
	clicked, panning, released := r.updateLeftGesture(leftPanHoldSeconds, false, true, false)
	if clicked || !panning || released {
		t.Fatalf("hold = clicked %v, panning %v, released %v; want false, true, false", clicked, panning, released)
	}

	clicked, panning, released = r.updateLeftGesture(0.01, false, false, false)
	if clicked || panning || !released {
		t.Fatalf("release = clicked %v, panning %v, released %v; want false, false, true", clicked, panning, released)
	}
}

func TestLeftGestureDragPansBeforeHoldDelay(t *testing.T) {
	start := v.Vec2{X: 100, Y: 100}
	r := WorldRenderer{MousePosition: start}

	r.updateLeftGesture(0.01, true, true, false)
	r.MousePosition.X += leftPanDragThreshold
	clicked, panning, released := r.updateLeftGesture(0.01, false, true, false)
	if clicked || !panning || released {
		t.Fatalf("drag = clicked %v, panning %v, released %v; want false, true, false", clicked, panning, released)
	}
	if r.PanStart != start {
		t.Fatalf("PanStart = %+v; want %+v", r.PanStart, start)
	}
}

func TestLeftGestureBlockedPressNeverCapturesWorld(t *testing.T) {
	r := WorldRenderer{MousePosition: v.Vec2{X: 100, Y: 100}}

	clicked, panning, released := r.updateLeftGesture(0.01, true, true, true)
	if clicked || panning || released {
		t.Fatalf("blocked press = clicked %v, panning %v, released %v; want false, false, false", clicked, panning, released)
	}

	clicked, panning, released = r.updateLeftGesture(0.01, false, false, false)
	if clicked || panning || released {
		t.Fatalf("release = clicked %v, panning %v, released %v; want false, false, false", clicked, panning, released)
	}
}

func TestLeftGestureReleaseOverUIDoesNotClickWorld(t *testing.T) {
	r := WorldRenderer{MousePosition: v.Vec2{X: 100, Y: 100}}

	r.updateLeftGesture(0.01, true, true, false)
	clicked, panning, released := r.updateLeftGesture(0.05, false, false, true)
	if clicked || panning || released {
		t.Fatalf("blocked release = clicked %v, panning %v, released %v; want false, false, false", clicked, panning, released)
	}
}

func TestWheelZoomAppliesIndependentlyOfWorldClickBlocking(t *testing.T) {
	r := WorldRenderer{
		MousePosition: v.Vec2{X: 320, Y: 180},
		TargetZoom:    1,
	}

	r.applyWheelZoom(1)

	if r.TargetZoom <= 1 {
		t.Fatalf("TargetZoom = %f, want greater than 1", r.TargetZoom)
	}
	if r.ZoomAnchor != r.MousePosition {
		t.Fatalf("ZoomAnchor = %v, want mouse position %v", r.ZoomAnchor, r.MousePosition)
	}
}

func TestKeyboardMovementAppliesIndependentlyOfWorldClickBlocking(t *testing.T) {
	r := WorldRenderer{
		Camera: rl.Camera2D{
			Target: rl.Vector2{X: 100, Y: 200},
			Zoom:   1,
		},
		InterpolateFocus: true,
	}

	r.applyKeyboardMovement(v.Vec2{X: 1}, 0.1)

	if r.Camera.Target.X <= 100 || r.Camera.Target.Y != 200 {
		t.Fatalf("camera target = %v, want movement to the right", r.Camera.Target)
	}
	if r.InterpolateFocus {
		t.Fatal("keyboard movement did not take ownership from focus interpolation")
	}
}
