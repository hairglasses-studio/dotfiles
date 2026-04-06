precision highp float;
// Shiny — animated diagonal specular highlight sweep
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders (via yshui #295)
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 c = texture(iChannel0, uv);

    // Animated sweep: diagonal highlight band moves across the terminal
    const float amt = 10.0; // cycle period in seconds
    float pct = mod(iTime, amt) / amt * 1000.0;
    float factor = max(iResolution.x, iResolution.y) / 150.0;
    pct *= factor;

    // Pixel position in screen space for sweep calculation
    vec2 pos = fragCoord;
    float diag = pos.x + pos.y;

    // Two-band highlight: primary + secondary trail
    if ((diag < pct * 4.0 && diag > pct * 4.0 - 0.5 * pct) ||
        (diag < pct * 4.0 - 0.8 * pct && diag > pct * 3.0))
        c.rgb *= 2.0;

    fragColor = c;
}
