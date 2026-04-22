// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Sound wave visualization — multi-channel oscilloscope traces with FFT-like bars

const int   CHANNELS = 3;
const int   WAVE_SAMPLES = 200;
const float INTENSITY = 0.55;

vec3 sw_pal(float t) {
    vec3 a = vec3(0.20, 0.95, 0.55);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(0.95, 0.30, 0.65);
    vec3 d = vec3(0.96, 0.85, 0.35);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Procedural waveform with harmonics, per-channel signature
float waveform(float x, int channel, float t) {
    float f1 = 1.0 + float(channel) * 0.5;
    float f2 = 3.0 + float(channel) * 0.7;
    float f3 = 7.0 - float(channel) * 0.3;
    float v = 0.0;
    v += 0.5 * sin(x * 6.28 * f1 + t * f1);
    v += 0.3 * sin(x * 6.28 * f2 + t * f2 * 0.9);
    v += 0.2 * sin(x * 6.28 * f3 + t * f3 * 1.1);
    // Beat modulation
    v *= 0.5 + 0.5 * sin(t * 0.5 + float(channel) * 1.5);
    return v * 0.35;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.02, 0.05);

    // Horizontal grid lines
    for (int k = -3; k <= 3; k++) {
        float gy = float(k) * 0.12;
        float gd = abs(p.y - gy);
        col += vec3(0.05, 0.08, 0.12) * smoothstep(0.002, 0.0, gd);
    }

    // Vertical grid lines
    for (int k = -6; k <= 6; k++) {
        float gx = float(k) * 0.12;
        float gd = abs(p.x - gx);
        col += vec3(0.05, 0.08, 0.12) * smoothstep(0.002, 0.0, gd);
    }

    // Draw each channel's waveform (find minimum distance to the trace)
    for (int c = 0; c < CHANNELS; c++) {
        float chOffset = (float(c) - 1.0) * 0.25;   // vertical stack
        float minD = 1e9;
        // Sample waveform densely and find nearest segment
        for (int i = 0; i < WAVE_SAMPLES - 1; i++) {
            float x1 = float(i) / float(WAVE_SAMPLES - 1) * 2.0 - 1.0;
            float x2 = float(i + 1) / float(WAVE_SAMPLES - 1) * 2.0 - 1.0;
            x1 *= x_WindowSize.x / x_WindowSize.y;
            x2 *= x_WindowSize.x / x_WindowSize.y;
            float y1 = waveform(x1, c, x_Time) + chOffset;
            float y2 = waveform(x2, c, x_Time) + chOffset;
            vec2 a2 = vec2(x1, y1);
            vec2 b2 = vec2(x2, y2);
            vec2 pa = p - a2;
            vec2 ba = b2 - a2;
            float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
            float d = length(pa - ba * h);
            minD = min(minD, d);
        }

        vec3 waveCol = sw_pal(fract(float(c) * 0.3 + x_Time * 0.04));
        float core = smoothstep(0.005, 0.0, minD);
        float glow = exp(-minD * minD * 800.0) * 0.4;
        col += waveCol * (core * 1.3 + glow);
    }

    // Bottom: FFT-like bars (fake from waveform amplitude at various freqs)
    if (uv.y < 0.25 && uv.y > 0.02) {
        float barFreq = 30.0;
        float barX = floor(uv.x * barFreq);
        float barAmp = abs(waveform(barX / barFreq * 2.0 - 1.0, 0, x_Time)) * 3.0;
        float barH = 0.01 + barAmp * 0.2;
        if (uv.y < 0.02 + barH) {
            vec3 barCol = sw_pal(fract(barX / barFreq + x_Time * 0.02));
            float yInBar = (uv.y - 0.02) / barH;
            col += barCol * (0.4 + yInBar * 0.6) * 0.5;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
