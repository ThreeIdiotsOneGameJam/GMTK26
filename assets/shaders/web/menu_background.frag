#version 100

precision highp float;

varying vec2 fragTexCoord;
varying vec4 fragColor;

uniform float time;
uniform vec2 size;

const float SQRT3 = 1.7320508;

vec3 cube_round(vec3 c) {
	float q = floor(c.x + 0.5);
	float r = floor(c.y + 0.5);
	float s = floor(c.z + 0.5);

	float qDiff = abs(q - c.x);
	float rDiff = abs(r - c.y);
	float sDiff = abs(s - c.z);

	if (qDiff > rDiff && qDiff > sDiff) {
		q = -r - s;
	} else if (rDiff > sDiff) {
		r = -q - s;
	} else {
		s = -q - r;
	}

	return vec3(q, r, s);
}

vec3 axial_to_cube(vec2 a) {
	return vec3(a.x, a.y, -a.x-a.y);
}

vec2 cube_to_axial(vec3 c) {
	return vec2(c.x, c.y);
}

vec2 axial_to_hex(vec2 a) {
	vec2 axial = cube_to_axial(cube_round(axial_to_cube(a)));
	float parity = mod(axial.x, 2.0);
	float col = axial.x;
	float row = axial.y + (axial.x-parity)/2.0;

	return vec2(col, row);
}

vec2 pixel_to_hex(vec2 p) {
	float q = (2.0 * p.x) / (3.0 * 48.0);
	float axialR = (-p.x)/(3.0*48.0) + p.y/(SQRT3*48.0);
	return axial_to_hex(vec2(q, axialR));
}

float hash(vec2 p) {
	p = fract(p * vec2(123.34, 456.21));
	p += dot(p, p + 45.32);
	return fract(p.x * p.y);
}

void main() {
	vec2 pos = fragTexCoord * size;
	vec2 hex = pixel_to_hex(pos);

	vec3 baseColor = vec3(49.0, 44.0, 88.0) / 255.0;
	vec3 focusColor = vec3(60.0, 65.0, 140.0) / 255.0;

	vec2 tHex = hex + vec2(hash(hex), hash(hex + hash(hex))) * 12.0;

	float t = sin((-time * 10.0 + tHex.y + tHex.x) * 0.2);
	t = clamp(t, 0.0, 1.0);

	vec3 color = mix(baseColor, focusColor, t);

	if (mod(hex.x, 2.0) >= 1.0) {
		color += 3.0 / 255.0;
	}
	if (mod(hex.y, 2.0) >= 1.0) {
		color += 6.0 / 255.0;
	}

	gl_FragColor = vec4(color, 1.0);
}
