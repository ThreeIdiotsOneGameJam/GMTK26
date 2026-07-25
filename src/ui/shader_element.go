package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/rlutil"
)

// frozenShaderTime holds the last animated time so reduced motion freezes
// the menu background in place instead of jumping to t=0.
var frozenShaderTime float32

func Shader() *ShaderElement {
	el := &ShaderElement{}
	el.BaseElement = NewBaseElement(el)

	return el
}

type ShaderElement struct {
	BaseElement[*ShaderElement]

	Shader *rl.Shader
}

func (el *ShaderElement) WithShader(shader *rl.Shader) *ShaderElement {
	el.Shader = shader
	return el
}

func shaderTime() float32 {
	if settings.Current.ReducedMotion {
		return frozenShaderTime
	}
	frozenShaderTime = float32(rl.GetTime())
	return frozenShaderTime
}

func (el *ShaderElement) draw() {
	shader := *el.Shader
	if !rl.IsShaderValid(shader) {
		return
	}

	s := el.Parent.Size().Vec2()
	p := el.Parent.AbsolutePos().Vec2()

	timeLoc := rl.GetShaderLocation(shader, "time")
	rl.SetShaderValue(shader, timeLoc, []float32{shaderTime()}, rl.ShaderUniformFloat)

	sizeLoc := rl.GetShaderLocation(shader, "size")
	rl.SetShaderValue(shader, sizeLoc, []float32{s.X, s.Y}, rl.ShaderUniformVec2)

	rl.BeginShaderMode(shader)
	rl.Begin(rl.Triangles)

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
}
