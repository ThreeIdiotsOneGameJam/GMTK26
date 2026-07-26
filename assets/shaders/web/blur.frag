#version 100

precision highp float;

varying vec2 fragTexCoord;
varying vec4 fragColor;

uniform sampler2D texture0;
uniform vec2 halfPixel;
uniform float offset;
uniform float upsample;

void main() {
	vec2 uv = fragTexCoord;
	vec2 hp = halfPixel * offset;
	vec4 sum;

	if (upsample < 0.5) {
		sum = texture2D(texture0, uv) * 4.0;
		sum += texture2D(texture0, uv - hp);
		sum += texture2D(texture0, uv + hp);
		sum += texture2D(texture0, uv + vec2(hp.x, -hp.y));
		sum += texture2D(texture0, uv - vec2(hp.x, -hp.y));
		gl_FragColor = (sum / 8.0) * fragColor;
	} else {
		sum = texture2D(texture0, uv + vec2(-hp.x * 2.0, 0.0));
		sum += texture2D(texture0, uv + vec2(-hp.x, hp.y)) * 2.0;
		sum += texture2D(texture0, uv + vec2(0.0, hp.y * 2.0));
		sum += texture2D(texture0, uv + vec2(hp.x, hp.y)) * 2.0;
		sum += texture2D(texture0, uv + vec2(hp.x * 2.0, 0.0));
		sum += texture2D(texture0, uv + vec2(hp.x, -hp.y)) * 2.0;
		sum += texture2D(texture0, uv + vec2(0.0, -hp.y * 2.0));
		sum += texture2D(texture0, uv + vec2(-hp.x, -hp.y)) * 2.0;
		gl_FragColor = (sum / 12.0) * fragColor;
	}
}
