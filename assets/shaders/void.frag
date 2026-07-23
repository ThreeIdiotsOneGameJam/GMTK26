#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

out vec4 finalColor;

uniform float time;

void main() {
	finalColor = vec4(0.0, 0.0, 0.0, min(max(mix(-2.0, 1.0, fragColor.r), 0.0), 1.0) * 0.3);
}
