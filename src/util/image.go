package util

import rl "github.com/gen2brain/raylib-go/raylib"

func SampleImageRepeat(image rl.Image, x, y int32) rl.Color {
	wx := x
	for wx < 0 {
		wx += image.Width
	}
	for wx >= image.Width {
		wx -= image.Width
	}
	wy := y
	for wy < 0 {
		wy += image.Width
	}
	for wy >= image.Width {
		wy -= image.Width
	}

	return rl.GetImageColor(image, wx, wy)
}
