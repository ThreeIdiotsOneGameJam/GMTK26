package math

type Vec2i struct {
	X, Y int32
}
// Operations
func (v Vec2i) Add(o Vec2i) Vec2i {
	return Vec2i{X: v.X + o.X, Y: v.Y + o.Y}
}
func (v Vec2i) Sub(o Vec2i) Vec2i {
	return Vec2i{X: v.X - o.X, Y: v.Y - o.Y}
}
func (v Vec2i) Mul(o Vec2i) Vec2i {
	return Vec2i{X: v.X * o.X, Y: v.Y * o.Y}
}
func (v Vec2i) Div(o Vec2i) Vec2i {
	return Vec2i{X: v.X / o.X, Y: v.Y / o.Y}
}
func (v Vec2i) Mod(o Vec2i) Vec2i {
	return Vec2i{X: v.X % o.X, Y: v.Y % o.Y}
}
// Swizzling
func (v Vec2i) XX() Vec2i {
	return Vec2i{X: v.X, Y: v.X}
}
func (v Vec2i) XY() Vec2i {
	return Vec2i{X: v.X, Y: v.Y}
}
func (v Vec2i) YX() Vec2i {
	return Vec2i{X: v.Y, Y: v.X}
}
func (v Vec2i) YY() Vec2i {
	return Vec2i{X: v.Y, Y: v.Y}
}
// Convertions
func (v Vec2i) Vec2() Vec2 {
	return Vec2{X: float32(v.X), Y: float32(v.Y)}
}

type Vec2 struct {
	X, Y float32
}
// Operations
func (v Vec2) Add(o Vec2) Vec2 {
	return Vec2{X: v.X + o.X, Y: v.Y + o.Y}
}
func (v Vec2) Sub(o Vec2) Vec2 {
	return Vec2{X: v.X - o.X, Y: v.Y - o.Y}
}
func (v Vec2) Mul(o Vec2) Vec2 {
	return Vec2{X: v.X * o.X, Y: v.Y * o.Y}
}
func (v Vec2) Div(o Vec2) Vec2 {
	return Vec2{X: v.X / o.X, Y: v.Y / o.Y}
}
func (v Vec2) Floor() Vec2 {
	return Vec2{
		X: float32(int32(v.X)),
		Y: float32(int32(v.Y)),
	}
}
func (v Vec2) Round() Vec2 {
	return Vec2{
		X: float32(int32(v.X + 0.5)),
		Y: float32(int32(v.Y + 0.5)),
	}
}
func (v Vec2) Ceil() Vec2 {
	return Vec2{
		X: float32(int32(v.X)+1),
		Y: float32(int32(v.Y)+1),
	}
}
// Swizzling
func (v Vec2) XX() Vec2 {
	return Vec2{X: v.X, Y: v.X}
}
func (v Vec2) XY() Vec2 {
	return Vec2{X: v.X, Y: v.Y}
}
func (v Vec2) YX() Vec2 {
	return Vec2{X: v.Y, Y: v.X}
}
func (v Vec2) YY() Vec2 {
	return Vec2{X: v.Y, Y: v.Y}
}
// Convertions
func (v Vec2) Vec2i() Vec2i {
	return Vec2i{X: int32(v.X), Y: int32(v.Y)}
}

type Vec3i struct {
	X, Y, Z int32
}

type Vec3 struct {
	X, Y, Z float32
}
