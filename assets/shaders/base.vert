#version 330

// Just taken from raylib website

// Input vertex attributes
in vec3 vertexPosition;
in vec2 vertexTexCoord;
in vec3 vertexNormal;
in vec4 vertexColor;

// Input uniform values
uniform mat4 mvp;

// Output vertex attributes (to fragment shader)
out vec2 fragTexCoord;
out vec3 fragNormal;
out vec4 fragColor;
out vec2 fragPos;

void main() {
    // Send vertex attributes to fragment shader
    fragTexCoord = vertexTexCoord;
    fragColor = vertexColor;
    fragNormal = vertexNormal;
    fragPos = vertexPosition.xy;

    // Calculate final vertex position
    gl_Position = mvp*vec4(vertexPosition, 1.0);
}
