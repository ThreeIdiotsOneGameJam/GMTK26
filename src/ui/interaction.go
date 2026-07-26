package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type elementRect struct {
	X, Y, Width, Height int32
}

var keyboardInputClaimed bool

// BeginInputFrame resets transient ownership before the active UI tree updates.
func BeginInputFrame() {
	keyboardInputClaimed = false
}

func claimKeyboardInput() bool {
	if keyboardInputClaimed {
		return false
	}
	keyboardInputClaimed = true
	return true
}

func (rect elementRect) contains(point vec.Vec2i) bool {
	return point.X >= rect.X &&
		point.X <= rect.X+rect.Width &&
		point.Y >= rect.Y &&
		point.Y <= rect.Y+rect.Height
}

func (rect elementRect) containsStrict(point vec.Vec2i) bool {
	return point.X > rect.X &&
		point.X < rect.X+rect.Width &&
		point.Y > rect.Y &&
		point.Y < rect.Y+rect.Height
}

func mousePosition() vec.Vec2i {
	return vec.Vec2i{
		X: int32(global.MousePosition.X),
		Y: int32(global.MousePosition.Y),
	}
}

func claimPointer(cursor int32) {
	global.MouseCursorState = cursor
	global.UIBlocksWorldInput = true
}

func controlState(enabled, hovered, active bool) UIState {
	switch {
	case !enabled:
		return StateDisabled
	case active:
		return StateClick
	case hovered:
		return StateHover
	default:
		return StateDefault
	}
}

func drawOutlinedRectangle(rect elementRect, outlineWidth int32, outline, fill rl.Color) {
	rl.DrawRectangle(
		rect.X-outlineWidth,
		rect.Y-outlineWidth,
		rect.Width+outlineWidth*2,
		rect.Height+outlineWidth*2,
		outline,
	)
	rl.DrawRectangle(rect.X, rect.Y, rect.Width, rect.Height, fill)
}
