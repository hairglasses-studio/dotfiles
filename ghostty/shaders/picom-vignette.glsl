precision highp float;
// Vignette — polynomial edge darkening with transparency on dark pixels
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 c = texture(iChannel0, uv);

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

    fragColor = c;
}
