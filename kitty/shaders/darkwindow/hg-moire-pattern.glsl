// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Moire interference — 3 rotating layered patterns for rich beat visuals

const float LAYER_FREQ_1 = 40.0;
const float LAYER_FREQ_2 = 42.0;
const float LAYER_FREQ_3 = 38.0;
const float INTENSITY    = 0.5;

vec3 mp_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.90, 0.20, 0.55); // magenta
    vec3 c = vec3(0.55, 0.30, 0.98); // violet
    vec3 d = vec3(0.20, 0.95, 0.60); // mint
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Grid pattern at angle — returns brightness [0,1]
float grid(vec2 p, float freq, float angle) {
    float cr = cos(angle), sr = sin(angle);
    vec2 rp = mat2(cr, -sr, sr, cr) * p;
    float gx = 0.5 + 0.5 * sin(rp.x * freq);
    float gy = 0.5 + 0.5 * sin(rp.y * freq);
    return gx * gy;
}

// Concentric ring pattern
float rings(vec2 p, float freq, vec2 center) {
    float r = length(p - center);
    return 0.5 + 0.5 * sin(r * freq);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Three rotating / drifting layers
    float t = x_Time * 0.08;

    // Layer 1: grid at angle θ_1
    float layer1 = grid(p, LAYER_FREQ_1, t * 0.3);

    // Layer 2: grid at slightly different angle (Moire beats)
    vec2 offset2 = vec2(0.02 * sin(x_Time * 0.2), 0.02 * cos(x_Time * 0.15));
    float layer2 = grid(p + offset2, LAYER_FREQ_2, t * 0.3 + 0.05);

    // Layer 3: concentric rings drifting across
    vec2 ringCenter = 0.3 * vec2(cos(x_Time * 0.1), sin(x_Time * 0.13));
    float layer3 = rings(p, LAYER_FREQ_3, ringCenter);

    // Multiplicative interference — classic Moire
    float interference = layer1 * layer2;
    // Add ring layer additively
    float combined = interference * 0.7 + layer3 * 0.4;

    // Sharpen peaks to make beat fringes pop
    combined = pow(combined, 2.2);

    // Color cycles across beat bands
    float hue = fract(combined * 2.0 + x_Time * 0.06);
    vec3 col = mp_pal(hue) * combined * 1.3;

    // Add a second color based on ring layer alone
    col += mp_pal(fract(layer3 + x_Time * 0.08)) * layer3 * 0.4;

    // Vignette
    float r = length(p);
    float vignette = 1.0 - smoothstep(0.4, 1.0, r);
    col *= vignette;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.8, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
