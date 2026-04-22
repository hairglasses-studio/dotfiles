// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Quantum probability field — 24 particles with Gaussian cloud + interference

const int   PARTICLES   = 24;
const int   CLOUD_TAPS  = 16;
const float CLOUD_SIZE  = 0.06;
const float FIELD_AMP   = 0.35;
const float INTENSITY   = 0.55;

vec3 qf_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.90, 0.20, 0.55); // magenta
    vec3 d = vec3(0.96, 0.85, 0.40); // warm
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float qf_hash(float n) { return fract(sin(n * 127.1) * 43758.5); }
vec2  qf_hash2(float n) { return vec2(qf_hash(n), qf_hash(n * 1.37 + 11.0)); }

// Animated particle position
vec2 particlePos(int i, float t) {
    float seed = float(i) * 7.31;
    vec2 base = qf_hash2(seed) * 1.6 - 0.8;
    base.x *= x_WindowSize.x / x_WindowSize.y;
    base += 0.18 * vec2(sin(t * (0.4 + qf_hash(seed * 3.0)) + seed),
                        cos(t * (0.3 + qf_hash(seed * 5.0)) + seed * 1.3));
    return base;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    // Quantum probability cloud per particle — sum of Gaussian samples at offsets
    for (int i = 0; i < PARTICLES; i++) {
        vec2 pp = particlePos(i, x_Time);
        float seed = float(i) * 7.31;
        vec3 pc = qf_pal(fract(float(i) * 0.04 + x_Time * 0.03));

        // Uncertainty: sample cloud at N offsets, accumulate probability
        float prob = 0.0;
        for (int k = 0; k < CLOUD_TAPS; k++) {
            float fk = float(k);
            vec2 offset = qf_hash2(seed + fk * 1.71) - 0.5;
            offset *= CLOUD_SIZE * 2.0;
            vec2 samp = pp + offset;
            float d = length(p - samp);
            prob += exp(-d * d * 80.0);
        }
        prob /= float(CLOUD_TAPS);

        // Quantum wave component — sinusoidal phase oscillation
        float wavePhase = length(p - pp) * 30.0 - x_Time * 3.0 + seed;
        float wave = (0.5 + 0.5 * cos(wavePhase)) * prob;

        col += pc * (prob * 1.5 + wave * 0.6);
    }

    // Global interference field — sum all particles' wave contributions as complex numbers
    float sumR = 0.0, sumI = 0.0;
    for (int i = 0; i < PARTICLES; i++) {
        vec2 pp = particlePos(i, x_Time);
        float seed = float(i) * 7.31;
        float phase = length(p - pp) * 25.0 - x_Time * 2.0 + seed;
        float amp = 1.0 / (0.1 + length(p - pp) * 5.0);
        sumR += cos(phase) * amp * 0.5;
        sumI += sin(phase) * amp * 0.5;
    }
    float fieldMag = sqrt(sumR * sumR + sumI * sumI) * 0.02;
    float fieldPhase = atan(sumI, sumR) / 6.28318 + 0.5;
    col += qf_pal(fract(fieldPhase + x_Time * 0.05)) * fieldMag * FIELD_AMP;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.8, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
