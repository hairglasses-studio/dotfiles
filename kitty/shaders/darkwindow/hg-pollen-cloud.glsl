// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Pollen cloud — drifting bright particle cloud with gentle breeze turbulence + warm light

const int   POLLEN = 180;
const int   OCTAVES = 4;
const float INTENSITY = 0.5;

vec3 pc_pal(float t) {
    vec3 warm = vec3(1.00, 0.80, 0.45);
    vec3 mint = vec3(0.40, 0.95, 0.65);
    vec3 pink = vec3(0.95, 0.55, 0.70);
    vec3 cyan = vec3(0.30, 0.85, 0.95);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(warm, mint, s);
    else if (s < 2.0) return mix(mint, pink, s - 1.0);
    else if (s < 3.0) return mix(pink, cyan, s - 2.0);
    else              return mix(cyan, warm, s - 3.0);
}

float pc_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float pc_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  pc_hash2(float n) { return vec2(pc_hash1(n), pc_hash1(n * 1.37 + 11.0)); }

float pc_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(pc_hash(i), pc_hash(i + vec2(1,0)), u.x),
               mix(pc_hash(i + vec2(0,1)), pc_hash(i + vec2(1,1)), u.x), u.y);
}

float pc_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * pc_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Warm soft backdrop
    vec3 bg = mix(vec3(0.08, 0.05, 0.10), vec3(0.15, 0.08, 0.04), smoothstep(0.3, -0.3, p.y));
    vec3 col = bg;

    // Breeze: FBM-driven wind vector field
    float breezePhase = x_Time * 0.15;
    vec2 wind = vec2(
        pc_fbm(p * 1.5 + vec2(breezePhase, 0.0)),
        pc_fbm(p * 1.5 + vec2(0.0, breezePhase))
    ) - 0.5;

    // Pollen particles
    for (int i = 0; i < POLLEN; i++) {
        float fi = float(i);
        float seed = fi * 3.71;
        vec2 basePos = pc_hash2(seed) * 2.4 - 1.2;
        basePos.x *= x_WindowSize.x / x_WindowSize.y;

        // Slow drift
        float driftPhase = x_Time * 0.08 * (0.5 + pc_hash1(seed * 3.7));
        basePos.x += 0.1 * sin(driftPhase + seed);
        basePos.y += 0.06 * cos(driftPhase * 0.7 + seed);

        // Wind displacement
        basePos += wind * 0.15 * (0.5 + pc_hash1(seed * 5.1));

        float d = length(p - basePos);
        if (d > 0.05) continue;

        // Twinkle
        float twinkle = 0.6 + 0.4 * sin(x_Time * (1.0 + pc_hash1(seed * 7.3) * 2.0) + fi);

        float size = 0.0025 + pc_hash1(seed * 11.0) * 0.003;
        float core = exp(-d * d / (size * size) * 2.0);
        float halo = exp(-d * d * 3000.0) * 0.2;

        vec3 pc = pc_pal(fract(seed * 0.02 + x_Time * 0.03));
        col += pc * (core * 1.5 + halo) * twinkle;
    }

    // Faint sunbeam from above-right
    vec2 sunDir = normalize(vec2(0.4, 0.7));
    float sunAlong = dot(p, sunDir);
    float sunPerp = abs(p.x * sunDir.y - p.y * sunDir.x);
    float beam = exp(-sunPerp * sunPerp * 20.0) * smoothstep(-0.5, 0.3, sunAlong);
    col += pc_pal(0.0) * beam * 0.1;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
