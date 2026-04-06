#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Shiny — animated diagonal specular highlight sweep
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders (via yshui #295)
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec4 c = texture(u_input, uv);

    // Animated sweep: diagonal highlight band moves across the terminal
    const float amt = 10.0; // cycle period in seconds
    float pct = mod(u_time, amt) / amt * 1000.0;
    float factor = max(u_resolution.x, u_resolution.y) / 150.0;
    pct *= factor;

    // Pixel position in screen space for sweep calculation
    vec2 pos = gl_FragCoord.xy;
    float diag = pos.x + pos.y;

    // Two-band highlight: primary + secondary trail
    if ((diag < pct * 4.0 && diag > pct * 4.0 - 0.5 * pct) ||
        (diag < pct * 4.0 - 0.8 * pct && diag > pct * 3.0))
        c.rgb *= 2.0;

    o_color = c;
}
