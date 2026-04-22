// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Kaleidoscopic fractal — 64-fold symmetry with domain-warped FBM inside

const int   FOLDS        = 8;     // base wedge folds; iterated × MIRRORS times
const int   MIRRORS      = 3;     // each mirror halves wedge angle → 8 × 2^3 = 64-fold
const float ROT_SPEED    = 0.06;
const float WARP_AMOUNT  = 1.2;
const float INTENSITY    = 0.5;

vec3 kal_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.90, 0.18, 0.60); // magenta
    vec3 d = vec3(0.96, 0.70, 0.25); // gold
    vec3 e = vec3(0.20, 0.95, 0.60); // mint
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

float kal_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453123);
}

float kal_vnoise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(
        mix(kal_hash(i), kal_hash(i + vec2(1,0)), u.x),
        mix(kal_hash(i + vec2(0,1)), kal_hash(i + vec2(1,1)), u.x),
        u.y);
}

// 7-octave rotated FBM — heavy
float kal_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < 7; i++) {
        v += a * kal_vnoise(p);
        p = rot * p * 2.07 + 0.17;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);
    float a = atan(p.y, p.x) + x_Time * ROT_SPEED;

    // Base wedge fold
    float wedge = 6.28318 / float(FOLDS);
    float fa = mod(a, wedge);
    fa = abs(fa - wedge * 0.5);

    // Iterative mirror — each pass halves the wedge, doubling symmetry
    for (int i = 0; i < MIRRORS; i++) {
        wedge *= 0.5;
        fa = abs(fa - wedge);
    }

    // Reconstruct folded point in cartesian space
    vec2 fp = vec2(cos(fa), sin(fa)) * r;

    // Domain warp — two-layer FBM offset for swirling detail
    vec2 q = vec2(kal_fbm(fp * 3.0 + x_Time * 0.1),
                  kal_fbm(fp * 3.0 + vec2(5.2, 1.3) + x_Time * 0.1));
    vec2 w = vec2(kal_fbm(fp * 3.0 + WARP_AMOUNT * q + vec2(1.7, 9.2)),
                  kal_fbm(fp * 3.0 + WARP_AMOUNT * q + vec2(8.3, 2.8)));
    float density = kal_fbm(fp * 3.0 + WARP_AMOUNT * w);

    // Ring pulses marching outward
    float ringCoord = r * 5.0 - x_Time * 0.9;
    float ringPhase = fract(ringCoord);
    float ring = smoothstep(0.47, 0.5, ringPhase) * (1.0 - smoothstep(0.5, 0.53, ringPhase));

    // Color evolves with radius + density + time
    vec3 col = kal_pal(fract(density * 0.9 + r * 0.3 + x_Time * 0.05));

    // Brightness shaping
    float bright = smoothstep(0.35, 0.9, density) * 0.9 + ring * 0.7;
    // Outer fade to keep corners clean
    float outerFade = 1.0 - smoothstep(0.35, 0.7, r);
    col *= bright * outerFade;

    // Composite
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
