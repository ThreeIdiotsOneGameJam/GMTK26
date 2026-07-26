#version 100

#ifdef GL_FRAGMENT_PRECISION_HIGH
precision highp float;
#else
precision mediump float;
#endif

varying vec2 fragTexCoord;
varying vec4 fragColor;

uniform float time;

void main() {
	vec2 uv = fragTexCoord + vec2(time * -30.0);
	float v = mod(floor(uv.x / 64.0) + floor(uv.y / 64.0), 2.0);
	vec3 col = mix(
		vec3(127.0, 127.0, 127.0) / 255.0,
		vec3(191.0, 191.0, 191.0) / 255.0,
		v
	);
	gl_FragColor = vec4(col, 1.0);
}
