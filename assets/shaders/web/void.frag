#version 100

precision highp float;

varying vec2 fragTexCoord;
varying vec3 fragNormal;
varying vec4 fragColor;
varying vec2 fragPos;

uniform float time;

void main() {
	vec2 p = fragPos * 0.5;

	float borderFade = mix(-2.0, 1.0, fragNormal.r);
	borderFade += sin(time * 5.0 + fragTexCoord.x + fragTexCoord.y + p.x + p.y) * 0.2;
	borderFade = clamp(borderFade, 0.0, 1.0);
	borderFade = floor(borderFade * 4.0 + 0.5) / 4.0;

	float baseAlpha = sin(time * 2.0 + fragTexCoord.x + fragTexCoord.y) * 0.5 + 0.5;
	baseAlpha *= 0.2;

	gl_FragColor = mix(vec4(fragColor.rgb, baseAlpha), fragColor, borderFade);
}
