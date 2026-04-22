// Shader attribution: hairglasses (original)
// Technique inspired by: Shane's Apollonian tunnel (Shadertoy t3ffRN) — inversive circle packing.
// License: MIT
// (Cyberpunk — showcase/heavy) — 2D Apollonian gasket with neon glow, 12 inversion iterations

const int   INV_ITER = 12;
const float ZOOM     = 1.8;
const float GLOW     = 0.02;
const float INTENSITY = 0.5;

vec3 ap_pal(float t) {
    vec3 a = vec3(0.90, 0.18, 0.60); // magenta
    vec3 b = vec3(0.10, 0.82, 0.92); // cyan
    vec3 c = vec3(0.55, 0.30, 0.98); // violet
    vec3 d = vec3(0.96, 0.70, 0.25); // gold
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    p *= ZOOM;

    // Slow breathing + rotation
    float t = x_Time * 0.12;
    float cr = cos(t * 0.5), sr = sin(t * 0.5);
    p = mat2(cr, -sr, sr, cr) * p;
    p *= 1.0 + 0.1 * sin(t);

    // Apollonian — iterated inversion & translation
    float scale = 1.0;
    float minDist = 1e9;
    vec2 pt = p;
    for (int i = 0; i < INV_ITER; i++) {
        // Fold into positive quadrant then translate
        pt = -1.0 + 2.0 * fract(0.5 * pt + 0.5);
        // Inversive step
        float r2 = dot(pt, pt);
        float k = 1.3 / r2;
        pt *= k;
        scale *= k;
        // Track nearest circle edge
        minDist = min(minDist, abs(r2 - 1.0));
    }

    // Distance to nearest Apollonian edge (post-iteration)
    float d = minDist / scale;

    // Main edge glow
    float edge = smoothstep(GLOW * 2.0, 0.0, d);
    float halo = exp(-d * 60.0) * 0.6;

    // Color from iteration count + time
    vec3 col = ap_pal(fract(scale * 0.003 + x_Time * 0.04)) * edge;
    col += ap_pal(fract(x_Time * 0.07)) * halo * 0.5;

    // Secondary rings — subtle concentric
    float rings = 0.5 + 0.5 * sin(length(p) * 8.0 - x_Time * 0.6);
    col += vec3(0.1, 0.2, 0.3) * smoothstep(0.95, 1.0, rings) * 0.15;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
