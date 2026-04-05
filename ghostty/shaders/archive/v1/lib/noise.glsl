// lib/noise.glsl — Value noise and FBM (Fractal Brownian Motion)
// Requires: lib/hash.glsl (hash(vec2) function)

// Interpolated value noise (smooth random, range [0, 1])
float vnoise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(hash(i + vec2(0.0, 0.0)), hash(i + vec2(1.0, 0.0)), u.x),
               mix(hash(i + vec2(0.0, 1.0)), hash(i + vec2(1.0, 1.0)), u.x), u.y);
}

// Signed value noise (range [-1, 1])
float noise(vec2 p) {
    return -1.0 + 2.0 * vnoise(p);
}

// Fractal Brownian Motion — layered noise for natural-looking patterns
float fbm(vec2 p, int octaves) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < 8; i++) {
        if (i >= octaves) break;
        v += a * vnoise(p);
        p = rot * p * 2.0 + 0.1;
        a *= 0.5;
    }
    return v;
}

// Simple FBM with 5 octaves (most common usage)
float fbm(vec2 p) {
    return fbm(p, 5);
}
