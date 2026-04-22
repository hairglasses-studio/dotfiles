// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Analog oscilloscope — XY plot of Lissajous/audio-like signals + phosphor glow + graticule

const int   WAVE_SAMPLES = 128;
const float GRID_SPACING = 0.1;
const float TRACE_WIDTH  = 0.004;
const float PHOSPHOR     = 0.35;
const float INTENSITY    = 0.55;

const vec3 SCOPE_GREEN = vec3(0.25, 0.95, 0.55);
const vec3 GRID_COL    = vec3(0.08, 0.30, 0.18);
const vec3 BG_COL      = vec3(0.01, 0.04, 0.02);

// Multi-channel waveform generation
vec2 waveform(float t, float samp) {
    // Two audio-like signals with harmonics — Lissajous figure
    float f1 = 1.3, f2 = 1.7;
    float ph1 = t * 0.4;
    float ph2 = t * 0.5 + 0.7;
    float x = sin(samp * 6.28 * f1 + ph1) + 0.3 * sin(samp * 6.28 * f1 * 3.0 + ph1 * 1.7);
    float y = sin(samp * 6.28 * f2 + ph2) + 0.25 * cos(samp * 6.28 * f2 * 5.0 + ph2 * 2.3);
    // Modulate shape over time for varied figures
    float morph = 0.5 + 0.5 * sin(t * 0.15);
    x = mix(x, x * cos(samp * 3.0 + ph1), morph * 0.5);
    y = mix(y, y * cos(samp * 2.3 + ph2), morph * 0.5);
    return vec2(x, y) * 0.4;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = BG_COL;

    // Graticule grid
    vec2 gridCoord = p / GRID_SPACING;
    vec2 gridDist = abs(fract(gridCoord + 0.5) - 0.5);
    float minor = smoothstep(0.01, 0.0, min(gridDist.x, gridDist.y));
    // Major every 5 units
    float majorInt = step(abs(mod(gridCoord.x, 5.0)), 0.5) * 0.4;
    majorInt += step(abs(mod(gridCoord.y, 5.0)), 0.5) * 0.4;
    float grid = minor + majorInt * minor;
    col += GRID_COL * grid * 1.2;

    // Center crosshairs (brighter)
    float hAxis = smoothstep(0.006, 0.0, abs(p.y));
    float vAxis = smoothstep(0.006, 0.0, abs(p.x));
    col += GRID_COL * (hAxis + vAxis) * 1.5;

    // Draw the waveform trace — sample N points, find minimum distance to trace
    float minDist = 1e9;
    float closestSamp = 0.0;
    for (int i = 0; i < WAVE_SAMPLES - 1; i++) {
        float s1 = float(i) / float(WAVE_SAMPLES);
        float s2 = float(i + 1) / float(WAVE_SAMPLES);
        vec2 w1 = waveform(x_Time, s1);
        vec2 w2 = waveform(x_Time, s2);
        // Segment distance
        vec2 pa = p - w1;
        vec2 ba = w2 - w1;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        float d = length(pa - ba * h);
        if (d < minDist) {
            minDist = d;
            closestSamp = mix(s1, s2, h);
        }
    }

    // Bright trace core
    float trace = smoothstep(TRACE_WIDTH, 0.0, minDist);
    // Phosphor glow (wider, softer)
    float glow = exp(-minDist * minDist * 800.0) * PHOSPHOR;
    col += SCOPE_GREEN * (trace * 1.5 + glow);

    // Flicker noise
    float flicker = fract(sin(x_PixelPos.y * 0.5 + x_Time * 60.0) * 43758.5);
    col *= 0.94 + flicker * 0.08;

    // Horizontal CRT scanlines
    float scan = 0.9 + 0.1 * sin(x_PixelPos.y * 1.5);
    col *= scan;

    // Bloom around screen edges
    float edgeDist = min(abs(p.x), abs(p.y));
    col *= smoothstep(0.0, 0.05, 0.5 - edgeDist);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.55);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
