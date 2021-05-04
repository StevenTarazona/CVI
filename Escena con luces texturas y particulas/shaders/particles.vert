#version 410 core
layout (location = 0) in vec3 aPos;
layout (location = 1) in vec4 aColor;
layout (location = 2) in float aSize;

out VS_OUT {
    vec4 color;
} vs_out;

out float size;

uniform mat4 model;
uniform mat4 view;

void main()
{
    vs_out.color = aColor;
    size = aSize;
    gl_Position = model * view * vec4(aPos, 1.0);
}