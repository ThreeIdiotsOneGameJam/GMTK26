package ui

import (
	"image/color"
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func assertDefaultDropShadow(
	t *testing.T,
	name string,
	shadowColor *color.RGBA,
	shadowOffset vec.Vec2i,
) {
	t.Helper()

	if shadowColor == nil {
		t.Fatalf("%s shadow color is nil", name)
	}
	if *shadowColor != (color.RGBA{R: 0, G: 0, B: 0, A: 160}) {
		t.Errorf("%s shadow color = %#v, want black with alpha 160", name, *shadowColor)
	}
	if shadowOffset != (vec.Vec2i{X: 4, Y: 4}) {
		t.Errorf("%s shadow offset = %#v, want {X:4 Y:4}", name, shadowOffset)
	}
}

func TestElementsHaveDefaultDropShadow(t *testing.T) {
	button := Button()
	input := Input()
	slider := Slider()
	toggle := Toggle()
	text := Text().WithTextSize(64)

	assertDefaultDropShadow(t, "button", button.ShadowColor, button.ShadowOffset)
	assertDefaultDropShadow(t, "input", input.ShadowColor, input.ShadowOffset)
	assertDefaultDropShadow(t, "slider", slider.ShadowColor, slider.ShadowOffset)
	assertDefaultDropShadow(t, "toggle", toggle.ShadowColor, toggle.ShadowOffset)
	assertDefaultDropShadow(t, "text", text.ShadowColor, text.resolvedShadowOffset())
}

func TestElementDropShadowCanBeCustomized(t *testing.T) {
	customColor := color.RGBA{R: 12, G: 34, B: 56, A: 78}
	customOffset := vec.Vec2i{X: -3, Y: 7}

	button := Button().WithShadow(customColor, customOffset)
	input := Input().WithShadow(customColor, customOffset)
	slider := Slider().WithShadow(customColor, customOffset)
	toggle := Toggle().WithShadow(customColor, customOffset)
	text := Text().WithTextShadow(customColor, customOffset).WithTextSize(96)

	for _, shadow := range []struct {
		name   string
		color  *color.RGBA
		offset vec.Vec2i
	}{
		{name: "button", color: button.ShadowColor, offset: button.ShadowOffset},
		{name: "input", color: input.ShadowColor, offset: input.ShadowOffset},
		{name: "slider", color: slider.ShadowColor, offset: slider.ShadowOffset},
		{name: "toggle", color: toggle.ShadowColor, offset: toggle.ShadowOffset},
		{name: "text", color: text.ShadowColor, offset: text.resolvedShadowOffset()},
	} {
		if shadow.color == nil || *shadow.color != customColor {
			t.Errorf("%s shadow color = %#v, want %#v", shadow.name, shadow.color, customColor)
		}
		if shadow.offset != customOffset {
			t.Errorf("%s shadow offset = %#v, want %#v", shadow.name, shadow.offset, customOffset)
		}
	}

	customColor.R = 255
	if *button.ShadowColor != (color.RGBA{R: 12, G: 34, B: 56, A: 78}) {
		t.Errorf("button shadow color changed with caller's copy: %#v", *button.ShadowColor)
	}
}

func TestDefaultTextShadowOffsetScalesWithFontSize(t *testing.T) {
	for _, test := range []struct {
		textSize int32
		want     int32
	}{
		{textSize: 0, want: 2},
		{textSize: 18, want: 2},
		{textSize: 36, want: 2},
		{textSize: 48, want: 3},
		{textSize: 64, want: 4},
		{textSize: 96, want: 6},
	} {
		text := Text().WithTextSize(test.textSize)
		want := vec.Vec2i{X: test.want, Y: test.want}

		if got := text.resolvedShadowOffset(); got != want {
			t.Errorf("text size %d shadow offset = %#v, want %#v", test.textSize, got, want)
		}
	}
}

func TestElementDropShadowCanBeDisabled(t *testing.T) {
	button := Button().WithoutShadow()
	input := Input().WithoutShadow()
	slider := Slider().WithoutShadow()
	toggle := Toggle().WithoutShadow()
	text := Text().WithoutShadow()

	for name, shadowColor := range map[string]*color.RGBA{
		"button": button.ShadowColor,
		"input":  input.ShadowColor,
		"slider": slider.ShadowColor,
		"toggle": toggle.ShadowColor,
		"text":   text.ShadowColor,
	} {
		if shadowColor != nil {
			t.Errorf("%s shadow color = %#v, want nil", name, shadowColor)
		}
	}
}
