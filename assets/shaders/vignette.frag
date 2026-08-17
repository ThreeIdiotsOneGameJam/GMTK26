#version 330

in vec2 fragTexCoord;

out vec4 finalColor;

uniform vec4 vignetteColor;
uniform float vignetteRadius;

void main() {
	vec2 uv = fragTexCoord;
	float d = distance(uv, vec2(0.5)) / 0.70710678; // 0.7... is the length of X or Y on a normalize vector that points NW
	float a = clamp((d - vignetteRadius) / (1.0 - vignetteRadius), 0.0, 1.0);
	float opacity = vignetteColor.a * a;

	finalColor = vec4(vignetteColor.rgb * opacity, opacity);
}
