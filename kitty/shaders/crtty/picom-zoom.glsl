#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Zoom — magnification / demagnification centered on terminal
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec2 center = vec2(0.5);

    const float scale = 0.5; // >1 = zoom in, <1 = zoom out

    // Displace UV around center
    vec2 newUv = (uv - center) * (1.0 / scale) + center;

    // Clamp to bounds — return black for out-of-range
    if (newUv.x > 1.0 || newUv.x < 0.0 || newUv.y > 1.0 || newUv.y < 0.0) {
        o_color = vec4(0.0);
        return;
    }

    o_color = texture(u_input, newUv);
}
