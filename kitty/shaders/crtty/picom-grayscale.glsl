#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Grayscale — CIELAB luminance-based desaturation filter
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec4 c = texture(u_input, uv);

    // CIELAB luma based on human tristimulus response
    float g = 0.2126 * c.r + 0.7152 * c.g + 0.0722 * c.b;

    o_color = vec4(vec3(g), c.a);
}
