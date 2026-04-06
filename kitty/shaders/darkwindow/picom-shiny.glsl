// Shiny — animated diagonal specular highlight sweep
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders (via yshui #295)
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void windowShader(inout vec4 color) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 c = x_Texture(uv);

    // Animated sweep: diagonal highlight band moves across the terminal
    const float amt = 10.0; // cycle period in seconds
    float pct = mod(x_Time, amt) / amt * 1000.0;
    float factor = max(x_WindowSize.x, x_WindowSize.y) / 150.0;
    pct *= factor;

    // Pixel position in screen space for sweep calculation
    vec2 pos = x_PixelPos;
    float diag = pos.x + pos.y;

    // Two-band highlight: primary + secondary trail
    if ((diag < pct * 4.0 && diag > pct * 4.0 - 0.5 * pct) ||
        (diag < pct * 4.0 - 0.8 * pct && diag > pct * 3.0))
        c.rgb *= 2.0;

    color = c;
}
