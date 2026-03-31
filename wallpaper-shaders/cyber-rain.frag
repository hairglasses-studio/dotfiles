// Cyberpunk neon rain — Shadertoy-compatible
// Vertical rain streaks with cyan/magenta glow on black

#define RAIN_SPEED 0.8
#define COLS 80.0
#define GLOW 0.6

float hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
}

float rain(vec2 uv, float t) {
    vec2 id = floor(uv * vec2(COLS, 1.0));
    float offset = hash(id) * 6.28;
    float speed = 0.5 + hash(id + 0.5) * RAIN_SPEED;
    float drop = fract((uv.y + t * speed + offset) * (3.0 + hash(id + 1.0) * 8.0));
    float streak = smoothstep(0.0, 0.15, drop) * smoothstep(1.0, 0.5, drop);
    float col = fract(uv.x * COLS);
    streak *= smoothstep(0.4, 0.5, col) * smoothstep(0.6, 0.5, col);
    return streak;
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;

    float r1 = rain(uv, iTime) * 0.8;
    float r2 = rain(uv * 1.3 + 0.5, iTime * 0.7) * 0.4;
    float r3 = rain(uv * 0.8 + 1.2, iTime * 1.1) * 0.3;
    float r = r1 + r2 + r3;

    // Cyan base with magenta highlights
    vec3 cyan = vec3(0.34, 0.78, 1.0);
    vec3 magenta = vec3(1.0, 0.42, 0.76);
    vec3 col = mix(cyan, magenta, r2 / max(r, 0.01)) * r;

    // Glow
    col += col * GLOW;

    // Dark base with subtle gradient
    vec3 bg = vec3(0.02, 0.02, 0.04) + vec3(0.0, 0.01, 0.03) * uv.y;
    col = bg + col;

    fragColor = vec4(col, 1.0);
}
