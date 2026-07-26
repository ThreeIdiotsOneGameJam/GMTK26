#version 100

precision highp float;

attribute vec3 vertexPosition;
attribute vec2 vertexTexCoord;
attribute vec3 vertexNormal;
attribute vec4 vertexColor;

uniform mat4 mvp;

varying vec2 fragTexCoord;
varying vec3 fragNormal;
varying vec4 fragColor;
varying vec2 fragPos;

void main() {
    fragTexCoord = vertexTexCoord;
    fragColor = vertexColor;
    fragNormal = vertexNormal;
    fragPos = vertexPosition.xy;
    gl_Position = mvp * vec4(vertexPosition, 1.0);
}
