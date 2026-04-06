precision highp float;
// Dissolve — Balatro-style burn dissolve with FBM noise
// Category: Post-FX | Cost: MED | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders
// Original used picom opacity transitions; adapted to iTime-driven continuous loop

float rand(vec2 co) {
    return fract(sin(dot(co, vec2(12.9898, 78.233))) * 43758.5453);
}

float value_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    float a = rand(i);
    float b = rand(i + vec2(1.0, 0.0));
    float c = rand(i + vec2(0.0, 1.0));
    float d = rand(i + vec2(1.0, 1.0));
    return mix(mix(a, b, u.x), mix(c, d, u.x), u.y);
}

float fbm(vec2 p) {
    float total = 0.0, freq = 1.0, amp = 0.5, maxVal = 0.0;
    for (int i = 0; i < 4; i++) {
        total += amp * value_noise(p * freq);
        maxVal += amp;
        amp *= 0.5;
        freq *= 2.0;
    }
    return total / maxVal;
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 term = texture(iChannel0, uv);

    const vec4 burn_color = vec4(1.0, 0.5, 0.0, 1.0); // orange burn edge
    const float burn_size = 0.03;
    const float noise_zoom = 5.0;

    // Continuous dissolve cycle: 0→1→0 over 8 seconds
    float progress = 0.5 + 0.5 * sin(iTime * 0.785); // ~8s full cycle

    float noise = fbm(uv * noise_zoom);

    float actual_burn = burn_size;
    if (progress < 0.001 || progress > 0.999) actual_burn = 0.0;

    float alpha = smoothstep(noise - actual_burn, noise, progress);
    float border = smoothstep(noise, noise + actual_burn, progress);

    vec4 result = term;
    result.rgb = mix(burn_color.rgb, term.rgb, border);
    result.a *= alpha;

    fragColor = result;
}
