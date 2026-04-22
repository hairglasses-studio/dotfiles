// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Light caustics through water — refractive pattern + surface ripples + deep blue gradient

const int   CAUSTIC_LAYERS = 6;
const float INTENSITY      = 0.5;

vec3 cw_pal(float t) {
    vec3 deep = vec3(0.02, 0.08, 0.22);
    vec3 mid  = vec3(0.15, 0.55, 0.85);
    vec3 edge = vec3(0.65, 0.95, 1.00);
    vec3 hot  = vec3(1.00, 0.95, 0.65);
    if (t < 0.33)      return mix(deep, mid, t * 3.0);
    else if (t < 0.66) return mix(mid, edge, (t - 0.33) * 3.0);
    else               return mix(edge, hot, (t - 0.66) * 3.0);
}

// Caustic ripple layer — iterative trigonometric fold (Elias Hasle / moni-dz technique)
float causticLayer(vec2 p, float t) {
    float inten = 0.006;
    for (int n = 0; n < 5; n++) {
        float fn = float(n + 1);
        p += vec2(
            0.7 / fn * sin(fn * p.y + t * 0.5 + 0.3 * fn),
            0.7 / fn * cos(fn * p.x + t * 0.5 + 0.3 * fn)
        );
    }
    return pow(inten / abs(sin(p.x - t * 0.8) * sin(p.y + t * 0.6)), 1.3);
}

// Surface ripple — shimmer layer
float surfaceRipple(vec2 p, float t) {
    float v = 0.0;
    for (int i = 1; i < 5; i++) {
        float fi = float(i);
        v += sin(dot(p, vec2(cos(fi * 1.3), sin(fi * 1.7))) * fi * 8.0 + t * fi * 0.4) / fi;
    }
    return v * 0.5 + 0.5;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = x_PixelPos / x_WindowSize.y * 5.0;

    // Multi-layer caustics with rotation + scale variation
    vec3 causticCol = vec3(0.0);
    for (int L = 0; L < CAUSTIC_LAYERS; L++) {
        float fL = float(L);
        float a = fL * 0.5 + x_Time * 0.03;
        float cr = cos(a), sr = sin(a);
        vec2 lp = mat2(cr, -sr, sr, cr) * p * (1.0 + fL * 0.2);
        float c = causticLayer(lp, x_Time + fL * 1.7);
        vec3 lc = cw_pal(clamp(c * 0.3 + fL * 0.1, 0.0, 1.0));
        causticCol += lc * c * 0.12;
    }

    // Surface ripple modulates caustic brightness
    float surfMod = surfaceRipple(uv * 6.0, x_Time);
    causticCol *= 0.4 + surfMod * 0.9;

    // Depth gradient: darker bottom, lighter top
    vec3 depthCol = mix(cw_pal(0.0), cw_pal(0.4), uv.y);

    // Base water color
    vec3 col = depthCol + causticCol;

    // Brighter highlights at caustic peaks
    float peaks = pow(length(causticCol), 2.5);
    col += vec3(1.0, 0.95, 0.8) * peaks * 0.3;

    // Surface wave line
    float waveLine = abs(uv.y - (0.75 + 0.015 * sin(uv.x * 20.0 + x_Time * 2.0)));
    float waveMask = exp(-waveLine * 500.0);
    col += vec3(0.6, 0.85, 1.0) * waveMask * 0.3;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.55);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
