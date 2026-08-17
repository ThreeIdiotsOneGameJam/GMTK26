//go:build !web

package shaders

import "github.com/threeidiotsonegamejam/gmtk26/src/global"

var vertexShaderPath = global.AssetDir + "/shaders/base.vert"

func fragmentShaderPath(name string) string {
	return global.AssetDir + "/shaders/" + name + ".frag"
}
