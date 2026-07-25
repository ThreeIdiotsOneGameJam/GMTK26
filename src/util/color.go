package util

import (
	"image/color"
)

func SimpleGrayscaleColor(c uint8) *color.RGBA {
	return &color.RGBA{
		R: c,
		G: c,
		B: c,
		A: 255,
	}
}

func ColorAdd(color color.RGBA, n uint8) *color.RGBA {
	color.R = ClampByte(int32(color.R) + int32(n))
	color.G = ClampByte(int32(color.G) + int32(n))
	color.B = ClampByte(int32(color.B) + int32(n))
	return &color
}

func ColorSub(color color.RGBA, n uint8) *color.RGBA {
	color.R = ClampByte(int32(color.R) - int32(n))
	color.G = ClampByte(int32(color.G) - int32(n))
	color.B = ClampByte(int32(color.B) - int32(n))
	return &color
}

func ColorOpacity(c color.RGBA, opacity float32) color.RGBA {
	opacity = max(float32(0), min(opacity, float32(1)))
	c.A = uint8(float32(c.A) * opacity)
	return c
}

type RGB [3]uint8

func (c RGB) toRGBA() color.RGBA {
	return color.RGBA{
		R: c[0],
		G: c[1],
		B: c[2],
		A: 255,
	}
}

func RGBAToRGB(c color.RGBA) RGB {
	return RGB{c.R, c.G, c.B}
}
