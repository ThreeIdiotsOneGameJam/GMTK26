//go:build !web

package shaders

const vertexShaderPath = "assets/shaders/base.vert"

func fragmentShaderPath(name string) string {
	return "assets/shaders/" + name + ".frag"
}
