package util

func Clamp(n, low, high int) int {
	if low > high {
		panic("util.Clamp: low is greater than high")
	}
	return max(low, min(n, high))
}

func ClampInt32(n, low, high int32) int32 {
	if low > high {
		panic("util.ClampInt32: low is greater than high")
	}
	return max(low, min(n, high))
}

func ClampByte(n int32) uint8 {
	return uint8(ClampInt32(n, 0, 255))
}

// AbsInt32 returns the absolute value as uint32 because abs(MinInt32)
// cannot be represented by int32.
func AbsInt32(n int32) uint32 {
	if n < 0 {
		return uint32(-int64(n))
	}
	return uint32(n)
}
