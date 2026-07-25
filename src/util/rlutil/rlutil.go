package rlutil

import rl "github.com/gen2brain/raylib-go/raylib"

func Vertex2f(x, y float32) {
	rl.Vertex3f(x, y, 0)
}

func Color4ub(r, g, b, a uint8) {
	rl.Color4f(float32(r)/255, float32(g)/255, float32(b)/255, float32(a)/255)
}
