#version 410 core
layout (points) in;
layout (triangle_strip, max_vertices = 380) out;

in VS_OUT {
    vec4 color;
} gs_in[];

uniform mat4 projection;
uniform float particle_size;

out vec2 fUV;
out vec4 fColor;
   
void main (void)
{
  vec4 P = gl_in[0].gl_Position;

  // a: left-bottom 
  vec2 va = P.xy + vec2(-0.5, -0.5) * particle_size;
  gl_Position = projection * vec4(va, P.zw);
  fUV = vec2(0.0, 0.0);
  fColor = gs_in[0].color;
  EmitVertex();  
  
  // b: left-top
  vec2 vb = P.xy + vec2(-0.5, 0.5) * particle_size;
  gl_Position = projection * vec4(vb, P.zw);
  fUV = vec2(0.0, 1.0);
  fColor = gs_in[0].color;
  EmitVertex();  
  
  // d: right-bottom
  vec2 vd = P.xy + vec2(0.5, -0.5) * particle_size;
  gl_Position = projection * vec4(vd, P.zw);
  fUV = vec2(1.0, 0.0);
  fColor = gs_in[0].color;
  EmitVertex();  

  // c: right-top
  vec2 vc = P.xy + vec2(0.5, 0.5) * particle_size;
  gl_Position = projection * vec4(vc, P.zw);
  fUV = vec2(1.0, 1.0);
  fColor = gs_in[0].color;
  EmitVertex();  

  EndPrimitive();  
}