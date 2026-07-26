#version 100

#ifdef GL_FRAGMENT_PRECISION_HIGH
precision highp float;
#else
precision mediump float;
#endif

varying vec2 fragTexCoord;

uniform vec4 vignetteColor;
uniform float vignetteRadius;

void main() {
	vec2 uv = fragTexCoord;
	float d = distance(uv, vec2(0.5)) / 0.70710678;
	float a = clamp((d - vignetteRadius) / (1.0 - vignetteRadius), 0.0, 1.0);
	float opacity = vignetteColor.a * a;

	gl_FragColor = vec4(vignetteColor.rgb * opacity, opacity);
}
