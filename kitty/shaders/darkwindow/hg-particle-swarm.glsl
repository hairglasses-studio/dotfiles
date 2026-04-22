// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Procedural particle swarm — 64 glowing particles on deterministic flocking paths

const int   PARTICLES   = 64;
const float PARTICLE_R  = 0.006;
const float TRAIL_LEN   = 0.03;
const int   TRAIL_TAPS  = 8;
const float INTENSITY   = 0.5;

vec3 ps_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.90, 0.20, 0.55); // magenta
    vec3 d = vec3(0.20, 0.95, 0.60); // mint
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float ps_hash(float n) { return fract(sin(n * 127.1) * 43758.5); }

// Procedural swarm motion: each particle orbits a slowly-drifting flock center
// with its own eccentricity and phase — gives the appearance of flocking
// without requiring per-particle state.
vec2 particlePath(int i, float t) {
    float fi = float(i);
    // Three swarms, each with their own drifting centroid
    float swarmIdx = floor(fi / float(PARTICLES / 3));
    float swarmSeed = swarmIdx * 13.7;
    vec2 center = vec2(
        sin(t * 0.18 + swarmSeed) * 0.5,
        cos(t * 0.14 + swarmSeed * 1.3) * 0.35
    );
    // Aspect correction
    center.x *= x_WindowSize.x / x_WindowSize.y;

    // Per-particle orbit params
    float seed = fi * 0.731 + swarmSeed;
    float orbitR = 0.08 + 0.18 * ps_hash(seed);
    float orbitSpeed = 0.8 + 1.2 * ps_hash(seed * 3.7);
    float phase = ps_hash(seed * 11.0) * 6.28318;

    // Eccentric orbit with occasional swaps (creates flocking-like rearrangement)
    float a = phase + t * orbitSpeed;
    vec2 offset = vec2(cos(a), sin(a * 1.1)) * orbitR;
    // Add high-freq jitter
    offset += 0.015 * vec2(sin(t * 5.0 + seed * 20.0), cos(t * 4.3 + seed * 15.0));

    return center + offset;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    // Each particle — head + trail
    for (int i = 0; i < PARTICLES; i++) {
        vec3 pc = ps_pal(fract(float(i) * 0.023 + x_Time * 0.04));

        // Motion-blurred trail: sample past positions
        for (int tIdx = 0; tIdx < TRAIL_TAPS; tIdx++) {
            float dt = float(tIdx) / float(TRAIL_TAPS) * TRAIL_LEN;
            vec2 pos = particlePath(i, x_Time - dt);
            float d = length(p - pos);
            float fade = 1.0 - float(tIdx) / float(TRAIL_TAPS);
            float r = PARTICLE_R * (0.7 + fade * 0.6);
            float k = exp(-d * d / (r * r) * 2.0);
            col += pc * k * fade * 0.3;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
