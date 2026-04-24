// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Heatmap grid — 32x12 rolling activity heatmap with cell intensities driven by time-shifted FBM + periodic spikes; cells flash bright when spiking; column separator lines, axis hint strip on left

const int   COLS = 32;
const int   ROWS = 12;
const float INTENSITY = 0.55;

vec3 hm_heat(float v) {
    // v ∈ [0, 1]
    vec3 cold    = vec3(0.02, 0.04, 0.15);
    vec3 blue    = vec3(0.15, 0.35, 0.75);
    vec3 teal    = vec3(0.25, 0.70, 0.85);
    vec3 amber   = vec3(1.00, 0.70, 0.25);
    vec3 red     = vec3(1.00, 0.28, 0.15);
    vec3 white   = vec3(1.00, 0.95, 0.80);
    if (v < 0.2)      return mix(cold, blue, v * 5.0);
    else if (v < 0.4) return mix(blue, teal, (v - 0.2) * 5.0);
    else if (v < 0.6) return mix(teal, amber, (v - 0.4) * 5.0);
    else if (v < 0.8) return mix(amber, red, (v - 0.6) * 5.0);
    else              return mix(red, white, (v - 0.8) * 5.0);
}

float hm_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float hm_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(hm_hash(i), hm_hash(i + vec2(1, 0)), u.x),
               mix(hm_hash(i + vec2(0, 1)), hm_hash(i + vec2(1, 1)), u.x), u.y);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.008, 0.020);

    // Grid area: occupies most of frame, with small margin left for "axis"
    float axisW = 0.08;
    // Map p.x ∈ [-1+axisW, 1] to col index [0, COLS-1]
    if (p.x > -1.0 + axisW && p.x < 1.0 && p.y > -0.5 && p.y < 0.5) {
        float gridX = (p.x - (-1.0 + axisW)) / (2.0 - axisW);
        float gridY = (p.y + 0.5) / 1.0;
        float scrollX = gridX - x_Time * 0.04;
        float cellX = floor(scrollX * float(COLS));
        float cellY = floor(gridY * float(ROWS));

        // Activity value: sum of FBM-driven temporal waves
        vec2 cellPos = vec2(cellX, cellY);
        float activity = hm_noise(cellPos * 0.25 + vec2(x_Time * 0.1, 0.0));
        activity = (activity - 0.2) * 1.5;
        // Row-specific modulation (some rows are hot, some cold)
        float rowBase = 0.15 + sin(cellY * 0.7) * 0.15;
        activity = clamp(activity + rowBase, 0.0, 1.0);

        // Spike events: occasional cells spike to max
        float spikeSeed = hm_hash(cellPos + vec2(floor(x_Time * 0.3), 0.0));
        if (spikeSeed > 0.96) {
            activity = max(activity, 0.92);
        }

        // Cell boundaries (slight dark border)
        vec2 cellFrac = fract(scrollX * float(COLS) * vec2(1.0, 1.0));  // just x
        vec2 gridFrac = vec2(fract(scrollX * float(COLS)), fract(gridY * float(ROWS)));
        float borderMask = 0.0;
        if (gridFrac.x < 0.02 || gridFrac.x > 0.98) borderMask = 1.0;
        if (gridFrac.y < 0.04 || gridFrac.y > 0.96) borderMask = 1.0;

        vec3 cellCol = hm_heat(activity);
        col = cellCol;
        // Slight interior gradient
        float ci = 1.0 - smoothstep(0.0, 0.5, abs(gridFrac.y - 0.5) * 1.5);
        col *= 0.85 + ci * 0.2;
        // Border
        col *= 1.0 - borderMask * 0.45;
    }

    // === Left-edge axis hint: vertical bars per row in muted palette ===
    if (p.x < -1.0 + axisW && p.x > -1.05 && p.y > -0.5 && p.y < 0.5) {
        float gridY = (p.y + 0.5) / 1.0;
        float cellY = floor(gridY * float(ROWS));
        vec3 axisCol = hm_heat(0.1 + cellY * 0.05);
        col = axisCol * 0.65;
        // Row separator
        float rowSep = fract(gridY * float(ROWS));
        if (rowSep < 0.05 || rowSep > 0.95) col *= 0.4;
    }

    // === Right-edge "now" indicator ===
    if (p.x > 0.98 && p.y > -0.5 && p.y < 0.5) {
        float pulse = 0.7 + 0.3 * sin(x_Time * 4.0);
        col += vec3(1.0, 0.95, 0.80) * (p.x - 0.98) * 50.0 * pulse;
    }

    // === Top/bottom axis labels (hint only — faint lines) ===
    if (abs(p.y) > 0.5 && abs(p.y) < 0.52) {
        col += vec3(0.20, 0.25, 0.35) * 0.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
