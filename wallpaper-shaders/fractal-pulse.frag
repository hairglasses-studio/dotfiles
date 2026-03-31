// Fractal pulse — animated Julia set with neon palette
// Shadertoy-compatible

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = (fragCoord - 0.5 * iResolution.xy) / iResolution.y;
    uv *= 2.5;

    // Animate Julia set parameter
    float t = iTime * 0.1;
    vec2 c = vec2(-0.7 + 0.15 * cos(t), 0.27015 + 0.1 * sin(t * 1.3));

    vec2 z = uv;
    float iter = 0.0;
    const float MAX_ITER = 64.0;

    for (float i = 0.0; i < MAX_ITER; i++) {
        z = vec2(z.x * z.x - z.y * z.y, 2.0 * z.x * z.y) + c;
        if (dot(z, z) > 4.0) break;
        iter = i;
    }

    float f = iter / MAX_ITER;

    // Smooth coloring
    f = sqrt(f);

    // Snazzy neon palette
    vec3 col = vec3(0.0);
    col += 0.5 + 0.5 * cos(6.28 * (f * 2.0 + vec3(0.0, 0.33, 0.67)));
    col *= vec3(0.34, 0.78, 1.0) * 0.5 + vec3(1.0, 0.42, 0.76) * 0.5;

    // Darken for wallpaper use
    col *= 0.3 * f;

    // Black for converged points (inside the set)
    if (iter >= MAX_ITER - 1.0) col = vec3(0.01, 0.01, 0.02);

    fragColor = vec4(col, 1.0);
}
