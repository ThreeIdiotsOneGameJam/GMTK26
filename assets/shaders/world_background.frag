#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

out vec4 finalColor;

uniform float time;

void main() {
	vec2 uv = fragTexCoord + vec2(time * -30.0);
	float v = mod(floor((floor(uv.x / 64.0) + floor(uv.y / 64.0))), 2.0);
	vec3 col = mix(vec3(127.0, 127.0, 127.0) / 255.0, vec3(191.0, 191.0, 191.0) / 255.0, v);
	finalColor = vec4(col, 1.0);
}
