// Chroma Key — color-distance transparency gradient
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders
// Makes pixels near a target color more transparent

void windowShader(inout vec4 color) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 c = x_Texture(uv);

    // Target color for maximum transparency (white by default)
    const vec3 median_color = vec3(1.0);
    // Maximum deviation per channel before full opacity
    const vec3 max_derivation = vec3(0.2);
    // Minimum opacity at the target color
    const float min_opacity = 0.9;
    // Gradient curvature exponent (1=linear, 2=quadratic)
    const float power = 2.0;

    // Normalized per-channel deviation from target
    vec3 norm_dev = abs(c.rgb - median_color) / max_derivation;
    float maxDev = max(max(norm_dev.r, norm_dev.g), norm_dev.b);

    if (c.a > 0.99 && maxDev < 1.0) {
        // Apply gradient curvature
        float curved = 1.0 - pow(1.0 - maxDev, power);
        float alpha = min_opacity + curved * (1.0 - min_opacity);
        c *= alpha;
    }

    color = c;
}
