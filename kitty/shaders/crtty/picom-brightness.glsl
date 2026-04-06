#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Brightness — proportional brightness adjustment
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec4 c = texture(u_input, uv);

    const float brightness_level = 0.7; // 0.0 = black, 1.0 = unchanged

    c.rgb *= brightness_level;

    o_color = c;
}
