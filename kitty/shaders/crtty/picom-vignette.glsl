#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Vignette — polynomial edge darkening with transparency on dark pixels
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec4 c = texture(u_input, uv);

    const float shadow_cutoff = 1.0;
    const int shadow_intensity = 3;

    vec2 center = vec2(0.5);
    vec2 dist = abs(uv - center) / center;

    // Darken pixels near edges using polynomial falloff
    float opacity = 1.0;
    opacity *= -pow(dist.y * shadow_cutoff, (5 / shadow_intensity) * 2) + 1.0;
    opacity *= -pow(dist.x * shadow_cutoff, (5 / shadow_intensity) * 2) + 1.0;

    // Apply vignette: darken edges, keep minimum brightness for readability
    float brightness = c.r + c.g + c.b;
    if (brightness < 0.6) {
        float vignetteAlpha = max(1.0 - opacity, 0.8);
        c.rgb *= vignetteAlpha;
    }

    o_color = c;
}
