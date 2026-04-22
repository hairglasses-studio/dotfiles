// Shader attribution: hairglasses (original)
// Technique inspired by: Mandelbox (Tom Lowe) — spherical folding fractal.
// License: MIT
// (Cyberpunk — showcase/heavy) — 2D slice through an animated mandelbox fractal, 14 iterations, neon escape-time coloring

const int   MAX_ITER = 14;
const float SCALE    = 2.3;
const float FIX_R2   = 1.0;
const float MIN_R2   = 0.25;
const float INTENSITY = 0.55;

vec3 mb_pal(float t) {
    vec3 a = vec3(0.02, 0.02, 0.12);  // deep navy
    vec3 b = vec3(0.55, 0.30, 0.98);  // violet
    vec3 c = vec3(0.90, 0.20, 0.55);  // magenta
    vec3 d = vec3(0.96, 0.70, 0.25);  // gold
    vec3 e = vec3(0.10, 0.82, 0.92);  // cyan edge
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

// Mandelbox fold operations
vec3 mb_boxFold(vec3 p) {
    return clamp(p, -1.0, 1.0) * 2.0 - p;
}

vec3 mb_sphereFold(vec3 p) {
    float r2 = dot(p, p);
    if (r2 < MIN_R2)       p *= FIX_R2 / MIN_R2;
    else if (r2 < FIX_R2)  p *= FIX_R2 / r2;
    return p;
}

// Signed distance estimator for mandelbox
float mb_de(vec3 p, float scale) {
    vec3 offset = p;
    float dr = 1.0;
    for (int i = 0; i < MAX_ITER; i++) {
        p = mb_boxFold(p);
        p = mb_sphereFold(p);
        p = p * scale + offset;
        dr = dr * abs(scale) + 1.0;
    }
    float r = length(p);
    return r / abs(dr);
}

// 2D slice at varying z — animates depth through the fractal
void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y * 3.0;

    // Slow breathing + rotation
    float t = x_Time * 0.08;
    float cr = cos(t * 0.5), sr = sin(t * 0.5);
    p = mat2(cr, -sr, sr, cr) * p;

    // Z coordinate animates — flythrough depth
    float z = sin(x_Time * 0.12) * 1.5;
    float sc = SCALE + 0.3 * sin(x_Time * 0.07);

    // Iteration count — approximate escape-time via distance estimator
    float d = mb_de(vec3(p, z), sc);

    // Nearness to the fractal: bright where we're close to the surface
    float edge = 1.0 - smoothstep(0.0, 0.1, d);
    float shell = exp(-d * 40.0) * 0.6;

    // Color from distance + angle + time
    float hue = fract(log(d + 0.001) * 0.2 + x_Time * 0.04);
    vec3 col = mb_pal(hue) * (edge + shell);

    // Inner glow — deep violet haze when inside the set
    float interior = smoothstep(0.01, 0.0, d);
    col += vec3(0.25, 0.05, 0.4) * interior * 0.8;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
