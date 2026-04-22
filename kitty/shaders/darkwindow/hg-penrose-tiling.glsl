// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Penrose-style aperiodic tiling — 5-fold quasicrystal pattern via de Bruijn grid

const int   GRID_SETS = 5;
const float GRID_FREQ = 8.0;
const float LINE_WIDTH = 0.020;
const float INTENSITY = 0.55;

vec3 pt_pal(float t) {
    vec3 a = vec3(0.90, 0.30, 0.70);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(0.55, 0.30, 0.98);
    vec3 d = vec3(0.96, 0.85, 0.40);
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

    // 5-fold rotational symmetry — sum of 5 sinusoidal grids at angles 0, 72°, 144°, 216°, 288°
    // Quasicrystal pattern (de Bruijn / dual grid method)
    float t = x_Time * 0.1;
    vec3 col = vec3(0.0);
    float nearestLineD = 1e9;
    float nearestHue = 0.0;

    for (int i = 0; i < GRID_SETS; i++) {
        float fi = float(i);
        float angle = fi * 6.28318 / float(GRID_SETS) + t * 0.5;
        vec2 dir = vec2(cos(angle), sin(angle));
        // Distance from p to nearest line in this grid family
        float along = dot(p, dir);
        float gridVal = along * GRID_FREQ;
        float distToLine = abs(fract(gridVal + 0.5) - 0.5);
        if (distToLine < nearestLineD) {
            nearestLineD = distToLine;
            nearestHue = fi / float(GRID_SETS);
        }
        // Line brightness at this family
        float lineMask = smoothstep(LINE_WIDTH * GRID_FREQ, 0.0, distToLine);
        vec3 lc = pt_pal(fract(fi / float(GRID_SETS) + x_Time * 0.03));
        col += lc * lineMask * 0.4;
    }

    // Highlight intersections via sum-of-cosines (quasicrystal brightness)
    float qc = 0.0;
    for (int i = 0; i < GRID_SETS; i++) {
        float fi = float(i);
        float angle = fi * 6.28318 / float(GRID_SETS) + t * 0.5;
        vec2 dir = vec2(cos(angle), sin(angle));
        qc += cos(dot(p, dir) * GRID_FREQ);
    }
    qc /= float(GRID_SETS);

    // Bright lattice points where qc ≈ 1 (all lines intersect)
    float peaks = pow(max(0.0, qc + 0.4), 3.0);
    col += pt_pal(fract(x_Time * 0.05)) * peaks * 0.7;

    // Regions: qc sign + magnitude gives region colors (rhombus interiors)
    float regionHue = qc * 0.2 + 0.5;
    vec3 regionCol = pt_pal(fract(regionHue + x_Time * 0.03)) * 0.15;
    col += regionCol * (1.0 - pow(qc, 2.0));

    // Vignette
    float r = length(p);
    col *= smoothstep(1.4, 0.3, r);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.9, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
