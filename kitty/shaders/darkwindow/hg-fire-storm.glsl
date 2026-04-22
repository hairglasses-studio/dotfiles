// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Fire storm — upward convection FBM flames with ember sparks + smoke

const int   EMBERS   = 32;
const float INTENSITY = 0.55;

vec3 fs_pal(float heat) {
    // Temperature-based fire palette
    vec3 black = vec3(0.02, 0.0, 0.0);
    vec3 dark  = vec3(0.35, 0.02, 0.08);
    vec3 mid   = vec3(0.95, 0.30, 0.05);
    vec3 hot   = vec3(1.00, 0.80, 0.20);
    vec3 bright = vec3(1.00, 0.95, 0.75);
    if (heat < 0.15)      return mix(black, dark, heat / 0.15);
    else if (heat < 0.45) return mix(dark, mid, (heat - 0.15) / 0.3);
    else if (heat < 0.75) return mix(mid, hot, (heat - 0.45) / 0.3);
    else                  return mix(hot, bright, min(1.0, (heat - 0.75) / 0.25));
}

float fs_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5);
}

float fs_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(fs_hash(i), fs_hash(i + vec2(1,0)), u.x),
               mix(fs_hash(i + vec2(0,1)), fs_hash(i + vec2(1,1)), u.x), u.y);
}

// 6-octave FBM with upward bias
float fs_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 6; i++) {
        v += a * fs_noise(p);
        p = p * 2.09 + vec2(0.13, -0.07);
        a *= 0.52;
    }
    return v;
}

float fs_hash1(float n) { return fract(sin(n * 127.1) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = x_PixelPos / x_WindowSize.y;
    float y = uv.y;

    // Upward convection — scroll noise upward + turbulence
    vec2 noiseP = p * vec2(2.0, 2.5);
    noiseP.y += x_Time * 1.2;
    // Horizontal disturbance based on FBM
    float disturb = fs_fbm(noiseP * 2.0 + x_Time * 0.4) - 0.5;
    noiseP.x += disturb * 0.35;

    // Primary heat field
    float heat = fs_fbm(noiseP);
    // Bias: hot near bottom, cooler toward top (flames rise)
    heat *= smoothstep(0.95, 0.25, y);
    // Narrow flame tongues: shape via smoothstep
    heat = smoothstep(0.35, 0.80, heat);
    // Flicker: high-frequency modulation
    heat *= 0.7 + 0.3 * fs_fbm(noiseP * 6.0 + x_Time * 3.0);

    vec3 col = fs_pal(heat);

    // Ember particles rising
    for (int i = 0; i < EMBERS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float emberX = fract(seed * 0.137) * (x_WindowSize.x / x_WindowSize.y);
        // Y cycles 0 → 1 over some period, some embers die early
        float fallSpeed = 0.25 + fs_hash1(seed) * 0.2;
        float phase = fract(x_Time * fallSpeed + fs_hash1(seed * 3.0));
        float emberY = 1.0 - phase;
        // Horizontal wobble
        emberX += 0.015 * sin(x_Time * 2.0 + seed * 10.0);
        vec2 ep = vec2(emberX, emberY);

        float ed = length(p - ep);
        float emberSize = 0.004 * (1.0 - phase * 0.6);
        float ember = exp(-ed * ed / (emberSize * emberSize) * 3.0);
        float emberHeat = 1.0 - phase * 0.7;   // cools as it rises
        col += fs_pal(emberHeat) * ember * 1.5;
        // Spark tail
        float tail = exp(-ed * ed * 2500.0) * smoothstep(0.0, 0.08, p.y - ep.y);
        col += fs_pal(emberHeat * 0.6) * tail * 0.3;
    }

    // Smoke rising above flames — dark grey FBM in upper region
    float smoke = fs_fbm(noiseP * 0.7 + vec2(0.0, x_Time * 0.8)) * smoothstep(0.55, 0.95, y);
    col = mix(col, vec3(0.08, 0.06, 0.08), smoke * 0.7);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
