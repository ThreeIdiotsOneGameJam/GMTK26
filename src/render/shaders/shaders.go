package shaders

import rl "github.com/gen2brain/raylib-go/raylib"

var MenuBackground rl.Shader
var WorldBackground rl.Shader
var WorldBackgroundTimeLoc int32
var Vignette rl.Shader
var Void rl.Shader
var VoidTimeLoc int32

func Load() {
	MenuBackground = LoadShader("menu_background")
	WorldBackground = LoadShader("world_background")
	WorldBackgroundTimeLoc = rl.GetShaderLocation(WorldBackground, "time")
	Vignette = LoadShader("vignette")
	Void = LoadShader("void")
	VoidTimeLoc = rl.GetShaderLocation(Void, "time")
}

// LoadShader loads the platform-appropriate GLSL version of a fragment shader.
func LoadShader(name string) rl.Shader {
	return rl.LoadShader(vertexShaderPath, fragmentShaderPath(name))
}

func Unload() {
	rl.UnloadShader(MenuBackground)
	rl.UnloadShader(WorldBackground)
	rl.UnloadShader(Vignette)
	rl.UnloadShader(Void)
}
