package mathutil

import "github.com/threeidiotsonegamejam/gmtk26/src/util"

func Mini(a, b int32) int32 {
	return util.IfElse(a < b, a, b)
}

func Maxi(a, b int32) int32 {
	return util.IfElse(a > b, a, b)
}

func Clampi(n, min, max int32) int32 {
	return Maxi(Mini(n, max), min)
}

func Clampb(n int32) uint8 {
	return uint8(Maxi(Mini(n, 255), 0))
}
