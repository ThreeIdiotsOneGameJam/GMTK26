#version 100

precision highp float;

varying vec2 fragTexCoord;
varying vec4 fragColor;
varying vec3 fragNormal;

void main() {
	float radius = fragNormal.x;

	vec2 uv = fragTexCoord;
	float d = distance(uv, vec2(0.5)) / 0.70710678;
	float a = clamp((d - radius) / (1.0 - radius), 0.0, 1.0);

	gl_FragColor = mix(vec4(fragColor.rgb, 0.0), fragColor, a);
}
