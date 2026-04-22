// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Quantum foam — subatomic vacuum bubbles popping in/out of existence at Planck scale

const int   BUBBLES = 120;
const float INTENSITY = 0.5;

vec3 qf_pal(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(0.95, 0.25, 0.70);
    vec3 d = vec3(0.96, 0.85, 0.35);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float qf_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  qf_hash2(float n) { return vec2(qf_hash(n), qf_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    for (int i = 0; i < BUBBLES; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        // Short lifetime — pops in and out
        float lifetime = 0.15 + qf_hash(seed * 2.0) * 0.2;
        float interCycle = lifetime * (3.0 + qf_hash(seed * 3.0) * 5.0);
        float phase = mod(x_Time + qf_hash(seed) * interCycle, interCycle);

        // Only visible during its lifetime
        if (phase > lifetime) continue;

        float lifeT = phase / lifetime;
        // Position resampled each cycle
        float cycleID = floor((x_Time + qf_hash(seed) * interCycle) / interCycle);
        vec2 pos = qf_hash2(seed + cycleID * 13.7) * 2.0 - 1.0;
        pos.x *= x_WindowSize.x / x_WindowSize.y;

        // Size grows then shrinks (fluctuation)
        float sizeEnv = sin(lifeT * 3.14);
        float size = 0.006 + 0.012 * sizeEnv * qf_hash(seed * 5.1);

        float d = length(p - pos);
        if (d > size * 3.0) continue;

        // Anti-particle pair — two bubbles close together
        vec2 pairDir = vec2(cos(seed * 3.17), sin(seed * 3.17)) * size * 0.5;
        vec2 pos2 = pos - pairDir;
        pos = pos + pairDir;

        float d1 = length(p - pos);
        float d2 = length(p - pos2);

        // Each particle — bright core + halo
        float core1 = exp(-d1 * d1 / (size * 0.3 * size * 0.3) * 2.0);
        float core2 = exp(-d2 * d2 / (size * 0.3 * size * 0.3) * 2.0);
        float halo1 = exp(-d1 * d1 * 2000.0) * 0.2;
        float halo2 = exp(-d2 * d2 * 2000.0) * 0.2;

        vec3 bc = qf_pal(fract(seed * 0.03 + x_Time * 0.04));
        vec3 antiBc = qf_pal(fract(seed * 0.03 + 0.5 + x_Time * 0.04));

        col += bc * (core1 * sizeEnv * 1.5 + halo1);
        col += antiBc * (core2 * sizeEnv * 1.5 + halo2);

        // Annihilation flash at end of life
        if (lifeT > 0.8) {
            vec2 midPoint = (pos + pos2) * 0.5;
            float md = length(p - midPoint);
            float flash = exp(-md * md * 5000.0) * ((lifeT - 0.8) / 0.2);
            col += vec3(1.0, 0.95, 0.9) * flash * 0.6;
        }
    }

    // Very subtle background tint
    col += vec3(0.01, 0.005, 0.015);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
