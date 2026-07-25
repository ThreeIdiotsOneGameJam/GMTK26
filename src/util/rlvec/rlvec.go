package rlvec

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func ToRL(v vec.Vec2) rl.Vector2 {
	return rl.Vector2{X: v.X, Y: v.Y}
}

func FromRL(v rl.Vector2) vec.Vec2 {
	return vec.Vec2{X: v.X, Y: v.Y}
}
