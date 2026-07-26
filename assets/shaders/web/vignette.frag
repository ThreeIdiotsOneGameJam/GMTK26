#version 100

#ifdef GL_FRAGMENT_PRECISION_HIGH
precision highp float;
#else
precision mediump float;
#endif

varying vec2 fragTexCoord;
varying vec4 fragColor;
varying vec3 fragNormal;

void main() {
	float radius = fragNormal.x;

	vec2 uv = fragTexCoord;
	float d = distance(uv, vec2(0.5)) / 0.70710678;
	float a = clamp((d - radius) / (1.0 - radius), 0.0, 1.0);
	float opacity = fragColor.a * a;

	gl_FragColor = vec4(fragColor.rgb * opacity, opacity);
}
