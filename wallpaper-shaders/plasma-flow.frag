// Plasma flow — swirling organic color fields
// Shadertoy-compatible

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    float t = iTime * 0.15;

    float v1 = sin(uv.x * 5.0 + t);
    float v2 = sin(uv.y * 5.0 + t * 1.3);
    float v3 = sin((uv.x + uv.y) * 5.0 + t * 0.7);
    float v4 = sin(length(uv - 0.5) * 8.0 - t * 2.0);
    float v = (v1 + v2 + v3 + v4) * 0.25;

    // Snazzy palette
    vec3 cyan = vec3(0.34, 0.78, 1.0);
    vec3 magenta = vec3(1.0, 0.42, 0.76);
    vec3 green = vec3(0.35, 0.97, 0.56);

    vec3 col = mix(cyan, magenta, v * 0.5 + 0.5);
    col = mix(col, green, sin(v * 3.14) * 0.3);

    // Keep it dark for readability
    col *= 0.15 + 0.05 * sin(t * 0.5);

    // Subtle vignette
    float vig = 1.0 - 0.4 * length(uv - 0.5);
    col *= vig;

    fragColor = vec4(col, 1.0);
}
