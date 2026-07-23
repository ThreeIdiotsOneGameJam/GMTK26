package vec

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Vec2i struct {
	X, Y int32
}

// Component-wise integer operations.
func (v Vec2i) Add(o Vec2i) Vec2i {
	return Vec2i{X: v.X + o.X, Y: v.Y + o.Y}
}

func (v Vec2i) Sub(o Vec2i) Vec2i {
	return Vec2i{X: v.X - o.X, Y: v.Y - o.Y}
}

func (v Vec2i) Mul(o Vec2i) Vec2i {
	return Vec2i{X: v.X * o.X, Y: v.Y * o.Y}
}

// Div performs component-wise integer division.
// It panics with a descriptive message when either divisor component is zero.
func (v Vec2i) Div(o Vec2i) Vec2i {
	result, ok := v.DivChecked(o)
	if !ok {
		panic("vec.Vec2i.Div: division by zero component")
	}
	return result
}

// DivChecked performs component-wise integer division and reports a zero divisor.
func (v Vec2i) DivChecked(o Vec2i) (r Vec2i, ok bool) {
	if o.X == 0 || o.Y == 0 {
		return Vec2i{}, false
	}
	return Vec2i{X: v.X / o.X, Y: v.Y / o.Y}, true
}

// Mod performs component-wise integer remainder.
// It panics with a descriptive message when either divisor component is zero.
func (v Vec2i) Mod(o Vec2i) Vec2i {
	result, ok := v.ModChecked(o)
	if !ok {
		panic("vec.Vec2i.Mod: modulo by zero component")
	}
	return result
}

// ModChecked performs component-wise integer remainder and reports a zero divisor.
func (v Vec2i) ModChecked(o Vec2i) (r Vec2i, ok bool) {
	if o.X == 0 || o.Y == 0 {
		return Vec2i{}, false
	}
	return Vec2i{X: v.X % o.X, Y: v.Y % o.Y}, true
}

// Swizzling.
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

// Conversions.
func (v Vec2i) Vec2() Vec2 {
	return Vec2{X: float32(v.X), Y: float32(v.Y)}
}

func (v Vec2i) String() string {
	return fmt.Sprintf("%d, %d", v.X, v.Y)
}

type Vec2 struct {
	X, Y float32
}

// Component-wise floating-point operations.
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

func (v Vec2) Lerp(o Vec2, t float32) Vec2 {
	return Vec2{
		X: v.X*(1.0-t) + o.X*t,
		Y: v.Y*(1.0-t) + o.Y*t,
	}
}

func (v Vec2) DistanceSqr(o Vec2) float32 {
	return v.Sub(o).MagnitudeSqr()
}

func (v Vec2) Distance(o Vec2) float32 {
	return v.Sub(o).Magnitude()
}

func (v Vec2) MagnitudeSqr() float32 {
	return v.X * v.X + v.Y * v.Y
}

func (v Vec2) Magnitude() float32 {
	return float32(math.Sqrt(float64(v.MagnitudeSqr())))
}

func (v Vec2) Normalize() Vec2 {
	mag := v.Magnitude()
	return Vec2{X: v.X / mag, Y: v.Y / mag}
}

// Trunc rounds each component toward zero.
func (v Vec2) Trunc() Vec2 {
	return Vec2{
		X: float32(math.Trunc(float64(v.X))),
		Y: float32(math.Trunc(float64(v.Y))),
	}
}

func (v Vec2) Floor() Vec2 {
	return Vec2{
		X: float32(math.Floor(float64(v.X))),
		Y: float32(math.Floor(float64(v.Y))),
	}
}

func (v Vec2) Round() Vec2 {
	return Vec2{
		X: float32(math.Round(float64(v.X))),
		Y: float32(math.Round(float64(v.Y))),
	}
}

func (v Vec2) Ceil() Vec2 {
	return Vec2{
		X: float32(math.Ceil(float64(v.X))),
		Y: float32(math.Ceil(float64(v.Y))),
	}
}

// Swizzling.
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

// Conversions.
// Vec2i converts by truncating each component toward zero.
func (v Vec2) Vec2i() Vec2i {
	return v.TruncToInt()
}

func (v Vec2) TruncToInt() Vec2i {
	return Vec2i{
		X: int32(math.Trunc(float64(v.X))),
		Y: int32(math.Trunc(float64(v.Y))),
	}
}

func (v Vec2) FloorToInt() Vec2i {
	return Vec2i{
		X: int32(math.Floor(float64(v.X))),
		Y: int32(math.Floor(float64(v.Y))),
	}
}

func (v Vec2) RoundToInt() Vec2i {
	return Vec2i{
		X: int32(math.Round(float64(v.X))),
		Y: int32(math.Round(float64(v.Y))),
	}
}

func (v Vec2) CeilToInt() Vec2i {
	return Vec2i{
		X: int32(math.Ceil(float64(v.X))),
		Y: int32(math.Ceil(float64(v.Y))),
	}
}

func (v Vec2) ToRL() rl.Vector2 {
	return rl.Vector2{X: v.X, Y: v.Y}
}

func Vec2FromRL(v rl.Vector2) Vec2 {
	return Vec2{X: v.X, Y: v.Y}
}

func (v Vec2) String() string {
	return fmt.Sprintf("%v, %v", v.X, v.Y)
}

type Vec3i struct {
	X, Y, Z int32
}

// Component-wise integer operations.
func (v Vec3i) Add(o Vec3i) Vec3i {
	return Vec3i{X: v.X + o.X, Y: v.Y + o.Y, Z: v.Z + o.Z}
}

func (v Vec3i) Sub(o Vec3i) Vec3i {
	return Vec3i{X: v.X - o.X, Y: v.Y - o.Y, Z: v.Z - o.Z}
}

func (v Vec3i) Mul(o Vec3i) Vec3i {
	return Vec3i{X: v.X * o.X, Y: v.Y * o.Y, Z: v.Z * o.Z}
}

// Div performs component-wise integer division.
// It panics with a descriptive message when any divisor component is zero.
func (v Vec3i) Div(o Vec3i) Vec3i {
	result, ok := v.DivChecked(o)
	if !ok {
		panic("vec.Vec3i.Div: division by zero component")
	}
	return result
}

// DivChecked performs component-wise integer division and reports a zero divisor.
func (v Vec3i) DivChecked(o Vec3i) (r Vec3i, ok bool) {
	if o.X == 0 || o.Y == 0 || o.Z == 0 {
		return Vec3i{}, false
	}
	return Vec3i{X: v.X / o.X, Y: v.Y / o.Y, Z: v.Z / o.Z}, true
}

// Mod performs component-wise integer remainder.
// It panics with a descriptive message when any divisor component is zero.
func (v Vec3i) Mod(o Vec3i) Vec3i {
	result, ok := v.ModChecked(o)
	if !ok {
		panic("vec.Vec3i.Mod: modulo by zero component")
	}
	return result
}

// ModChecked performs component-wise integer remainder and reports a zero divisor.
func (v Vec3i) ModChecked(o Vec3i) (r Vec3i, ok bool) {
	if o.X == 0 || o.Y == 0 || o.Z == 0 {
		return Vec3i{}, false
	}
	return Vec3i{X: v.X % o.X, Y: v.Y % o.Y, Z: v.Z % o.Z}, true
}

// Swizzling.
func (v Vec3i) XX() Vec2i {
	return Vec2i{X: v.X, Y: v.X}
}
func (v Vec3i) XY() Vec2i {
	return Vec2i{X: v.X, Y: v.Y}
}
func (v Vec3i) XZ() Vec2i {
	return Vec2i{X: v.X, Y: v.Z}
}
func (v Vec3i) YX() Vec2i {
	return Vec2i{X: v.Y, Y: v.X}
}
func (v Vec3i) YY() Vec2i {
	return Vec2i{X: v.Y, Y: v.Y}
}
func (v Vec3i) YZ() Vec2i {
	return Vec2i{X: v.Y, Y: v.Z}
}
func (v Vec3i) ZX() Vec2i {
	return Vec2i{X: v.Z, Y: v.X}
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
func (v Vec3i) XXY() Vec3i {
	return Vec3i{X: v.X, Y: v.X, Z: v.Y}
}
func (v Vec3i) XXZ() Vec3i {
	return Vec3i{X: v.X, Y: v.X, Z: v.Z}
}
func (v Vec3i) XYX() Vec3i {
	return Vec3i{X: v.X, Y: v.Y, Z: v.X}
}
func (v Vec3i) XYY() Vec3i {
	return Vec3i{X: v.X, Y: v.Y, Z: v.Y}
}
func (v Vec3i) XYZ() Vec3i {
	return Vec3i{X: v.X, Y: v.Y, Z: v.Z}
}
func (v Vec3i) XZX() Vec3i {
	return Vec3i{X: v.X, Y: v.Z, Z: v.X}
}
func (v Vec3i) XZY() Vec3i {
	return Vec3i{X: v.X, Y: v.Z, Z: v.Y}
}
func (v Vec3i) XZZ() Vec3i {
	return Vec3i{X: v.X, Y: v.Z, Z: v.Z}
}
func (v Vec3i) YXX() Vec3i {
	return Vec3i{X: v.Y, Y: v.X, Z: v.X}
}
func (v Vec3i) YXY() Vec3i {
	return Vec3i{X: v.Y, Y: v.X, Z: v.Y}
}
func (v Vec3i) YXZ() Vec3i {
	return Vec3i{X: v.Y, Y: v.X, Z: v.Z}
}
func (v Vec3i) YYX() Vec3i {
	return Vec3i{X: v.Y, Y: v.Y, Z: v.X}
}
func (v Vec3i) YYY() Vec3i {
	return Vec3i{X: v.Y, Y: v.Y, Z: v.Y}
}
func (v Vec3i) YYZ() Vec3i {
	return Vec3i{X: v.Y, Y: v.Y, Z: v.Z}
}
func (v Vec3i) YZX() Vec3i {
	return Vec3i{X: v.Y, Y: v.Z, Z: v.X}
}
func (v Vec3i) YZY() Vec3i {
	return Vec3i{X: v.Y, Y: v.Z, Z: v.Y}
}
func (v Vec3i) YZZ() Vec3i {
	return Vec3i{X: v.Y, Y: v.Z, Z: v.Z}
}
func (v Vec3i) ZXX() Vec3i {
	return Vec3i{X: v.Z, Y: v.X, Z: v.X}
}
func (v Vec3i) ZXY() Vec3i {
	return Vec3i{X: v.Z, Y: v.X, Z: v.Y}
}
func (v Vec3i) ZXZ() Vec3i {
	return Vec3i{X: v.Z, Y: v.X, Z: v.Z}
}
func (v Vec3i) ZYX() Vec3i {
	return Vec3i{X: v.Z, Y: v.Y, Z: v.X}
}
func (v Vec3i) ZYY() Vec3i {
	return Vec3i{X: v.Z, Y: v.Y, Z: v.Y}
}
func (v Vec3i) ZYZ() Vec3i {
	return Vec3i{X: v.Z, Y: v.Y, Z: v.Z}
}
func (v Vec3i) ZZX() Vec3i {
	return Vec3i{X: v.Z, Y: v.Z, Z: v.X}
}
func (v Vec3i) ZZY() Vec3i {
	return Vec3i{X: v.Z, Y: v.Z, Z: v.Y}
}
func (v Vec3i) ZZZ() Vec3i {
	return Vec3i{X: v.Z, Y: v.Z, Z: v.Z}
}

// Conversions.
func (v Vec3i) Vec3() Vec3 {
	return Vec3{X: float32(v.X), Y: float32(v.Y), Z: float32(v.Z)}
}

func (v Vec3i) String() string {
	return fmt.Sprintf("%d, %d, %d", v.X, v.Y, v.Z)
}

type Vec3 struct {
	X, Y, Z float32
}

// Component-wise floating-point operations.
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

// Trunc rounds each component toward zero.
func (v Vec3) Trunc() Vec3 {
	return Vec3{
		X: float32(math.Trunc(float64(v.X))),
		Y: float32(math.Trunc(float64(v.Y))),
		Z: float32(math.Trunc(float64(v.Z))),
	}
}

func (v Vec3) Floor() Vec3 {
	return Vec3{
		X: float32(math.Floor(float64(v.X))),
		Y: float32(math.Floor(float64(v.Y))),
		Z: float32(math.Floor(float64(v.Z))),
	}
}

func (v Vec3) Round() Vec3 {
	return Vec3{
		X: float32(math.Round(float64(v.X))),
		Y: float32(math.Round(float64(v.Y))),
		Z: float32(math.Round(float64(v.Z))),
	}
}

func (v Vec3) Ceil() Vec3 {
	return Vec3{
		X: float32(math.Ceil(float64(v.X))),
		Y: float32(math.Ceil(float64(v.Y))),
		Z: float32(math.Ceil(float64(v.Z))),
	}
}

// Swizzling.
func (v Vec3) XX() Vec2 {
	return Vec2{X: v.X, Y: v.X}
}
func (v Vec3) XY() Vec2 {
	return Vec2{X: v.X, Y: v.Y}
}
func (v Vec3) XZ() Vec2 {
	return Vec2{X: v.X, Y: v.Z}
}
func (v Vec3) YX() Vec2 {
	return Vec2{X: v.Y, Y: v.X}
}
func (v Vec3) YY() Vec2 {
	return Vec2{X: v.Y, Y: v.Y}
}
func (v Vec3) YZ() Vec2 {
	return Vec2{X: v.Y, Y: v.Z}
}
func (v Vec3) ZX() Vec2 {
	return Vec2{X: v.Z, Y: v.X}
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
func (v Vec3) XXY() Vec3 {
	return Vec3{X: v.X, Y: v.X, Z: v.Y}
}
func (v Vec3) XXZ() Vec3 {
	return Vec3{X: v.X, Y: v.X, Z: v.Z}
}
func (v Vec3) XYX() Vec3 {
	return Vec3{X: v.X, Y: v.Y, Z: v.X}
}
func (v Vec3) XYY() Vec3 {
	return Vec3{X: v.X, Y: v.Y, Z: v.Y}
}
func (v Vec3) XYZ() Vec3 {
	return Vec3{X: v.X, Y: v.Y, Z: v.Z}
}
func (v Vec3) XZX() Vec3 {
	return Vec3{X: v.X, Y: v.Z, Z: v.X}
}
func (v Vec3) XZY() Vec3 {
	return Vec3{X: v.X, Y: v.Z, Z: v.Y}
}
func (v Vec3) XZZ() Vec3 {
	return Vec3{X: v.X, Y: v.Z, Z: v.Z}
}
func (v Vec3) YXX() Vec3 {
	return Vec3{X: v.Y, Y: v.X, Z: v.X}
}
func (v Vec3) YXY() Vec3 {
	return Vec3{X: v.Y, Y: v.X, Z: v.Y}
}
func (v Vec3) YXZ() Vec3 {
	return Vec3{X: v.Y, Y: v.X, Z: v.Z}
}
func (v Vec3) YYX() Vec3 {
	return Vec3{X: v.Y, Y: v.Y, Z: v.X}
}
func (v Vec3) YYY() Vec3 {
	return Vec3{X: v.Y, Y: v.Y, Z: v.Y}
}
func (v Vec3) YYZ() Vec3 {
	return Vec3{X: v.Y, Y: v.Y, Z: v.Z}
}
func (v Vec3) YZX() Vec3 {
	return Vec3{X: v.Y, Y: v.Z, Z: v.X}
}
func (v Vec3) YZY() Vec3 {
	return Vec3{X: v.Y, Y: v.Z, Z: v.Y}
}
func (v Vec3) YZZ() Vec3 {
	return Vec3{X: v.Y, Y: v.Z, Z: v.Z}
}
func (v Vec3) ZXX() Vec3 {
	return Vec3{X: v.Z, Y: v.X, Z: v.X}
}
func (v Vec3) ZXY() Vec3 {
	return Vec3{X: v.Z, Y: v.X, Z: v.Y}
}
func (v Vec3) ZXZ() Vec3 {
	return Vec3{X: v.Z, Y: v.X, Z: v.Z}
}
func (v Vec3) ZYX() Vec3 {
	return Vec3{X: v.Z, Y: v.Y, Z: v.X}
}
func (v Vec3) ZYY() Vec3 {
	return Vec3{X: v.Z, Y: v.Y, Z: v.Y}
}
func (v Vec3) ZYZ() Vec3 {
	return Vec3{X: v.Z, Y: v.Y, Z: v.Z}
}
func (v Vec3) ZZX() Vec3 {
	return Vec3{X: v.Z, Y: v.Z, Z: v.X}
}
func (v Vec3) ZZY() Vec3 {
	return Vec3{X: v.Z, Y: v.Z, Z: v.Y}
}
func (v Vec3) ZZZ() Vec3 {
	return Vec3{X: v.Z, Y: v.Z, Z: v.Z}
}

// Conversions.
// Vec3i converts by truncating each component toward zero.
func (v Vec3) Vec3i() Vec3i {
	return v.TruncToInt()
}

func (v Vec3) TruncToInt() Vec3i {
	return Vec3i{
		X: int32(math.Trunc(float64(v.X))),
		Y: int32(math.Trunc(float64(v.Y))),
		Z: int32(math.Trunc(float64(v.Z))),
	}
}

func (v Vec3) FloorToInt() Vec3i {
	return Vec3i{
		X: int32(math.Floor(float64(v.X))),
		Y: int32(math.Floor(float64(v.Y))),
		Z: int32(math.Floor(float64(v.Z))),
	}
}

func (v Vec3) RoundToInt() Vec3i {
	return Vec3i{
		X: int32(math.Round(float64(v.X))),
		Y: int32(math.Round(float64(v.Y))),
		Z: int32(math.Round(float64(v.Z))),
	}
}

func (v Vec3) CeilToInt() Vec3i {
	return Vec3i{
		X: int32(math.Ceil(float64(v.X))),
		Y: int32(math.Ceil(float64(v.Y))),
		Z: int32(math.Ceil(float64(v.Z))),
	}
}

func (v Vec3) String() string {
	return fmt.Sprintf("%v, %v, %v", v.X, v.Y, v.Z)
}
