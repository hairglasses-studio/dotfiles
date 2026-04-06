#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Gamma — power-law gamma correction
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec4 c = texture(u_input, uv);

    const float gamma = 0.7;
    const float inv_gamma = 1.0 / gamma;

    c.r = pow(c.r, inv_gamma);
    c.g = pow(c.g, inv_gamma);
    c.b = pow(c.b, inv_gamma);

    o_color = c;
}
