package shaders

import rl "github.com/gen2brain/raylib-go/raylib"

var MenuBackground rl.Shader
var WorldBackground rl.Shader
var WorldBackgroundTimeLoc int32
var Vignette rl.Shader
var Void rl.Shader
var VoidTimeLoc int32

func Load() {
	MenuBackground = rl.LoadShader("assets/shaders/base.vert", "assets/shaders/menu_background.frag")
	WorldBackground = rl.LoadShader("assets/shaders/base.vert", "assets/shaders/world_background.frag")
	WorldBackgroundTimeLoc = rl.GetShaderLocation(WorldBackground, "time")
	Vignette = rl.LoadShader("assets/shaders/base.vert", "assets/shaders/vignette.frag")
	Void = rl.LoadShader("assets/shaders/base.vert", "assets/shaders/void.frag")
	VoidTimeLoc = rl.GetShaderLocation(Void, "time")
}

func Unload() {
	rl.UnloadShader(MenuBackground)
	rl.UnloadShader(WorldBackground)
	rl.UnloadShader(Vignette)
	rl.UnloadShader(Void)
}
