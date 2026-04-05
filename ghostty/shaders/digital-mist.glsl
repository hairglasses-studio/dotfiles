precision highp float;
// Digital Mist — Sparse rising particles on dark background
// Category: Ambient | Cost: LOW | Source: original (demoscene research)
// Tiny luminous particles drift upward like digital dust or embers.
// Extremely subtle — designed for long coding sessions.

float hash21(vec2 p) {
    p = fract(p * vec2(123.34, 456.21));
    p += dot(p, p + 45.32);
    return fract(p.x * p.y);
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 term = texture(iChannel0, uv);
    float termLum = dot(term.rgb, vec3(0.299, 0.587, 0.114));

    // Snazzy palette
    vec3 cyan    = vec3(0.341, 0.780, 1.0);   // #57c7ff
    vec3 magenta = vec3(1.0, 0.416, 0.757);   // #ff6ac1
    vec3 green   = vec3(0.353, 0.969, 0.557);  // #5af78e

    vec3 particles = vec3(0.0);
    float t = iTime;

    for (int i = 0; i < 30; i++) {
        float fi = float(i);
        float seed = hash21(vec2(fi, fi * 13.7));

        // Deterministic position: slow rise with gentle sway
        float life = fract(t * (0.03 + seed * 0.04) + seed);
        vec2 pos;
        pos.x = fract(seed * 7.13 + sin(t * 0.3 + fi * 0.7) * 0.03);
        pos.y = life;

        float dist = length((uv - pos) * vec2(iResolution.x / iResolution.y, 1.0));
        float size = 0.002 * (1.0 - life * 0.5);
        float brightness = smoothstep(size * 3.0, size * 0.3, dist);

        // Fade in at bottom, fade out at top
        brightness *= smoothstep(0.0, 0.1, life) * smoothstep(1.0, 0.8, life);

        // Color: cycle through palette based on particle seed
        vec3 pcol = seed < 0.33 ? cyan : (seed < 0.66 ? magenta : green);
        particles += pcol * brightness * 0.15;
    }

    // Mask particles behind bright text
    vec3 result = term.rgb + particles * (1.0 - termLum * 0.85);

    fragColor = vec4(result, term.a);
}
