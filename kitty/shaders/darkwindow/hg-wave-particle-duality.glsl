// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Wave-particle duality — double-slit experiment: parallel waves + particle detection hits

const int   PARTICLES = 80;
const float SLIT_SEP  = 0.1;
const float SCREEN_X  = 0.5;
const float SOURCE_X  = -0.5;
const float INTENSITY = 0.55;

vec3 wp_pal(float t) {
    vec3 cyan = vec3(0.20, 0.85, 0.95);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 mag  = vec3(0.95, 0.30, 0.70);
    vec3 gold = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, gold, s - 2.0);
    else              return mix(gold, cyan, s - 3.0);
}

float wp_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.02, 0.05);

    // Slit barrier — vertical wall
    float barrierX = 0.0;
    float slitY1 = SLIT_SEP * 0.5;
    float slitY2 = -SLIT_SEP * 0.5;
    if (abs(p.x - barrierX) < 0.005 && abs(p.y - slitY1) > 0.015 && abs(p.y - slitY2) > 0.015) {
        col = vec3(0.15, 0.15, 0.2);
        _wShaderOut = vec4(col, 1.0);
        return;
    }

    // Source point on left
    float sourceD = length(p - vec2(SOURCE_X, 0.0));
    col += vec3(1.0, 0.95, 0.8) * exp(-sourceD * sourceD * 3000.0) * 0.9;

    // Wave pattern (left of barrier) — circular wavefront from source
    if (p.x < barrierX) {
        float waveD = length(p - vec2(SOURCE_X, 0.0));
        float wave = sin(waveD * 80.0 - x_Time * 8.0);
        float waveIntensity = pow(max(0.0, wave), 2.0);
        col += wp_pal(0.7) * waveIntensity * 0.3;
    }

    // Wave pattern (right of barrier) — two sources (from each slit)
    if (p.x > barrierX) {
        vec2 slit1 = vec2(barrierX, slitY1);
        vec2 slit2 = vec2(barrierX, slitY2);
        float d1 = length(p - slit1);
        float d2 = length(p - slit2);
        float wave1 = sin(d1 * 80.0 - x_Time * 8.0) / (0.3 + d1 * 3.0);
        float wave2 = sin(d2 * 80.0 - x_Time * 8.0) / (0.3 + d2 * 3.0);
        float combined = wave1 + wave2;
        // Interference pattern
        float inter = pow(max(0.0, combined), 2.0) * 0.8;
        col += wp_pal(fract(p.y * 2.0 + x_Time * 0.04)) * inter * 0.4;

        // Slit emitter dots
        col += vec3(1.0) * exp(-d1 * d1 * 2000.0) * 0.6;
        col += vec3(1.0) * exp(-d2 * d2 * 2000.0) * 0.6;
    }

    // Detection screen on right
    if (abs(p.x - SCREEN_X) < 0.005) {
        col = vec3(0.15, 0.17, 0.22);
    }

    // Particle detection hits — accumulated on detection screen
    for (int i = 0; i < PARTICLES; i++) {
        float fi = float(i);
        float seed = fi * 3.71;
        // Lifetime (fade in/out)
        float hitCycle = 4.0;
        float hitPhase = fract((x_Time + wp_hash(seed) * hitCycle) / hitCycle);
        // Y position follows interference probability distribution
        float yBias = (wp_hash(seed * 3.1) - 0.5) * 0.5;
        // Probability peaks (sinusoidal)
        float weighted = sin(yBias * 25.0) * 0.5 + 0.5;
        if (weighted < 0.3) continue;  // filter by probability
        vec2 hitPos = vec2(SCREEN_X, yBias);

        float hd = length(p - hitPos);
        float lifeFade = 1.0 - hitPhase;
        float hitMask = exp(-hd * hd * 30000.0) * lifeFade;
        col += wp_pal(fract(seed * 0.02 + x_Time * 0.04)) * hitMask * 1.5;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
