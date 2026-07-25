#version 330

// Dual Kawase blur (downsample / upsample). Much softer on text and hard
// edges than a plus-shaped gaussian.

in vec2 fragTexCoord;
in vec4 fragColor;

out vec4 finalColor;

uniform sampler2D texture0;
uniform vec2 halfPixel;
uniform float offset;
uniform float upsample; // 0 = downsample, 1 = upsample

void main() {
	vec2 uv = fragTexCoord;
	vec2 hp = halfPixel * offset;
	vec4 sum;

	if (upsample < 0.5) {
		sum = texture(texture0, uv) * 4.0;
		sum += texture(texture0, uv - hp);
		sum += texture(texture0, uv + hp);
		sum += texture(texture0, uv + vec2(hp.x, -hp.y));
		sum += texture(texture0, uv - vec2(hp.x, -hp.y));
		finalColor = (sum / 8.0) * fragColor;
	} else {
		sum = texture(texture0, uv + vec2(-hp.x * 2.0, 0.0));
		sum += texture(texture0, uv + vec2(-hp.x, hp.y)) * 2.0;
		sum += texture(texture0, uv + vec2(0.0, hp.y * 2.0));
		sum += texture(texture0, uv + vec2(hp.x, hp.y)) * 2.0;
		sum += texture(texture0, uv + vec2(hp.x * 2.0, 0.0));
		sum += texture(texture0, uv + vec2(hp.x, -hp.y)) * 2.0;
		sum += texture(texture0, uv + vec2(0.0, -hp.y * 2.0));
		sum += texture(texture0, uv + vec2(-hp.x, -hp.y)) * 2.0;
		finalColor = (sum / 12.0) * fragColor;
	}
}
