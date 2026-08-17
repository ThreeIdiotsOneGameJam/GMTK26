package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/render/shaders"
)

func Vignette() *VignetteElement {
	el := &VignetteElement{}
	el.BaseElement = NewBaseElement(el)

	el.Color = color.RGBA{R: 10, G: 0, B: 5, A: 80}
	el.Radius = 0.6

	return el
}

type VignetteElement struct {
	BaseElement[*VignetteElement]

	Color  color.RGBA
	Radius float32
}

func (el *VignetteElement) draw() {
	if !rl.IsShaderValid(shaders.Vignette) {
		return
	}

	s := el.Parent.Size().Vec2()
	p := el.Parent.AbsolutePos().Vec2()

	colorLoc := rl.GetShaderLocation(shaders.Vignette, "vignetteColor")
	radiusLoc := rl.GetShaderLocation(shaders.Vignette, "vignetteRadius")
	vignetteColor := []float32{
		float32(el.Color.R) / 255,
		float32(el.Color.G) / 255,
		float32(el.Color.B) / 255,
		float32(el.Color.A) / 255,
	}
	rl.SetShaderValue(shaders.Vignette, colorLoc, vignetteColor, rl.ShaderUniformVec4)
	rl.SetShaderValue(
		shaders.Vignette,
		radiusLoc,
		[]float32{el.Radius},
		rl.ShaderUniformFloat,
	)

	// The shader writes premultiplied color so the vignette has the same blend
	// result on the desktop framebuffer and every WebGL canvas implementation.
	rl.BeginBlendMode(rl.BlendAlphaPremultiply)
	drawShaderQuad(shaders.Vignette, p, s, nil)
	rl.EndBlendMode()
}

func (el *VignetteElement) WithColor(col color.RGBA) *VignetteElement {
	el.Color = col
	return el
}
func (el *VignetteElement) WithAlpha(alpha uint8) *VignetteElement {
	el.Color.A = alpha
	return el
}

func (el *VignetteElement) WithRadius(radius float32) *VignetteElement {
	el.Radius = radius
	return el
}
