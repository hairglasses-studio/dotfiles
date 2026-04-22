// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Slime mould / branching growth — Voronoi-style cells with fibrous connections

const int   SEEDS   = 20;
const float INTENSITY = 0.5;

vec3 sm_pal(float t) {
    vec3 a = vec3(0.20, 0.95, 0.60);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.25, 0.70);
    vec3 d = vec3(0.15, 0.75, 0.95);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float sm_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  sm_hash2(float n) { return vec2(sm_hash(n), sm_hash(n * 1.37 + 11.0)); }

vec2 seedPos(int i, float t) {
    float fi = float(i);
    vec2 base = sm_hash2(fi * 3.71) * 1.4 - 0.7;
    base.x *= x_WindowSize.x / x_WindowSize.y;
    // Slow wandering
    base += 0.05 * vec2(sin(t * 0.1 + fi), cos(t * 0.15 + fi * 1.3));
    return base;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Find nearest 2 seeds + their indices
    float d1 = 1e9, d2 = 1e9;
    int i1 = 0, i2 = 0;
    for (int i = 0; i < SEEDS; i++) {
        vec2 sp = seedPos(i, x_Time);
        float d = length(p - sp);
        if (d < d1) {
            d2 = d1; i2 = i1;
            d1 = d; i1 = i;
        } else if (d < d2) {
            d2 = d; i2 = i;
        }
    }

    // Voronoi edge distance
    float edgeDist = d2 - d1;
    // Filaments: where edgeDist is small, we're on a cell boundary
    float filament = smoothstep(0.03, 0.0, edgeDist);

    // Branch from each seed — bright connecting line to its nearest neighbor
    // Approx: if we're on the midpoint line between i1 and its nearest other, mark it bright
    vec3 col = vec3(0.02, 0.03, 0.05);

    // Pulsing growth from each seed (slow waves emanating outward)
    float pulseWave = 0.0;
    {
        float wavePhase = fract(x_Time * 0.2 + sm_hash(float(i1) * 7.3));
        float waveR = wavePhase * 0.5;
        float waveD = abs(d1 - waveR);
        pulseWave = exp(-waveD * waveD * 400.0);
    }

    vec3 cellCol = sm_pal(fract(float(i1) * 0.07 + x_Time * 0.03));

    // Cell interior: darker toward edge, brighter toward seed
    float interior = 1.0 - smoothstep(0.0, 0.3, d1);
    col += cellCol * interior * 0.1;

    // Filament/edge
    col += cellCol * filament * 0.9;
    col += vec3(1.0) * pulseWave * 0.6;

    // Seed nuclei — bright point at each
    for (int i = 0; i < SEEDS; i++) {
        vec2 sp = seedPos(i, x_Time);
        float sd = length(p - sp);
        vec3 sc = sm_pal(fract(float(i) * 0.07 + x_Time * 0.03));
        float core = exp(-sd * sd * 4000.0);
        float halo = exp(-sd * sd * 200.0) * 0.2;
        col += sc * (core * 1.4 + halo);
    }

    // Organic pulse — brightness fluctuates over whole colony
    col *= 0.8 + 0.2 * sin(x_Time * 0.5);

    // Fine grain noise for living-texture feel
    float grain = sm_hash(sm_hash(p.x * 50.0) + sm_hash(p.y * 50.0) + x_Time * 5.0);
    col *= 0.93 + grain * 0.07;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
