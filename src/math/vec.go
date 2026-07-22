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
		X: float32(int32(v.X) + 1),
		Y: float32(int32(v.Y) + 1),
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

// Operations
func (v Vec3i) Add(o Vec3i) Vec3i {
	return Vec3i{X: v.X + o.X, Y: v.Y + o.Y, Z: v.Z + o.Z}
}
func (v Vec3i) Sub(o Vec3i) Vec3i {
	return Vec3i{X: v.X - o.X, Y: v.Y - o.Y, Z: v.Z - o.Z}
}
func (v Vec3i) Mul(o Vec3i) Vec3i {
	return Vec3i{X: v.X * o.X, Y: v.Y * o.Y, Z: v.Z * o.Z}
}
func (v Vec3i) Div(o Vec3i) Vec3i {
	return Vec3i{X: v.X / o.X, Y: v.Y / o.Y, Z: v.Z / o.Z}
}
func (v Vec3i) Mod(o Vec3i) Vec3i {
	return Vec3i{X: v.X % o.X, Y: v.Y % o.Y, Z: v.Z % o.Z}
}

// Swizzling
func (v Vec3i) XX() Vec2i {
	return Vec2i{X: v.X, Y: v.X}
}
func (v Vec3i) XY() Vec2i {
	return Vec2i{X: v.X, Y: v.Y}
}
func (v Vec3i) YX() Vec2i {
	return Vec2i{X: v.Y, Y: v.X}
}
func (v Vec3i) YY() Vec2i {
	return Vec2i{X: v.Y, Y: v.Y}
}
func (v Vec3i) XZ() Vec2i {
	return Vec2i{X: v.X, Y: v.Z}
}
func (v Vec3i) ZX() Vec2i {
	return Vec2i{X: v.Z, Y: v.X}
}
func (v Vec3i) YZ() Vec2i {
	return Vec2i{X: v.Y, Y: v.Z}
}
func (v Vec3i) ZY() Vec2i {
	return Vec2i{X: v.Z, Y: v.Y}
}
func (v Vec3i) ZZ() Vec2i {
	return Vec2i{X: v.Z, Y: v.Z}
}
func (v Vec3i) XXX() Vec3i {
	return Vec3i{X: v.X, Y: v.X, Z: v.X}
}
func (v Vec3i) XYX() Vec3i {
	return Vec3i{X: v.X, Y: v.Y, Z: v.X}
}
func (v Vec3i) YXX() Vec3i {
	return Vec3i{X: v.Y, Y: v.X, Z: v.X}
}
func (v Vec3i) YYX() Vec3i {
	return Vec3i{X: v.Y, Y: v.Y, Z: v.X}
}
func (v Vec3i) XZX() Vec3i {
	return Vec3i{X: v.X, Y: v.Z, Z: v.X}
}
func (v Vec3i) ZXX() Vec3i {
	return Vec3i{X: v.Z, Y: v.X, Z: v.X}
}
func (v Vec3i) YZX() Vec3i {
	return Vec3i{X: v.Y, Y: v.Z, Z: v.X}
}
func (v Vec3i) ZYX() Vec3i {
	return Vec3i{X: v.Z, Y: v.Y, Z: v.X}
}
func (v Vec3i) ZZX() Vec3i {
	return Vec3i{X: v.Z, Y: v.Z, Z: v.X}
}
func (v Vec3i) XXY() Vec3i {
	return Vec3i{X: v.X, Y: v.X, Z: v.Y}
}
func (v Vec3i) XYY() Vec3i {
	return Vec3i{X: v.X, Y: v.Y, Z: v.Y}
}
func (v Vec3i) YXY() Vec3i {
	return Vec3i{X: v.Y, Y: v.X, Z: v.Y}
}
func (v Vec3i) YYY() Vec3i {
	return Vec3i{X: v.Y, Y: v.Y, Z: v.Y}
}
func (v Vec3i) XZY() Vec3i {
	return Vec3i{X: v.X, Y: v.Z, Z: v.Y}
}
func (v Vec3i) ZXY() Vec3i {
	return Vec3i{X: v.Z, Y: v.X, Z: v.Y}
}
func (v Vec3i) YZY() Vec3i {
	return Vec3i{X: v.Y, Y: v.Z, Z: v.Y}
}
func (v Vec3i) ZYY() Vec3i {
	return Vec3i{X: v.Z, Y: v.Y, Z: v.Y}
}
func (v Vec3i) ZZY() Vec3i {
	return Vec3i{X: v.Z, Y: v.Z, Z: v.Y}
}
func (v Vec3i) XXZ() Vec3i {
	return Vec3i{X: v.X, Y: v.X, Z: v.Z}
}
func (v Vec3i) XYZ() Vec3i {
	return Vec3i{X: v.X, Y: v.Y, Z: v.Z}
}
func (v Vec3i) YXZ() Vec3i {
	return Vec3i{X: v.Y, Y: v.X, Z: v.Z}
}
func (v Vec3i) YYZ() Vec3i {
	return Vec3i{X: v.Y, Y: v.Y, Z: v.Z}
}
func (v Vec3i) XZZ() Vec3i {
	return Vec3i{X: v.X, Y: v.Z, Z: v.Z}
}
func (v Vec3i) ZXZ() Vec3i {
	return Vec3i{X: v.Z, Y: v.X, Z: v.Z}
}
func (v Vec3i) YZZ() Vec3i {
	return Vec3i{X: v.Y, Y: v.Z, Z: v.Z}
}
func (v Vec3i) ZYZ() Vec3i {
	return Vec3i{X: v.Z, Y: v.Y, Z: v.Z}
}
func (v Vec3i) ZZZ() Vec3i {
	return Vec3i{X: v.Z, Y: v.Z, Z: v.Z}
}

// Convertions
func (v Vec3i) Vec3() Vec3 {
	return Vec3{X: float32(v.X), Y: float32(v.Y), Z: float32(v.Z)}
}

type Vec3 struct {
	X, Y, Z float32
}

// Operations
func (v Vec3) Add(o Vec3) Vec3 {
	return Vec3{X: v.X + o.X, Y: v.Y + o.Y, Z: v.Z + o.Z}
}
func (v Vec3) Sub(o Vec3) Vec3 {
	return Vec3{X: v.X - o.X, Y: v.Y - o.Y, Z: v.Z - o.Z}
}
func (v Vec3) Mul(o Vec3) Vec3 {
	return Vec3{X: v.X * o.X, Y: v.Y * o.Y, Z: v.Z * o.Z}
}
func (v Vec3) Div(o Vec3) Vec3 {
	return Vec3{X: v.X / o.X, Y: v.Y / o.Y, Z: v.Z / o.Z}
}
func (v Vec3) Floor() Vec3 {
	return Vec3{
		X: float32(int32(v.X)),
		Y: float32(int32(v.Y)),
		Z: float32(int32(v.Z)),
	}
}
func (v Vec3) Round() Vec3 {
	return Vec3{
		X: float32(int32(v.X + 0.5)),
		Y: float32(int32(v.Y + 0.5)),
		Z: float32(int32(v.Z + 0.5)),
	}
}
func (v Vec3) Ceil() Vec3 {
	return Vec3{
		X: float32(int32(v.X) + 1),
		Y: float32(int32(v.Y) + 1),
		Z: float32(int32(v.Z) + 1),
	}
}

// Swizzling
func (v Vec3) XX() Vec2 {
	return Vec2{X: v.X, Y: v.X}
}
func (v Vec3) XY() Vec2 {
	return Vec2{X: v.X, Y: v.Y}
}
func (v Vec3) YX() Vec2 {
	return Vec2{X: v.Y, Y: v.X}
}
func (v Vec3) YY() Vec2 {
	return Vec2{X: v.Y, Y: v.Y}
}
func (v Vec3) XZ() Vec2 {
	return Vec2{X: v.X, Y: v.Z}
}
func (v Vec3) ZX() Vec2 {
	return Vec2{X: v.Z, Y: v.X}
}
func (v Vec3) YZ() Vec2 {
	return Vec2{X: v.Y, Y: v.Z}
}
func (v Vec3) ZY() Vec2 {
	return Vec2{X: v.Z, Y: v.Y}
}
func (v Vec3) ZZ() Vec2 {
	return Vec2{X: v.Z, Y: v.Z}
}
func (v Vec3) XXX() Vec3 {
	return Vec3{X: v.X, Y: v.X, Z: v.X}
}
func (v Vec3) XYX() Vec3 {
	return Vec3{X: v.X, Y: v.Y, Z: v.X}
}
func (v Vec3) YXX() Vec3 {
	return Vec3{X: v.Y, Y: v.X, Z: v.X}
}
func (v Vec3) YYX() Vec3 {
	return Vec3{X: v.Y, Y: v.Y, Z: v.X}
}
func (v Vec3) XZX() Vec3 {
	return Vec3{X: v.X, Y: v.Z, Z: v.X}
}
func (v Vec3) ZXX() Vec3 {
	return Vec3{X: v.Z, Y: v.X, Z: v.X}
}
func (v Vec3) YZX() Vec3 {
	return Vec3{X: v.Y, Y: v.Z, Z: v.X}
}
func (v Vec3) ZYX() Vec3 {
	return Vec3{X: v.Z, Y: v.Y, Z: v.X}
}
func (v Vec3) ZZX() Vec3 {
	return Vec3{X: v.Z, Y: v.Z, Z: v.X}
}
func (v Vec3) XXY() Vec3 {
	return Vec3{X: v.X, Y: v.X, Z: v.Y}
}
func (v Vec3) XYY() Vec3 {
	return Vec3{X: v.X, Y: v.Y, Z: v.Y}
}
func (v Vec3) YXY() Vec3 {
	return Vec3{X: v.Y, Y: v.X, Z: v.Y}
}
func (v Vec3) YYY() Vec3 {
	return Vec3{X: v.Y, Y: v.Y, Z: v.Y}
}
func (v Vec3) XZY() Vec3 {
	return Vec3{X: v.X, Y: v.Z, Z: v.Y}
}
func (v Vec3) ZXY() Vec3 {
	return Vec3{X: v.Z, Y: v.X, Z: v.Y}
}
func (v Vec3) YZY() Vec3 {
	return Vec3{X: v.Y, Y: v.Z, Z: v.Y}
}
func (v Vec3) ZYY() Vec3 {
	return Vec3{X: v.Z, Y: v.Y, Z: v.Y}
}
func (v Vec3) ZZY() Vec3 {
	return Vec3{X: v.Z, Y: v.Z, Z: v.Y}
}
func (v Vec3) XXZ() Vec3 {
	return Vec3{X: v.X, Y: v.X, Z: v.Z}
}
func (v Vec3) XYZ() Vec3 {
	return Vec3{X: v.X, Y: v.Y, Z: v.Z}
}
func (v Vec3) YXZ() Vec3 {
	return Vec3{X: v.Y, Y: v.X, Z: v.Z}
}
func (v Vec3) YYZ() Vec3 {
	return Vec3{X: v.Y, Y: v.Y, Z: v.Z}
}
func (v Vec3) XZZ() Vec3 {
	return Vec3{X: v.X, Y: v.Z, Z: v.Z}
}
func (v Vec3) ZXZ() Vec3 {
	return Vec3{X: v.Z, Y: v.X, Z: v.Z}
}
func (v Vec3) YZZ() Vec3 {
	return Vec3{X: v.Y, Y: v.Z, Z: v.Z}
}
func (v Vec3) ZYZ() Vec3 {
	return Vec3{X: v.Z, Y: v.Y, Z: v.Z}
}
func (v Vec3) ZZZ() Vec3 {
	return Vec3{X: v.Z, Y: v.Z, Z: v.Z}
}

// Convertions
func (v Vec3) Vec3i() Vec3i {
	return Vec3i{X: int32(v.X), Y: int32(v.Y), Z: int32(v.Z)}
}
