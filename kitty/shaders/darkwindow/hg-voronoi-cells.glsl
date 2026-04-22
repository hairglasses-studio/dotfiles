// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk) — Voronoi cells with neon edges, slow chromatic drift

const float CELL_SCALE = 6.0;
const float EDGE_WIDTH = 0.025;
const float INTENSITY  = 0.4;
const float DRIFT      = 0.15;   // how much cell centers wobble

vec3 voronoiPalette(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);  // cyan
    vec3 b = vec3(0.90, 0.18, 0.60);  // magenta
    vec3 c = vec3(0.55, 0.30, 0.98);  // violet
    vec3 d = vec3(0.20, 0.95, 0.60);  // mint
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

vec2 vor_hash(vec2 p) {
    p = vec2(dot(p, vec2(127.1, 311.7)), dot(p, vec2(269.5, 183.3)));
    return fract(sin(p) * 43758.5453);
}

// Returns vec3(F1_dist, F2_dist, cellSeed)
vec3 voronoi(vec2 uv, float t) {
    vec2 iPos = floor(uv);
    vec2 fPos = fract(uv);
    float f1 = 8.0, f2 = 8.0;
    float cellSeed = 0.0;
    for (int y = -1; y <= 1; y++) {
        for (int x = -1; x <= 1; x++) {
            vec2 neighbor = vec2(float(x), float(y));
            vec2 h = vor_hash(iPos + neighbor);
            vec2 p = neighbor + 0.5 + DRIFT * vec2(sin(t + h.x * 6.28), cos(t + h.y * 6.28));
            float d = length(p - fPos);
            if (d < f1) {
                f2 = f1;
                f1 = d;
                cellSeed = fract(dot(iPos + neighbor, vec2(12.9898, 78.233)));
            } else if (d < f2) {
                f2 = d;
            }
        }
    }
    return vec3(f1, f2, cellSeed);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    vec3 vr = voronoi(p * CELL_SCALE, x_Time * 0.3);

    // Edge distance: smaller = closer to cell boundary
    float edge = vr.y - vr.x;
    float edgeMask = 1.0 - smoothstep(0.0, EDGE_WIDTH, edge);

    // Cell fill: each cell gets a color based on its seed, drifting over time
    vec3 cellColor = voronoiPalette(fract(vr.z + x_Time * 0.05));
    // Soft fall-off from center to edge
    float cellGlow = 0.15 * (1.0 - smoothstep(0.0, 0.3, vr.x));

    vec3 effect = cellColor * edgeMask + cellColor * cellGlow;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.7);
    vec3 result = mix(terminal.rgb, effect, visibility * clamp(length(effect), 0.0, 1.0));

    _wShaderOut = vec4(result, 1.0);
}
