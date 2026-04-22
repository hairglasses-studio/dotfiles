// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Ember field — 150 drifting glowing embers with fire-palette heat variation

const int   EMBERS = 150;
const float INTENSITY = 0.55;

vec3 ef_col(float heat) {
    vec3 dark   = vec3(0.25, 0.02, 0.04);
    vec3 red    = vec3(0.95, 0.22, 0.08);
    vec3 orange = vec3(1.00, 0.55, 0.20);
    vec3 yellow = vec3(1.00, 0.90, 0.45);
    vec3 white  = vec3(1.00, 0.98, 0.85);
    if (heat < 0.25)      return mix(dark, red, heat * 4.0);
    else if (heat < 0.5)  return mix(red, orange, (heat - 0.25) * 4.0);
    else if (heat < 0.75) return mix(orange, yellow, (heat - 0.5) * 4.0);
    else                  return mix(yellow, white, (heat - 0.75) * 4.0);
}

float ef_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  ef_hash2(float n) { return vec2(ef_hash(n), ef_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Dark warm backdrop — fire glow from below
    vec3 bg = mix(vec3(0.02, 0.005, 0.0), vec3(0.14, 0.03, 0.01), smoothstep(0.8, -0.4, p.y));
    vec3 col = bg;

    // Embers drifting upward
    for (int i = 0; i < EMBERS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        // Falls upward from below screen to above
        float fallSpeed = 0.08 + ef_hash(seed * 3.7) * 0.12;
        float cycle = 1.5 / fallSpeed;
        float phase = fract(x_Time * fallSpeed + ef_hash(seed) * 10.0);
        // Base X
        float baseX = (ef_hash(seed * 5.1) - 0.5) * 2.4 * x_WindowSize.x / x_WindowSize.y;
        // Horizontal wobble
        float wobblePhase = x_Time * 0.4 + seed * 2.0;
        float wobble = 0.015 * sin(wobblePhase) + 0.008 * sin(wobblePhase * 3.3);
        vec2 ep = vec2(baseX + wobble, -0.8 + phase * 1.8);

        float d = length(p - ep);
        if (d > 0.04) continue;
        // Heat decays as ember rises (cools)
        float heat = 1.0 - phase;
        heat = pow(heat, 1.3);
        // Size shrinks as ember cools
        float size = 0.003 * (0.6 + heat * 0.8);
        float core = exp(-d * d / (size * size) * 2.0);
        float halo = exp(-d * d * 1500.0) * 0.4 * heat;
        float outerHalo = exp(-d * d * 150.0) * 0.08 * heat;
        col += ef_col(heat) * (core * 1.4 + halo + outerHalo);
    }

    // Bottom glow from fire source (not rendered, implied)
    float bottomGlow = exp(-pow(p.y + 0.5, 2.0) * 3.0);
    col += ef_col(0.55) * bottomGlow * 0.15;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
