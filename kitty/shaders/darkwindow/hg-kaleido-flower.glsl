// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Kaleidoscope flower — 12-fold mirror fold with nested floral petal SDFs

const int   FOLDS = 12;
const int   PETAL_RINGS = 4;
const float INTENSITY = 0.55;

vec3 kf_pal(float t) {
    vec3 a = vec3(0.95, 0.30, 0.70);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.10, 0.82, 0.92);
    vec3 d = vec3(0.20, 0.95, 0.60);
    vec3 e = vec3(0.96, 0.85, 0.35);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

float kf_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    float r = length(p);
    float a = atan(p.y, p.x) + x_Time * 0.1;

    // 12-fold mirror fold
    float wedge = 6.28318 / float(FOLDS);
    float fa = mod(a, wedge);
    fa = abs(fa - wedge * 0.5);
    vec2 wp = vec2(cos(fa), sin(fa)) * r;

    vec3 col = vec3(0.01, 0.01, 0.03);

    // Petal rings at different radii
    for (int L = 0; L < PETAL_RINGS; L++) {
        float fL = float(L);
        float layerR = 0.08 + fL * 0.09;
        float petalW = 0.03 + fL * 0.008;
        float rDiff = abs(r - layerR);
        if (rDiff > petalW * 2.0) continue;

        // Petal shape — elongated in a direction
        // petalAng is angle within fold
        float petalAng = atan(wp.y, wp.x);
        float petalShape = sin(petalAng * (2.0 + fL));   // lobes
        float petalBoundary = layerR + petalShape * petalW * 0.5 - r;
        float petalMask = smoothstep(0.005, 0.0, abs(petalBoundary));
        float petalFill = smoothstep(-0.01, 0.0, petalBoundary);

        // Color per layer + time drift
        vec3 lc = kf_pal(fract(fL * 0.15 + petalAng * 0.1 + x_Time * 0.04));
        col = mix(col, lc * 0.5, petalFill * 0.7);
        col += lc * petalMask * 0.8;
    }

    // Central bright core with flower-star shape
    if (r < 0.08) {
        float starAng = atan(wp.y, wp.x);
        float starShape = 0.04 + 0.02 * cos(starAng * 8.0 + x_Time * 0.5);
        float starMask = smoothstep(starShape + 0.005, starShape - 0.005, r);
        col += kf_pal(fract(x_Time * 0.06)) * starMask * 1.3;
    }

    // Outer ring glow
    float ringD = abs(r - 0.45);
    col += kf_pal(fract(x_Time * 0.05)) * exp(-ringD * ringD * 1500.0) * 0.4;

    // Vignette
    col *= smoothstep(1.2, 0.3, r);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
