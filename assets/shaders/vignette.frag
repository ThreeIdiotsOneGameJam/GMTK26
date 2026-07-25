#version 330

in vec2 fragTexCoord;
in vec4 fragColor;
in vec3 fragNormal;

out vec4 finalColor;

uniform float time;

void main() {
	float radius = fragNormal.x;
	float invRadius = 1.0 - radius;

	vec2 uv = fragTexCoord;
	float d = distance(uv, vec2(0.5)) / 0.70710678; // 0.7... is the length of X or Y on a normalize vector that points NW
	float a = clamp((d - radius) / (1.0 - radius), 0.0, 1.0);

	finalColor = mix(vec4(fragColor.rgb, 0.0), fragColor, a);
}
