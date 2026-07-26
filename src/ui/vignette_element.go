package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/render/shaders"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/rlutil"
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

	// The shader writes premultiplied color so the vignette has the same blend
	// result on the desktop framebuffer and every WebGL canvas implementation.
	rl.BeginBlendMode(rl.BlendAlphaPremultiply)
	rl.BeginShaderMode(shaders.Vignette)
	rl.Begin(rl.Triangles)

	rlutil.Color4ub(el.Color.R, el.Color.G, el.Color.B, el.Color.A)
	rl.Normal3f(el.Radius, 0.0, 1.0)

	rl.TexCoord2f(0.0, 0.0)
	rlutil.Vertex2f(p.X, p.Y)

	rl.TexCoord2f(1.0, 1.0)
	rlutil.Vertex2f(p.X+s.X, p.Y+s.Y)

	rl.TexCoord2f(1.0, 0.0)
	rlutil.Vertex2f(p.X+s.X, p.Y)

	rl.TexCoord2f(0.0, 1.0)
	rlutil.Vertex2f(p.X, p.Y+s.Y)

	rl.TexCoord2f(1.0, 1.0)
	rlutil.Vertex2f(p.X+s.X, p.Y+s.Y)

	rl.TexCoord2f(0.0, 0.0)
	rlutil.Vertex2f(p.X, p.Y)

	rl.End()
	rl.EndShaderMode()
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
