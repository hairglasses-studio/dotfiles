// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Crystalline light caustics — interference patterns via trigonometric distortion

const int   CAUSTIC_LAYERS = 5;
const float SCALE  = 8.0;
const float INTENSITY = 0.5;
const float DISPLACE = 0.12;

vec3 cc_pal(float t) {
    vec3 a = vec3(0.15, 0.90, 0.96);  // bright cyan
    vec3 b = vec3(0.55, 0.30, 0.98);  // violet
    vec3 c = vec3(0.92, 0.70, 0.35);  // amber
    vec3 d = vec3(0.95, 0.25, 0.70);  // magenta
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Single-layer caustic — iterative trig displacement (cheap, dense detail)
float causticLayer(vec2 p, float t) {
    vec3 c = vec3(1.0, 1.0, 1.0);
    float inten = 0.005;
    for (int n = 0; n < 5; n++) {
        float fn = float(n + 1);
        p += vec2(
            0.6 / fn * sin(fn * p.y + t * 0.4 + 0.3 * fn),
            0.6 / fn * cos(fn * p.x + t * 0.4 + 0.3 * fn)
        );
    }
    return pow(inten / abs(sin(p.x - t) * sin(p.y + t)), 1.2);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos / x_WindowSize.y) * SCALE - vec2(4.0);

    vec3 col = vec3(0.0);
    for (int L = 0; L < CAUSTIC_LAYERS; L++) {
        float fL = float(L);
        // Each layer rotated/scaled for richer interference
        float a = fL * 0.7 + x_Time * 0.02;
        float cr = cos(a), sr = sin(a);
        vec2 lp = mat2(cr, -sr, sr, cr) * p * (1.0 + fL * 0.18);
        lp += DISPLACE * vec2(sin(x_Time * 0.3 + fL), cos(x_Time * 0.25 + fL * 1.3));
        float c = causticLayer(lp, x_Time + fL * 2.0);
        vec3 lc = cc_pal(fract(fL * 0.2 + x_Time * 0.04));
        col += lc * c * 0.15;
    }

    // Sharpen peaks (specular-like sparkle)
    col = pow(col, vec3(1.4));

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.7, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
