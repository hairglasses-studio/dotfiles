#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Pixelize — animated block mosaic that oscillates between sharp and pixelated
// Category: Post-FX | Cost: MED | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;

    float t = 0.5 + 0.5 * sin(u_time * 1.57);
    float block = mix(40.0, 1.0, t);

    vec2 blockUv = (floor(gl_FragCoord.xy / block) * block + block * 0.5) / u_resolution;
    blockUv = clamp(blockUv, vec2(0.0), vec2(1.0));

    o_color = texture(u_input, blockUv);
}
