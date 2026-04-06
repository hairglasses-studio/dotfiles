#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;


void main()
{
    vec2 uv = gl_FragCoord.xy/u_resolution;
    vec4 color = texture(u_input, uv);
    o_color = vec4(1.0 - color.x, 1.0 - color.y, 1.0 - color.z, color.w);
}
