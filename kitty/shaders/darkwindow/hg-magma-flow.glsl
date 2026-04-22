// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Magma flow — cracked black crust over bright molten rock with drift + ember particles

const int   OCTAVES = 6;
const int   EMBERS  = 30;
const float INTENSITY = 0.55;

vec3 mg_col(float heat) {
    vec3 black = vec3(0.04, 0.01, 0.02);
    vec3 dark  = vec3(0.35, 0.04, 0.02);
    vec3 red   = vec3(0.95, 0.20, 0.05);
    vec3 orange = vec3(1.00, 0.50, 0.10);
    vec3 yellow = vec3(1.00, 0.90, 0.40);
    vec3 white  = vec3(1.00, 0.98, 0.80);
    if (heat < 0.2)      return mix(black, dark, heat * 5.0);
    else if (heat < 0.4) return mix(dark, red, (heat - 0.2) * 5.0);
    else if (heat < 0.6) return mix(red, orange, (heat - 0.4) * 5.0);
    else if (heat < 0.8) return mix(orange, yellow, (heat - 0.6) * 5.0);
    else                 return mix(yellow, white, (heat - 0.8) * 5.0);
}

float mg_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float mg_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }

float mg_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(mg_hash(i), mg_hash(i + vec2(1,0)), u.x),
               mix(mg_hash(i + vec2(0,1)), mg_hash(i + vec2(1,1)), u.x), u.y);
}

float mg_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * mg_noise(p);
        p = p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = x_PixelPos / x_WindowSize.y;

    // Drift: magma scrolls slowly
    vec2 drifted = p + vec2(x_Time * 0.02, x_Time * 0.04);

    // Large-scale crust pattern
    float crust = mg_fbm(drifted * 2.0);
    // Smaller-scale cracks
    float cracks = mg_fbm(drifted * 10.0);

    // Heat beneath crust is high; cracks expose it
    // Low crust value = crack (molten visible); high = solid rock
    float heat = 1.0 - smoothstep(0.35, 0.55, crust);
    // Cracks bias further
    float crackDepth = smoothstep(0.55, 0.75, cracks) * (1.0 - heat);
    heat = max(heat, crackDepth);

    // Slow pulse
    heat *= 0.8 + 0.2 * sin(x_Time * 0.5);

    vec3 col = mg_col(heat);

    // Bright highlight along crack edges
    float edge = abs(crust - 0.45);
    float edgeMask = exp(-edge * edge * 300.0);
    col += mg_col(0.9) * edgeMask * 0.5;

    // Embers rising
    for (int i = 0; i < EMBERS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float fallSpeed = 0.15 + mg_hash1(seed) * 0.15;
        float phase = fract(x_Time * fallSpeed + mg_hash1(seed * 3.0));
        float baseX = (mg_hash1(seed * 5.1) - 0.5) * 2.4;
        baseX *= x_WindowSize.x / x_WindowSize.y;
        float wobble = 0.01 * sin(x_Time * 0.8 + seed * 10.0);
        vec2 ep = vec2(baseX + wobble, -0.3 + phase * 1.5);
        vec2 pC = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
        float ed = length(pC - ep);
        float emberHeat = 1.0 - phase;
        float core = exp(-ed * ed * 20000.0) * emberHeat;
        col += mg_col(0.7 + emberHeat * 0.3) * core * 1.3;
    }

    // Subtle radiant shimmer over whole view
    col *= 0.9 + 0.15 * mg_fbm(drifted * 30.0);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.55);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
