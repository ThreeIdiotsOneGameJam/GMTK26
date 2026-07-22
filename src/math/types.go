package math

type Vec2i struct {
	X, Y int32
}

type Vec2 struct {
	X, Y float32
}

type Vec3i struct {
	X, Y, Z int32
}

type Vec3 struct {
	X, Y, Z float32
}

func (v *Vec2i) Add(o Vec2i) {
	v.X += o.X
	v.Y += o.Y
}

func (v *Vec2i) Sub(o Vec2i) {
	v.X -= o.X
	v.Y -= o.Y
}
