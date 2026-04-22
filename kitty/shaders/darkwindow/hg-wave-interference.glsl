// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — 5 wave sources with circular interference, palette-cycled nodes

const int   SOURCES    = 5;
const float FREQ       = 12.0;     // wave cycles per normalized unit
const float WAVE_SPEED = 1.3;
const float DECAY      = 0.45;     // attenuation with distance
const float INTENSITY  = 0.55;

vec3 wi_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92); // cyan
    vec3 b = vec3(0.90, 0.20, 0.55); // magenta
    vec3 c = vec3(0.55, 0.30, 0.98); // violet
    vec3 d = vec3(0.96, 0.70, 0.25); // gold
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float wi_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  wi_hash2(float n) { return vec2(wi_hash(n), wi_hash(n * 1.37 + 11.0)); }

// Animated wave source
vec2 sourcePos(int i, float t) {
    float fi = float(i);
    vec2 base = wi_hash2(fi * 3.71) * 1.5 - 0.75;       // [-0.75, 0.75]
    base *= vec2(x_WindowSize.x / x_WindowSize.y, 1.0);
    // Lissajous drift
    base += 0.12 * vec2(sin(t * (0.3 + fi * 0.15) + fi),
                        cos(t * (0.25 + fi * 0.11) + fi * 1.3));
    return base;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Sum complex wave contributions (real + imaginary components for interference math)
    float wSumR = 0.0, wSumI = 0.0;
    vec3 col = vec3(0.0);
    for (int i = 0; i < SOURCES; i++) {
        float fi = float(i);
        vec2 sp = sourcePos(i, x_Time);
        float d = length(p - sp);
        // Wave: e^(i(kd - ωt)) / sqrt(d) with 1/√d falloff
        float phase = d * FREQ - x_Time * WAVE_SPEED - fi * 0.8;
        float amp = 1.0 / (0.5 + d * d * DECAY);
        wSumR += cos(phase) * amp;
        wSumI += sin(phase) * amp;

        // Source bright-point halo
        float sGlow = exp(-d * d * 40.0) * 0.9;
        vec3 sCol = wi_pal(fract(fi * 0.2 + x_Time * 0.04));
        col += sCol * sGlow;
    }

    // Interference magnitude → intensity; phase → hue shift
    float interMag = sqrt(wSumR * wSumR + wSumI * wSumI) * 0.3;
    float interPhase = atan(wSumI, wSumR) / 6.28318 + 0.5;  // [0,1]
    vec3 interCol = wi_pal(fract(interPhase + x_Time * 0.05));
    col += interCol * interMag * 0.6;

    // Crest sharpening — bright edges at constructive peaks
    float crest = smoothstep(0.7, 0.95, interMag);
    col += vec3(0.9, 0.95, 1.0) * crest * 0.3;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.7, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
