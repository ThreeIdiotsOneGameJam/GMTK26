package mathutil

import "image/color"

func ColorAdd(color color.RGBA, n uint8) *color.RGBA {
	color.R = Clampb(int32(color.R) + int32(n))
	color.G = Clampb(int32(color.G) + int32(n))
	color.B = Clampb(int32(color.B) + int32(n))
	return &color
}

func ColorSub(color color.RGBA, n uint8) *color.RGBA {
	color.R = Clampb(int32(color.R) - int32(n))
	color.G = Clampb(int32(color.G) - int32(n))
	color.B = Clampb(int32(color.B) - int32(n))
	return &color
}
