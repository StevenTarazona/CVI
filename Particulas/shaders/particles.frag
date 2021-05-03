#version 410
uniform sampler2D tex0;
out vec4 FragColor;
in vec2 fUV;
in vec4 fColor;
void main (void)
{
  vec2 uv = fUV.xy;
  uv.y *= -1.0;
  vec4 texColor = texture(tex0, uv);
  FragColor = texColor * fColor;
}