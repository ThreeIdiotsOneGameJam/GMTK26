package global

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/mathutil/vec"
)

var MouseCursorState = rl.MouseCursorDefault
var MousePosition vec.Vec2
var ViewportSize = vec.Vec2i{X: 640, Y: 360}
var ViewportTargetHeight int32 = 360
