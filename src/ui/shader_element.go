package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/rlutil"
)

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

func (el *ShaderElement) draw() {
	shader := *el.Shader
	if !rl.IsShaderValid(shader) {
		return
	}

	s := el.Parent.Size().Vec2()
	p := el.Parent.AbsolutePos().Vec2()

	timeLoc := rl.GetShaderLocation(shader, "time")
	rl.SetShaderValue(shader, timeLoc, []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)

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
