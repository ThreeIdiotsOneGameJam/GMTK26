package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Vignette() *VignetteElement {
	el := &VignetteElement{}
	el.BaseElement = NewBaseElement(el)

	el.Color = color.RGBA{R: 10, G: 0, B: 5, A: 80}
	el.Radius = 0.6

	return el
}

var vignetteShader rl.Shader

type VignetteElement struct {
	BaseElement[*VignetteElement]

	Color  color.RGBA
	Radius float32
}

func (el *VignetteElement) prepare() {
	if !rl.IsShaderValid(vignetteShader) {
		vignetteShader = rl.LoadShader("assets/shaders/base.vert", "assets/shaders/vignette.frag")
	}
}

func (el *VignetteElement) draw() {
	if !rl.IsShaderValid(vignetteShader) {
		return
	}

	s := el.Parent.Size().Vec2()
	p := el.Parent.AbsolutePos().Vec2()

	rl.BeginShaderMode(vignetteShader)
	rl.Begin(rl.Triangles)

	rl.Color4ub(el.Color.R, el.Color.G, el.Color.B, el.Color.A)
	rl.Normal3f(el.Radius, 0.0, 1.0)

	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex2f(p.X, p.Y)

	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex2f(p.X+s.X, p.Y+s.Y)

	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex2f(p.X+s.X, p.Y)

	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex2f(p.X, p.Y+s.Y)

	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex2f(p.X+s.X, p.Y+s.Y)

	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex2f(p.X, p.Y)

	rl.End()
	rl.EndShaderMode()
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
