// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Ambient shimmer dust — 256 chromatic motes with parallax + twinkle

const int   MOTES       = 256;
const float INTENSITY   = 0.5;
const float MOTE_SIZE   = 0.002;

vec3 sd_pal(float t) {
    vec3 a = vec3(0.15, 0.90, 0.95); // cyan
    vec3 b = vec3(0.55, 0.30, 0.98); // violet
    vec3 c = vec3(0.96, 0.85, 0.45); // warm gold
    vec3 d = vec3(0.95, 0.30, 0.70); // pink
    vec3 e = vec3(0.25, 0.95, 0.60); // mint
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

float sd_hash(float n) { return fract(sin(n * 127.1) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    // Distribute motes across screen with parallax layers
    for (int i = 0; i < MOTES; i++) {
        float fi = float(i);
        float depthLayer = fract(fi * 0.137);  // [0,1] parallax depth

        // Grid-based position (avoid clustering)
        float gridN = 16.0;  // 16x16 grid, 256 cells
        float cx = mod(fi, gridN);
        float cy = floor(fi / gridN);
        vec2 basePos = vec2(cx + sd_hash(fi * 3.1),
                            cy + sd_hash(fi * 7.3)) / gridN * 2.0 - 1.0;
        basePos.x *= x_WindowSize.x / x_WindowSize.y;

        // Slow drift — each mote orbits slightly
        float orbitR = 0.02 + 0.03 * depthLayer;
        float orbitSpeed = 0.3 + sd_hash(fi * 5.7) * 0.3;
        float orbitPhase = sd_hash(fi * 11.3) * 6.28;
        vec2 motePos = basePos + orbitR * vec2(
            cos(x_Time * orbitSpeed + orbitPhase),
            sin(x_Time * orbitSpeed * 0.87 + orbitPhase * 1.3)
        );

        // Global drift (wind)
        motePos.x += 0.02 * sin(x_Time * 0.1 + fi * 0.05);

        float d = length(p - motePos);
        float moteSize = MOTE_SIZE * (0.5 + depthLayer * 1.5);

        // Twinkle per mote (pulse in/out of visibility)
        float twinkle = 0.5 + 0.5 * sin(x_Time * (1.0 + depthLayer * 2.0) + fi * 0.7);
        twinkle = pow(twinkle, 4.0);

        // Chromatic aberration — offset per channel
        float chrom = moteSize * 0.4;
        float dR = length(p - motePos - vec2(chrom, 0.0));
        float dB = length(p - motePos + vec2(chrom, 0.0));

        vec3 motCol = sd_pal(fract(fi * 0.011 + x_Time * 0.04));
        float kR = exp(-dR * dR / (moteSize * moteSize) * 3.0);
        float kG = exp(-d  * d  / (moteSize * moteSize) * 3.0);
        float kB = exp(-dB * dB / (moteSize * moteSize) * 3.0);

        col.r += motCol.r * kR * twinkle * depthLayer;
        col.g += motCol.g * kG * twinkle * depthLayer;
        col.b += motCol.b * kB * twinkle * depthLayer;

        // Soft halo
        col += motCol * exp(-d * d * 8000.0) * twinkle * depthLayer * 0.1;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.8, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
