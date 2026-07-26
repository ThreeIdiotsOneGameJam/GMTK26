//go:build web

package shaders

const vertexShaderPath = "assets/shaders/web/base.vert"

func fragmentShaderPath(name string) string {
	return "assets/shaders/web/" + name + ".frag"
}
