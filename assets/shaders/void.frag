#version 330

in vec2 fragTexCoord;
in vec3 fragNormal;
in vec4 fragColor;
in vec2 fragPos;

out vec4 finalColor;

uniform float time;

void main() {
	vec2 p = fragPos * 0.5;

	float borderFade = mix(-2.0, 1.0, fragNormal.r);
	borderFade += sin(time * 5.0 + fragTexCoord.x + fragTexCoord.y + p.x + p.y) * 0.2;
	borderFade = min(max(borderFade, 0.0), 1.0);
	borderFade = round(borderFade * 4.0) / 4.0;

	float baseAlpha = sin(time * 2.0 + fragTexCoord.x + fragTexCoord.y) * 0.5 + 0.5;
	baseAlpha *= 0.2;

	finalColor = mix(vec4(fragColor.rgb, baseAlpha), fragColor, borderFade);
}
