// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Ripple tank — 2 wave sources showing interference pattern with moving crests

const float WAVE_FREQ = 60.0;
const float INTENSITY = 0.55;

vec3 rt_pal(float t) {
    vec3 blue_deep = vec3(0.05, 0.20, 0.45);
    vec3 cyan = vec3(0.20, 0.80, 0.95);
    vec3 white = vec3(0.95, 0.98, 1.00);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    if (t < 0.33)      return mix(blue_deep, cyan, t * 3.0);
    else if (t < 0.66) return mix(cyan, white, (t - 0.33) * 3.0);
    else               return mix(white, vio, (t - 0.66) * 3.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec2 src1 = vec2(-0.25 + 0.05 * sin(x_Time * 0.2), 0.0);
    vec2 src2 = vec2(0.25 + 0.05 * cos(x_Time * 0.17), 0.0);

    float d1 = length(p - src1);
    float d2 = length(p - src2);

    // Wave interference
    float w1 = sin(d1 * WAVE_FREQ - x_Time * 4.0) / (0.5 + d1 * 3.0);
    float w2 = sin(d2 * WAVE_FREQ - x_Time * 4.0) / (0.5 + d2 * 3.0);
    float wave = w1 + w2;

    // Heights → color
    float waveNorm = wave * 0.5 + 0.5;
    vec3 col = rt_pal(clamp(waveNorm, 0.0, 1.0));

    // Crest highlight (where wave is at high positive value)
    float crest = smoothstep(0.5, 0.9, wave);
    col += vec3(0.9, 0.95, 1.0) * crest * 0.6;

    // Trough darken
    float trough = smoothstep(-0.5, -0.9, wave);
    col *= 1.0 - trough * 0.4;

    // Source dots
    float s1D = length(p - src1);
    float s2D = length(p - src2);
    col += vec3(1.0, 0.9, 0.7) * exp(-s1D * s1D * 3000.0) * 1.2;
    col += vec3(1.0, 0.9, 0.7) * exp(-s2D * s2D * 3000.0) * 1.2;

    // Tank bottom texture — subtle FBM-like ripples
    float bottom = sin(p.x * 30.0 + x_Time * 0.3) * sin(p.y * 30.0 + x_Time * 0.3) * 0.03;
    col += rt_pal(0.2) * bottom;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.5);
    vec3 result = mix(terminal.rgb, col, visibility * 0.85);

    _wShaderOut = vec4(result, 1.0);
}
