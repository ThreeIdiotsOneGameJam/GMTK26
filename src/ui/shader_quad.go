package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/rlutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func drawShaderQuad(shader rl.Shader, position, size vec.Vec2, configureVertices func()) {
	rl.BeginShaderMode(shader)
	rl.Begin(rl.Triangles)

	if configureVertices != nil {
		configureVertices()
	}

	rl.TexCoord2f(0, 0)
	rlutil.Vertex2f(position.X, position.Y)

	rl.TexCoord2f(1, 1)
	rlutil.Vertex2f(position.X+size.X, position.Y+size.Y)

	rl.TexCoord2f(1, 0)
	rlutil.Vertex2f(position.X+size.X, position.Y)

	rl.TexCoord2f(0, 1)
	rlutil.Vertex2f(position.X, position.Y+size.Y)

	rl.TexCoord2f(1, 1)
	rlutil.Vertex2f(position.X+size.X, position.Y+size.Y)

	rl.TexCoord2f(0, 0)
	rlutil.Vertex2f(position.X, position.Y)

	rl.End()
	rl.EndShaderMode()
}
