// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Procedural spectrum analyzer — 48 bars with peaks, gradients, reflection

const int   BARS        = 48;
const float BASELINE    = 0.08;
const float MAX_HEIGHT  = 0.4;
const float INTENSITY   = 0.55;

vec3 sa_bar_col(float height, float barIdx) {
    float t = barIdx / float(BARS);
    vec3 low  = vec3(0.10, 0.50, 0.95);  // blue-cyan low bars
    vec3 mid  = vec3(0.20, 0.95, 0.60);  // green mid
    vec3 high = vec3(1.00, 0.80, 0.20);  // yellow high
    vec3 peak = vec3(1.00, 0.25, 0.35);  // red peak
    float heightBlend = height / MAX_HEIGHT;
    vec3 col;
    if (heightBlend < 0.33)      col = mix(low, mid, heightBlend * 3.0);
    else if (heightBlend < 0.66) col = mix(mid, high, (heightBlend - 0.33) * 3.0);
    else                         col = mix(high, peak, (heightBlend - 0.66) * 3.0);
    return col;
}

float sa_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Synthetic "frequency response" — low bars tend low, high bars drop
float barHeight(float barIdx, float t) {
    float fi = barIdx / float(BARS);
    // Base amplitude profile: log-falling toward high freq
    float base = pow(1.0 - fi, 0.7) * 0.6;
    // Multiple oscillating tones with different rates — creates music-like dynamics
    float v = 0.0;
    v += 0.4 * sin(t * 2.0 + fi * 3.0);
    v += 0.3 * sin(t * 3.7 + fi * 5.0);
    v += 0.25 * sin(t * 1.3 + fi * 0.7);
    v += 0.2 * sin(t * 7.0 + fi * 2.3);
    // Beat pattern (per-bar random offset for independent motion)
    v += 0.25 * sin(t * 4.0 + sa_hash(barIdx) * 6.28);
    float height = base + v * 0.2;
    return clamp(height, 0.02, MAX_HEIGHT);
}

// Peak hold — slowly-falling marker on top of each bar
float peakHeight(float barIdx, float t) {
    // Approximate max over recent time window via decaying average
    float current = barHeight(barIdx, t);
    float past = max(barHeight(barIdx, t - 0.1), barHeight(barIdx, t - 0.3));
    past = max(past, barHeight(barIdx, t - 0.5));
    return max(current, past * 0.95);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec3 bg = vec3(0.01, 0.02, 0.05);
    vec3 col = bg;

    float aspect = x_WindowSize.x / x_WindowSize.y;
    float barSpacing = 1.0 / float(BARS);
    float barWidthInGap = barSpacing * 0.7;

    float barIdxF = uv.x * float(BARS);
    float barIdx = floor(barIdxF);
    float barFrac = fract(barIdxF);

    // Only render within bar width (with gap)
    if (barFrac < 0.7) {
        float h = barHeight(barIdx, x_Time);
        float p = peakHeight(barIdx, x_Time);

        // Main bar — rises from baseline
        float barY = uv.y - BASELINE;
        if (barY > 0.0 && barY < h) {
            float heightBlend = barY / MAX_HEIGHT;
            vec3 barCol = sa_bar_col(barY, barIdx);
            col = barCol;
            // Vertical gradient (brighter top)
            col *= 0.5 + 0.5 * (barY / h);
        }

        // Peak indicator line
        float peakY = BASELINE + p;
        if (abs(uv.y - peakY) < 0.004) {
            col = sa_bar_col(p, barIdx) * 1.6;
        }

        // Reflection below baseline
        float reflY = BASELINE - (uv.y - BASELINE) * 2.0 * -1.0 + BASELINE;
        // Actually simpler: uv.y < BASELINE, reflect around BASELINE
        if (uv.y < BASELINE && uv.y > 0.0) {
            float reflD = BASELINE - uv.y;
            if (reflD < h) {
                float fade = 1.0 - reflD / h;
                vec3 reflCol = sa_bar_col(reflD, barIdx);
                col = mix(bg, reflCol * fade * 0.25, barFrac < 0.7 ? 1.0 : 0.0);
            }
        }
    }

    // Grid baseline
    if (abs(uv.y - BASELINE) < 0.002) {
        col = vec3(0.10, 0.40, 0.50);
    }

    // Frequency labels (major gridlines every 8 bars)
    for (int k = 1; k < 6; k++) {
        float gridX = float(k) / 6.0;
        if (abs(uv.x - gridX) < 0.001 && uv.y > BASELINE * 0.5 && uv.y < BASELINE + MAX_HEIGHT) {
            col += vec3(0.2, 0.4, 0.5) * 0.5;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
